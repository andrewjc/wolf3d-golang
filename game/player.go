package game

import (
	"github.com/faiface/pixel"
	"math"
)

type Player struct {
	view         *RenderView
	game         *GameInstance
	controller   *PlayerController
	isPlayerDead bool
	old_position pixel.Vec
	is_moving    bool
}

func (p *Player) update(delta float64) {
	// Check if player is dead eg player1's position is too close to player2

}

func (p *Player) getPosition() pixel.Vec {
	return p.view.position
}

func (p *Player) getRotation() pixel.Vec {
	return p.view.direction
}

func (p *Player) getPlane() pixel.Vec {
	return p.view.plane
}

func (p *Player) getReward() float32 {
	// Calculate euclidean distance between player 1 and player 2
	p1Position := p.view.position
	p2Position := p.game.player2Controller.player.view.position

	eucDistance := math.Sqrt(math.Pow(p1Position.X-p2Position.X, 2) + math.Pow(p1Position.Y-p2Position.Y, 2))

	// Push log onto the stack
	p.game.player1Controller.distanceStack = append(p.game.player1Controller.distanceStack, 100-eucDistance)

	// If the stack is full, pop the oldest value off the stack
	if len(p.game.player1Controller.distanceStack) > 10 {
		p.game.player1Controller.distanceStack = p.game.player1Controller.distanceStack[1:]
	}

	// Calculate the average of the stack
	sum := 0.0
	for _, value := range p.game.player1Controller.distanceStack {
		sum += value
	}

	average := sum / float64(len(p.game.player1Controller.distanceStack))

	bonuses := 1.0

	if p.is_moving {
		bonuses = 2
	}

	if p.view.isOtherPlayerSpriteVisible {
		bonuses = 3
	}

	// If the log of the average is less than 0, then the players are getting closer
	return float32(math.Log(average * bonuses))
}

func (p *Player) getIntensityValuesAroundPlayer() [][]float64 {
	enemy := p.game.player2Controller.player
	intensity := getIntensityValues(p.game.mapData, enemy.getPosition())

	// Filter intensity values to only include values in a 1 grid cell radius around player's position
	filteredIntensity := [][]float64{}
	playerPos := p.getPosition()
	mapData := p.game.mapData
	for i := playerPos.X - 1; i <= playerPos.X+1; i++ {
		row := []float64{}
		for j := playerPos.Y - 1; j <= playerPos.Y+1; j++ {
			if i >= 0 && int(i) < len(mapData) && j >= 0 && int(j) < len(mapData[0]) {
				row = append(row, intensity[int(i)][int(j)])
			}
		}
		filteredIntensity = append(filteredIntensity, row)
	}

	return filteredIntensity
}

func (p *Player) isDead() bool {
	return p.isPlayerDead
}

func (p *Player) isDone() bool {
	// Calculate euclidean distance between player 1 and player 2
	p1Position := p.view.position
	p2Position := p.game.player2Controller.player.view.position

	eucDistance := math.Sqrt(math.Pow(p1Position.X-p2Position.X, 2) + math.Pow(p1Position.Y-p2Position.Y, 2))

	// If euclidean distance is less than 1, then the players are touching
	if eucDistance < 1 {
		return true
	}

	return false
}
