package game

import (
	"github.com/faiface/pixel/pixelgl"
	"math"
)

type Player struct {
	view       *RenderView
	game       *GameInstance
	controller *PlayerController
}

func (p *Player) update(timeDelta float64) {
	p.view.render()
}

func (p *Player) render(timeDelta float64) {
	p.view.render()
}

func (p *Player) processInput(win *pixelgl.Window, timeDelta float64) {
	p.processForwardBackAcceleration(win)

	p.processLeftRightAcceleration(win)

	if win.Pressed(pixelgl.KeyRight) {
		p.turnRight(1.2 * timeDelta)
	}

	if win.Pressed(pixelgl.KeyLeft) {
		p.turnLeft(1.2 * timeDelta)
	}

	mouseVector := win.MousePosition().Sub(win.MousePreviousPosition())
	if mouseVector.X > 0 {
		p.turnRight(mouseVector.X * 0.01)
	} else {
		p.turnLeft(mouseVector.X * -0.01)
	}
}

func (p *Player) processForwardBackAcceleration(win *pixelgl.Window) {
	accelTriggered := false
	if win.Pressed(pixelgl.KeyUp) || win.Pressed(pixelgl.KeyW) {
		p.accelerateForward()
		accelTriggered = true
	}

	if win.Pressed(pixelgl.KeyDown) || win.Pressed(pixelgl.KeyS) {
		p.accelerateBackward()
		accelTriggered = true
	}

	if !accelTriggered {
		if p.view.velocity > 0 {
			p.view.velocity += backward_acceleration
			if p.view.velocity < 0 {
				p.view.velocity = 0
			}
		} else if p.view.velocity < 0 {
			p.view.velocity -= backward_acceleration
			if p.view.velocity > 0 {
				p.view.velocity = 0
			}
		}
	}
}

func (p *Player) processLeftRightAcceleration(win *pixelgl.Window) {
	accelTriggered := false
	if win.Pressed(pixelgl.KeyA) {
		p.accelerateLeft()
		accelTriggered = true
	}

	if win.Pressed(pixelgl.KeyD) {
		p.accelerateRight()
		accelTriggered = true
	}

	if !accelTriggered {
		if p.view.horizontalVelocity > 0 {
			p.view.horizontalVelocity += backward_acceleration
			if p.view.horizontalVelocity < 0 {
				p.view.horizontalVelocity = 0
			}
		} else if p.view.horizontalVelocity < 0 {
			p.view.horizontalVelocity -= backward_acceleration
			if p.view.horizontalVelocity > 0 {
				p.view.horizontalVelocity = 0
			}
		}
	}
}

func (p *Player) moveForward(s float64) {
	mapData := p.game.mapData
	_, ok := interface{}(mapData).([][]int)

	if ok {

		if p.view.distanceToWall > 0.3 {
			if mapData[int(p.view.position.X+p.view.direction.X*s)][int(p.view.position.Y)] == 0 {
				p.view.position.X += p.view.direction.X * s
			}

			if mapData[int(p.view.position.X)][int(p.view.position.Y+p.view.direction.Y*s)] == 0 {
				p.view.position.Y += p.view.direction.Y * s
			}
		}
	}
}

const forward_acceleration = 0.003
const backward_acceleration = -0.003
const max_velocity = 0.1
const min_velocity = -0.1

func (p *Player) accelerateForward() {
	p.view.velocity += forward_acceleration

	if p.view.velocity > max_velocity {
		p.view.velocity = max_velocity
	}

	p.moveForward(float64(p.view.velocity))
}

func (p *Player) accelerateBackward() {
	p.view.velocity -= backward_acceleration

	if p.view.velocity < min_velocity {
		p.view.velocity = min_velocity
	}

	p.moveBackwards(float64(p.view.velocity))
}

func (p *Player) accelerateLeft() {
	p.view.horizontalVelocity += forward_acceleration

	if p.view.horizontalVelocity > max_velocity {
		p.view.horizontalVelocity = max_velocity
	}

	p.moveLeft(float64(p.view.horizontalVelocity))
}

func (p *Player) accelerateRight() {
	p.view.horizontalVelocity += forward_acceleration

	if p.view.horizontalVelocity > max_velocity {
		p.view.horizontalVelocity = max_velocity
	}

	p.moveRight(float64(p.view.horizontalVelocity))
}

func (p *Player) moveLeft(s float64) {
	mapData := p.game.mapData
	_, ok := interface{}(mapData).([][]int)

	if ok {
		if mapData[int(p.view.position.X-p.view.plane.X*s)][int(p.view.position.Y)] == 0 {
			p.view.position.X -= p.view.plane.X * s
		}

		if mapData[int(p.view.position.X)][int(p.view.position.Y-p.view.plane.Y*s)] == 0 {
			p.view.position.Y -= p.view.plane.Y * s
		}
	}
}

