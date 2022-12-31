package main

import (
	"flag"
	"gameenv_ai/game"
	"gameenv_ai/ipc"
	"github.com/faiface/pixel/pixelgl"
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
			IpcName:           "wolf3d_ipc",
			Timeout:           0,
			MaxMsgSize:        1024,
			Encryption:        true,
			UnmaskPermissions: false,
		},
	}
	ipcServer.Start()

	pixelgl.Run(g.GameLoop)
}

func newGame(width int, height int, scale float64, fullscreen bool) *game.GameInstance {
	return &game.GameInstance{
		RenderWidth:      width,
		RenderHeight:     height,
		RenderScale:      scale,
		RenderFullscreen: fullscreen,
	}
}

func NewIpcServer(g *game.GameInstance) *ipc.IpcServer {
	return &ipc.IpcServer{
		Game: g,
	}
}
