package ipc

/*
	Source: https://github.com/james-barrow/golang-ipc
*/

import (
	"bufio"
	"errors"
	"gameenv_ai/game"
	"log"
	"time"
)

type IpcServer struct {
	Game       *game.GameInstance
	Connection *IpcConnection
	Config     *ServerConfig
}

func (i *IpcServer) Start() (*IpcServer, error) {
	err := checkIpcName(i.Config.IpcName)
	if err != nil {
		return nil, err
	}

	sc := &IpcConnection{
		name:     i.Config.IpcName,
		status:   NotConnected,
		recieved: make(chan *Message),
		toWrite:  make(chan *Message),
	}

	if i.Config == nil {
		sc.timeout = 0
		sc.maxMsgSize = maxMsgSize
		sc.encryption = true
		sc.unMask = false

	} else {

		if i.Config.Timeout < 0 {
			sc.timeout = 0
		} else {
			sc.timeout = i.Config.Timeout
		}

		if i.Config.MaxMsgSize < 1024 {
			sc.maxMsgSize = maxMsgSize
		} else {
			sc.maxMsgSize = i.Config.MaxMsgSize
		}

		if i.Config.Encryption == false {
			sc.encryption = false
		} else {
			sc.encryption = true
		}

		if i.Config.UnmaskPermissions == true {
			sc.unMask = true
		} else {
			sc.unMask = false
		}

		sc.port = i.Config.Port
	}

	go func() {
		err := sc.beginListening()
		if err != nil {
			sc.recieved <- &Message{err: err, MsgType: -2}
		}
	}()

	i.Connection = sc

	return i, err
}

func (sc *IpcConnection) acceptLoop() {
	for {
		conn, err := sc.listen.Accept()
		if err != nil {
			break
		}

		if sc.status == Listening || sc.status == ReConnecting {

			sc.conn = conn

			err2 := sc.handshake()
			if err2 != nil {
				sc.recieved <- &Message{err: err2, MsgType: -2}

				// Reset the socket status to listening
				sc.status = Listening
				//sc.listen.Close()
				sc.conn.Close()

			} else {
				go sc.read()
				go sc.write()

				sc.status = Connected
				sc.recieved <- &Message{Status: sc.status.String(), MsgType: -1}
				sc.connChannel <- true
				log.Println("Client connection established!")
			}

		}

	}

}

func (sc *IpcConnection) connectionTimer() error {

	if sc.timeout != 0 {

		timeout := make(chan bool)

		go func() {
			time.Sleep(sc.timeout * time.Second)
			timeout <- true
		}()

		select {

		case <-sc.connChannel:
			return nil
		case <-timeout:
			sc.listen.Close()
			return errors.New("Timed out waiting for client to connect")
		}
	}

	select {

	case <-sc.connChannel:
		return nil
	}

}

func (sc *IpcConnection) read() {

	bLen := make([]byte, 4)

	for {

		res := sc.readData(bLen)
		if res == false {
			break
		}

		mLen := bytesToInt(bLen)

		msgRecvd := make([]byte, mLen)

		res = sc.readData(msgRecvd)
		if res == false {
			break
		}

		if sc.encryption == true {
			msgFinal, err := decrypt(*sc.enc.cipher, msgRecvd)
			if err != nil {
				sc.recieved <- &Message{err: err, MsgType: -2}
				continue
			}

			if bytesToInt(msgFinal[:4]) == 0 {
				//  type 0 = control message
			} else {
				sc.recieved <- &Message{Data: msgFinal[4:], MsgType: bytesToInt(msgFinal[:4])}
			}

		} else {
			if bytesToInt(msgRecvd[:4]) == 0 {
				//  type 0 = control message
			} else {
				sc.recieved <- &Message{Data: msgRecvd[4:], MsgType: bytesToInt(msgRecvd[:4])}
			}
		}

	}
}

func (sc *IpcConnection) readData(buff []byte) bool {

	_, err := sc.conn.Read(buff)
	if err != nil {

		if sc.status == Closing {

			sc.status = Closed
			sc.recieved <- &Message{Status: sc.status.String(), MsgType: -1}
			sc.recieved <- &Message{err: errors.New("IpcConnection has closed the connection"), MsgType: -2}
			return false
		}

		go sc.reConnect()
		return false

	}

	return true

}

func (sc *IpcConnection) reConnect() {

	sc.status = ReConnecting
	sc.recieved <- &Message{Status: sc.status.String(), MsgType: -1}

	err := sc.connectionTimer()
	if err != nil {
		sc.status = Timeout
		sc.recieved <- &Message{Status: sc.status.String(), MsgType: -1}

		sc.recieved <- &Message{err: err, MsgType: -2}

	}

}

// Read - blocking function that waits until an non multipart message is recieved

func (sc *IpcConnection) Read() (*Message, error) {

	m, ok := (<-sc.recieved)
	if ok == false {
		return nil, errors.New("the recieve channel has been closed")
	}

	if m.err != nil {
		close(sc.recieved)
		close(sc.toWrite)
		return nil, m.err
	}

	return m, nil

}

// Write - writes a non multipart message to the ipc connection.
// msgType - denotes the type of data being sent. 0 is a reserved type for internal messages and errors.
func (sc *IpcConnection) Write(msgType int, message []byte) error {

	if msgType == 0 {
		return errors.New("Message type 0 is reserved")
	}

	mlen := len(message)

	if mlen > sc.maxMsgSize {
		return errors.New("Message exceeds maximum message length")
	}

	if sc.status == Connected {

		sc.toWrite <- &Message{MsgType: msgType, Data: message}

	} else {
		return errors.New(sc.status.String())
	}

	return nil

}

func (sc *IpcConnection) write() {

	for {

		m, ok := <-sc.toWrite

		if ok == false {
			break
		}

		toSend := intToBytes(m.MsgType)

		writer := bufio.NewWriter(sc.conn)

		if sc.encryption == true {
			toSend = append(toSend, m.Data...)
			toSendEnc, err := encrypt(*sc.enc.cipher, toSend)
			if err != nil {
				//return err
			}
			toSend = toSendEnc
		} else {

			toSend = append(toSend, m.Data...)

		}

		writer.Write(intToBytes(len(toSend)))
		writer.Write(toSend)

		err := writer.Flush()
		if err != nil {
			//return err
		}

		time.Sleep(2 * time.Millisecond)

	}

}

// getStatus - get the current status of the connection
func (sc *IpcConnection) getStatus() Status {

	return sc.status

}

// StatusCode - returns the current connection status
func (sc *IpcConnection) StatusCode() Status {
	return sc.status
}

// Status - returns the current connection status as a string
func (sc *IpcConnection) Status() string {

	return sc.status.String()

}

// Close - closes the connection
func (sc *IpcConnection) Close() {

	sc.status = Closing
	sc.listen.Close()
	sc.conn.Close()

}
