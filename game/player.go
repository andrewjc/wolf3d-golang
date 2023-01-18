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

func (p *Player) getReward() float64 {
	// Calculate euclidean distance between player 1 and player 2
	p1Position := p.view.position
	p2Position := p.game.player2Controller.player.view.position

	eucDistance := math.Sqrt(math.Pow(p1Position.X-p2Position.X, 2) + math.Pow(p1Position.Y-p2Position.Y, 2))

	// Calculate reward that is inversely proportional to the euclidian distance
	reward := 1 / eucDistance
	return reward
}

func (p *Player) getIntensityValuesAroundPlayer() [][]float64 {
	intensity := getIntensityValues(p.game.mapData, p.getPosition())

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
