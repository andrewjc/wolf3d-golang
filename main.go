package main

import (
	"flag"
	"gameenv_ai/game"
	"gameenv_ai/ipc"
	"github.com/faiface/pixel/pixelgl"
	"log"
)

var (
	fullscreen = false
	width      = 640
	height     = 480
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
			IpcName:           "wolf3d_ipc_player",
			Timeout:           0,
			MaxMsgSize:        1024,
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

func playerMessageLoop(sc *ipc.Server) {
	for {
		m, err := sc.Read()

		if err == nil {
			if m.MsgType > 0 {
				log.Println("Server recieved: "+string(m.Data)+" - Message type: ", m.MsgType)
			}

			handleServerPlayerMessage(sc, m)

		} else {
			log.Println("Server error")
			log.Println(err)
			break
		}
	}
}

func handleServerPlayerMessage(sc *ipc.Server, m *ipc.Message) {
	if m.MsgType == 666 && string(m.Data) == "ping" {
		sc.Write(667, []byte("pong"))
	}
	if m.MsgType == 100 && string(m.Data) == "begin control" {
		sc.Write(101, []byte("control granted"))
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
