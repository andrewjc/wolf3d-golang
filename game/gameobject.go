package game

import (
	"github.com/faiface/pixel"
)

type GameObject interface {
	update(delta float64)
	getPosition() pixel.Vec
	getRotation() pixel.Vec
}
