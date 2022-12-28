package game

import "github.com/faiface/pixel/pixelgl"

type GameObject interface {
	update(delta float64)
	processInput(window *pixelgl.Window, timeDelta float64)
	render(delta float64)
}