func (p *Player) moveBackwards(s float64) {
	mapData := p.game.mapData
	_, ok := interface{}(mapData).([][]int)

	if ok {
		// Check if new position is within map bounds
		newX := int(p.view.position.X - p.view.direction.X*s)
		if newX >= 0 && newX < len(mapData) {
			// Check if new position is empty space
			if mapData[newX][int(p.view.position.Y)] == 0 {
				p.view.position.X -= p.view.direction.X * s
			}
		}

		// Check if new position is within map bounds
		newY := int(p.view.position.Y - p.view.direction.Y*s)
		if newY >= 0 && newY < len(mapData[0]) {
			// Check if new position is empty space
			if mapData[int(p.view.position.X)][newY] == 0 {
				p.view.position.Y -= p.view.direction.Y * s
			}
		}
	}
}

func (p *Player) moveRight(s float64) {
	mapData := p.game.mapData
	_, ok := interface{}(mapData).([][]int)

	if ok {
		if mapData[int(p.view.position.X+p.view.plane.X*s)][int(p.view.position.Y)] == 0 {
			p.view.position.X += p.view.plane.X * s
		}

		if mapData[int(p.view.position.X)][int(p.view.position.Y+p.view.plane.Y*s)] == 0 {
			p.view.position.Y += p.view.plane.Y * s
		}
	}
}

func (p *Player) turnRight(s float64) {
	oldDirX := p.view.direction.X

	p.view.direction.X = p.view.direction.X*math.Cos(-s) - p.view.direction.Y*math.Sin(-s)
	p.view.direction.Y = oldDirX*math.Sin(-s) + p.view.direction.Y*math.Cos(-s)

	oldPlaneX := p.view.plane.X

	p.view.plane.X = p.view.plane.X*math.Cos(-s) - p.view.plane.Y*math.Sin(-s)
	p.view.plane.Y = oldPlaneX*math.Sin(-s) + p.view.plane.Y*math.Cos(-s)
}

func (p *Player) turnLeft(s float64) {
	oldDirX := p.view.direction.X

	p.view.direction.X = p.view.direction.X*math.Cos(s) - p.view.direction.Y*math.Sin(s)
	p.view.direction.Y = oldDirX*math.Sin(s) + p.view.direction.Y*math.Cos(s)

	oldPlaneX := p.view.plane.X

	p.view.plane.X = p.view.plane.X*math.Cos(s) - p.view.plane.Y*math.Sin(s)
	p.view.plane.Y = oldPlaneX*math.Sin(s) + p.view.plane.Y*math.Cos(s)
}

func (p *Player) getReward() float64 {
	// Calculate euclidian distance between player 1 and player 2
	p1Position := p.view.position
	p2Position := p.game.player2Controller.player.view.position

	eucDistance := math.Sqrt(math.Pow(p1Position.X-p2Position.X, 2) + math.Pow(p1Position.Y-p2Position.Y, 2))

	// Calculate reward that is inversely proportional to the euclidian distance
	reward := 1 / eucDistance
	return reward
}

func (p *Player) getIntensityValuesAroundPlayer(m *Map, enemyPos, playerPos []int, soundData [][]float64) [][]float64 {
	intensity := getIntensityValues(m, enemyPos, playerPos, soundData)

	// Filter intensity values to only include values in a 1 grid cell radius around player's position
	filteredIntensity := [][]float64{}
	for i := playerPos[0] - 1; i <= playerPos[0]+1; i++ {
		row := []float64{}
		for j := playerPos[1] - 1; j <= playerPos[1]+1; j++ {
			if i >= 0 && i < m.rows && j >= 0 && j < m.cols {
				row = append(row, intensity[i][j])
			}
		}
		filteredIntensity = append(filteredIntensity, row)
	}

	return filteredIntensity
}

func getIntensityValues(m *Map, enemyPos, playerPos []int, soundData [][]float64) [][]float64 {
	intensity := [][]float64{}
	for i := 0; i < m.rows; i++ {
		intensity = append(intensity, make([]float64, m.cols))
	}

	// Calculate intensity of sound at each grid cell based on distance from source and obstacles
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			intensity[i][j] = calcIntensity(i, j, playerPos[0], playerPos[1], m)
		}
	}

	return intensity
}

func calcIntensity(x, y, sourceX, sourceY int, m *Map) float64 {
	// Calculate intensity of sound at (x, y) based on distance from source and obstacles
	// For example, using the inverse square law:
	distance := math.Sqrt(math.Pow(float64(x-sourceX), 2) + math.Pow(float64(y-sourceY), 2))
	intensity := 1.0 / math.Pow(distance, 2)

	// Adjust intensity based on obstacles in the way
	for i := x - 1; i <= x+1; i++ {
		for j := y - 1; j <= y+1; j++ {
			if i >= 0 && i < m.rows && j >= 0 && j < m.cols && m.mapData[i][j] == 2 {
				intensity *= 0.5
			}
		}
	}

	return intensity
}
