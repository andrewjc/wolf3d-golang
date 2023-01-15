package game

import (
	"github.com/faiface/pixel"
)

type Enemy struct {
	view         *RenderView
	game         *GameInstance
	controller   *EnemyController
	isPlayerDead bool
}

func (p *Enemy) update(delta float64) {
	p.controller.update(delta)
}

func (p *Enemy) getPosition() pixel.Vec {
	return p.view.position
}

func (p *Enemy) getRotation() pixel.Vec {
	return p.view.direction
}

func (p *Enemy) getPlane() pixel.Vec {
	return p.view.plane
}
