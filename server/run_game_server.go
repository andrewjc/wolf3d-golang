package main

import (
	"flag"
	"fmt"
	"gameenv_ai/game"
	"gameenv_ai/ipc"
	"github.com/faiface/pixel/pixelgl"
	"log"
)

var (
	fullscreen = false
	width      = 640
	height     = 480
	scale      = 1.0
)

func main() {
	flag.BoolVar(&fullscreen, "f", fullscreen, "fullscreen")
	flag.IntVar(&width, "w", width, "width")
	flag.IntVar(&height, "h", height, "height")
	flag.Float64Var(&scale, "s", scale, "scale")
	flag.Parse()

	g := newGame(width, height, scale, fullscreen)

	// Set up the ipc servers used for controlling each player
	ipcServer := &ipc.IpcServer{
		Game: g,
		Config: &ipc.ServerConfig{
			IpcName:           "wolf3d_ipc_player",
			Timeout:           0,
			MaxMsgSize:        1024 * 1024,
			Encryption:        false,
			UnmaskPermissions: false,
		},
	}

	sc, err := ipcServer.Start()
	if err != nil {
		return
	}

	go playerMessageLoop(sc)

	pixelgl.Run(g.GameLoop)
}

func playerMessageLoop(sc *ipc.IpcServer) {
	for {
		m, err := sc.Connection.Read()

		if err == nil {
			if m.MsgType > 0 {
				log.Println("IpcConnection recieved: "+string(m.Data)+" - Message type: ", m.MsgType)
			}

			handleServerPlayerMessage(sc, m)

		} else {
			log.Println("IpcConnection error")
			log.Println(err)
			break
		}
	}
}

func handleServerPlayerMessage(sc *ipc.IpcServer, m *ipc.Message) {
	if m.MsgType == 666 && string(m.Data) == "ping" {
		sc.Connection.Write(667, []byte("pong"))
	} else if m.MsgType == 100 && string(m.Data) == "begin control" {
		sc.Connection.Write(101, []byte("control granted"))
	} else if m.MsgType == 102 && string(m.Data) == "get observation" {
		err := sc.Connection.Write(103, sc.Game.GetPlayer1Observation())
		if err != nil {
			fmt.Println("Error writing observation: ", err)
		}
	} else if m.MsgType == 200 {
		sc.Game.TakePlayer1Action(game.RLAction(m.Data[0]))
		resultMessage := "action result: OK"

		writeError := sc.Connection.Write(201, []byte(resultMessage))
		if writeError != nil {
			fmt.Println("Error writing result message: ", writeError)
		}
		return
	} else if m.MsgType == -1 {
		// Control messages
		return
	} else {
		log.Fatal("Unknown message type: ", m.MsgType)
	}
}

func newGame(width int, height int, scale float64, fullscreen bool) *game.GameInstance {
	return &game.GameInstance{
		RenderWidth:      width,
		RenderHeight:     height,
		RenderScale:      scale,
		RenderFullscreen: fullscreen,
	}
}
