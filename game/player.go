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

func (p *Player) getReward() float32 {
	// Calculate euclidean distance between player 1 and player 2
	p1Position := p.view.position
	p2Position := p.game.player2Controller.player.view.position

	eucDistance := math.Sqrt(math.Pow(p1Position.X-p2Position.X, 2) + math.Pow(p1Position.Y-p2Position.Y, 2))

	eucDistanceBonus := float32(100 - eucDistance)

	previousEucDistance := p.game.previousEucDistance

	// If euclidean distance is less than 1, then the players are touching
	if eucDistance < 1 {
		return 2
	}

	// Add a bonus if player2 is in player1's render field
	bonus := float32(1)
	if p.view.isOtherPlayerSpriteVisible {
		bonus = 2
	}

	gameTickBonus := (p.game.currentTick - p.game.timeBonusStartTick) / 1000

	if eucDistance < previousEucDistance {
		p.game.previousEucDistance = eucDistance
		return (float32(1 + gameTickBonus)) * bonus * eucDistanceBonus
	} else {
		p.game.previousEucDistance = eucDistance
		p.game.timeBonusStartTick = p.game.currentTick

		return -1
	}

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
