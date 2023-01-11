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
	// Check if player is done eg player1's position is too close to player2
	// Or if the elapsed time is greater than the max time
	return false
}

func getIntensityValues(m [][]int, playerPos pixel.Vec) [][]float64 {
	intensity := [][]float64{}
	for i := 0; i < len(m); i++ {
		intensity = append(intensity, make([]float64, len(m[0])))
	}

	// Calculate intensity of sound at each grid cell based on distance from source and obstacles
	for i := 0; i < len(m); i++ {
		for j := 0; j < len(m[0]); j++ {
			intensity[i][j] = calcIntensity(i, j, playerPos.X, playerPos.Y, m)
		}
	}

	return intensity
}

func calcIntensity(x, y int, sourceX, sourceY float64, m [][]int) float64 {
	// Calculate intensity of sound at (x, y) based on distance from source and obstacles
	// For example, using the inverse square law:
	distance := math.Sqrt(math.Pow(float64(x-int(sourceX)), 2) + math.Pow(float64(y-int(sourceY)), 2))
	intensity := 1.0 / math.Pow(distance, 2)

	// Adjust intensity based on obstacles in the way
	for i := x - 1; i <= x+1; i++ {
		for j := y - 1; j <= y+1; j++ {
			if i >= 0 && i < len(m) && j >= 0 && j < len(m[0]) && m[i][j] == 2 {
				intensity *= 0.5
			}
		}
	}

	return intensity
}
