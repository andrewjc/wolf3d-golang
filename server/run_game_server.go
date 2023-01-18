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
	port       = 0 // random
)

func main() {
	flag.BoolVar(&fullscreen, "f", fullscreen, "fullscreen")
	flag.IntVar(&width, "w", width, "width")
	flag.IntVar(&height, "h", height, "height")
	flag.Float64Var(&scale, "s", scale, "scale")
	flag.IntVar(&port, "p", port, "port")
	flag.Parse()

	g := newGame(width, height, scale, fullscreen)

	// Set up the ipc servers used for controlling each player
	ipcServer := &ipc.IpcServer{
		Game: g,
		Config: &ipc.ServerConfig{
			IpcName: "wolf3d_ipc_player",
			Port:    port,
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
	} else if m.MsgType == 13 && string(m.Data) == "reset" {
		sc.Game.Reset()
		sc.Connection.Write(14, []byte("reset ok"))
	} else if m.MsgType == 16 && string(m.Data) == "begin control" {
		sc.Connection.Write(17, []byte("control granted"))
	} else if m.MsgType == 18 && string(m.Data) == "get observation" {

		p1, p2 := sc.Game.GetPlayer1Observation()
		result := game.RLActionResult{Reward: 0.0, Done: false, Info: "dummy", Observation: p2, Observation_Pos: p1}
		resultJson := result.ToJson()

		err := sc.Connection.Write(19, []byte(*resultJson))

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
