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
	width      = 320
	height     = 240
	scale      = 3.0
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
			IpcName: "wolf3d_ipc_player",
			Timeout: 0,
			//MaxMsgSize:        1024 * 1024 * 10,
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
	if m.MsgType == 11 && string(m.Data) == "ping" {
		sc.Connection.Write(12, []byte("pong"))
	} else if m.MsgType == 16 && string(m.Data) == "begin control" {
		sc.Connection.Write(17, []byte("control granted"))
	} else if m.MsgType == 18 && string(m.Data) == "get observation" {
		err := sc.Connection.Write(19, sc.Game.GetPlayer1Observation())
		if err != nil {
			fmt.Println("Error writing observation: ", err)
		}
	} else if m.MsgType == 20 {
		result := sc.Game.TakePlayer1Action(game.RLAction(m.Data[0]))
		resultJson := result.ToJson()

		if resultJson != nil {
			writeError := sc.Connection.Write(21, []byte(*resultJson))
			if writeError != nil {
				fmt.Println("Error writing result message: ", writeError)
			}
		} else {
			fmt.Println("Error converting action result to json")
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
