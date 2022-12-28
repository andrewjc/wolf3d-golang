package main

import (
	"flag"
	"gameenv_ai/game"
	"github.com/faiface/pixel/pixelgl"
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
