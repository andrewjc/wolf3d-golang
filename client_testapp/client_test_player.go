package main

import (
    "gameenv_ai/ipc"
    "log"
    "time"
)

func main() {

    config := &ipc.ClientConfig{Encryption: false, Timeout: 0}

    cc, err := ipc.StartClient("wolf3d_ipc_player", config)
    if err != nil {
        log.Println(err)
        return
    }

    go clientPlayerMessageLoop(cc)

    clientPlayerWaitLoop(cc)
}

func clientPlayerMessageLoop(cc *ipc.Client) {
    for {
        m, err := cc.Read()

        if err != nil {
            // An error is only returned if the recieved channel has been closed,
            //so you know the connection has either been intentionally closed or has timmed out waiting to connect/re-connect.
            break
        }

        if m.MsgType == -1 { // message type -1 is status change
            log.Println("Status: " + m.Status)
            if m.Status == "Connected" {
                cc.Write(100, []byte("begin control"))
            }
        }

        if m.MsgType == -2 { // message type -2 is an error, these won't automatically cause the recieve channel to close.
            log.Println("Error: " + err.Error())
        }

        if m.MsgType > 0 { // all message types above 0 have been recieved over the connection

            handleClientPlayerMessage(cc, m)
        }
        //}
    }
}

func handleClientPlayerMessage(cc *ipc.Client, m *ipc.Message) {

    log.Println(" Message type: ", m.MsgType)
    log.Println("Client recieved: " + string(m.Data))

    if m.MsgType == 101 && string(m.Data) == "control granted" {
        cc.Write(102, []byte("move forward"))
    }

}

func clientPlayerWaitLoop(cc *ipc.Client) {

    for {
        _ = cc.Write(666, []byte("ping"))

        time.Sleep(time.Second * 15)

    }
}
