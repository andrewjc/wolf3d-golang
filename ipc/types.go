package ipc

import (
	"bytes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"net"
	"time"
)

type IpcConnection struct {
	name        string
	listen      net.Listener
	conn        net.Conn
	status      Status
	recieved    chan (*Message)
	connChannel chan bool
	toWrite     chan (*Message)
	timeout     time.Duration
	encryption  bool
	maxMsgSize  int
	enc         *encryption
	unMask      bool
	port        int
}

// Client - holds the details of the client connection and config.
type Client struct {
	Name          string
	conn          net.Conn
	status        Status
	timeout       float64       //
	retryTimer    time.Duration // number of seconds before trying to connect again
	recieved      chan (*Message)
	toWrite       chan (*Message)
	encryption    bool
	encryptionReq bool
	maxMsgSize    int
	enc           *encryption
}

// Message - contains the  recieved message
type Message struct {
	err     error  // details of any error
	MsgType int    // type of message sent - 0 is reserved
	Data    []byte // message data recieved
	Status  string
}

// Status - Status of the connection
type Status int

const (

	// NotConnected - 0
	NotConnected Status = iota
	// Listening - 1
	Listening Status = iota
	// Connecting - 2
	Connecting Status = iota
	// Connected - 3
	Connected Status = iota
	// ReConnecting - 4
	ReConnecting Status = iota
	// Closed - 5
	Closed Status = iota
	// Closing - 6
	Closing Status = iota
	// Error - 7
	Error Status = iota
	// Timeout - 8
	Timeout Status = iota
)

// ServerConfig - used to pass configuation overrides to ServerStart()
type ServerConfig struct {
	IpcName           string
	Timeout           time.Duration
	MaxMsgSize        int
	Encryption        bool
	UnmaskPermissions bool
	Port              int
}

// ClientConfig - used to pass configuation overrides to ClientStart()
type ClientConfig struct {
	Timeout    float64
	RetryTimer time.Duration
	Encryption bool
}

// Encryption - encryption settings
type encryption struct {
	keyExchange string
	encryption  string
	cipher      *cipher.AEAD
}

func (status *Status) String() string {

	switch *status {
	case NotConnected:
		return "Not Connected"
	case Connecting:
		return "Connecting"
	case Connected:
		return "Connected"
	case Listening:
		return "Listening"
	case Closing:
		return "Closing"
	case ReConnecting:
		return "Re-connecting"
	case Timeout:
		return "Timeout"
	case Closed:
		return "Closed"
	case Error:
		return "Error"
	default:
		return "Status not found"
	}
}

// checks the name passed into the start function to ensure it's ok/will work.
func checkIpcName(ipcName string) error {

	if len(ipcName) == 0 {
		return errors.New("ipcName cannot be an empty string")
	}

	return nil

}

func intToBytes(mLen int) []byte {

	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(mLen))

	return b

}

func bytesToInt(b []byte) int {

	var mlen uint32

	binary.Read(bytes.NewReader(b[:]), binary.BigEndian, &mlen) // message length

	return int(mlen)

}
