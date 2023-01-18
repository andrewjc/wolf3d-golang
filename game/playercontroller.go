package game

import (
    "github.com/faiface/pixel/pixelgl"
    "math"
)

type PlayerController struct {
    player *Player
}

const forward_acceleration = 0.003
const backward_acceleration = -0.003
const max_velocity = 0.1
const min_velocity = -0.1

func (p *PlayerController) processInput(win *pixelgl.Window, dt float64) {
    p.processForwardBackAcceleration(win)

    p.processLeftRightAcceleration(win)

    if win.Pressed(pixelgl.KeyRight) {
        p.turnRight(1.2 * dt)
    }

    if win.Pressed(pixelgl.KeyLeft) {
        p.turnLeft(1.2 * dt)
    }

    mouseVector := win.MousePosition().Sub(win.MousePreviousPosition())
    if mouseVector.X > 0 {
        p.turnRight(mouseVector.X * 0.01)
    } else {
        p.turnLeft(mouseVector.X * -0.01)
    }
}

func (p *PlayerController) processForwardBackAcceleration(win *pixelgl.Window) {
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
        if p.player.view.velocity > 0 {
            p.player.view.velocity += backward_acceleration
            if p.player.view.velocity < 0 {
                p.player.view.velocity = 0
            }
        } else if p.player.view.velocity < 0 {
            p.player.view.velocity -= backward_acceleration
            if p.player.view.velocity > 0 {
                p.player.view.velocity = 0
            }
        }
    }
}

func (p *PlayerController) processLeftRightAcceleration(win *pixelgl.Window) {
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
        if p.player.view.horizontalVelocity > 0 {
            p.player.view.horizontalVelocity += backward_acceleration
            if p.player.view.horizontalVelocity < 0 {
                p.player.view.horizontalVelocity = 0
            }
        } else if p.player.view.horizontalVelocity < 0 {
            p.player.view.horizontalVelocity -= backward_acceleration
            if p.player.view.horizontalVelocity > 0 {
                p.player.view.horizontalVelocity = 0
            }
        }
    }
}

func (p *PlayerController) moveForward(s float64) {
    mapData := p.player.game.mapData
    _, ok := interface{}(mapData).([][]int)

    if ok {

        if p.player.view.distanceToWall > 0.3 {
            if mapData[int(p.player.view.position.X+p.player.view.direction.X*s)][int(p.player.view.position.Y)] == 0 {
                p.player.view.position.X += p.player.view.direction.X * s
            }

            if mapData[int(p.player.view.position.X)][int(p.player.view.position.Y+p.player.view.direction.Y*s)] == 0 {
                p.player.view.position.Y += p.player.view.direction.Y * s
            }
        }
    }

}

func (p *PlayerController) accelerateForward() {
    p.player.view.velocity += forward_acceleration

    if p.player.view.velocity > max_velocity {
        p.player.view.velocity = max_velocity
    }

    p.moveForward(float64(p.player.view.velocity))
}

func (p *PlayerController) accelerateBackward() {
    p.player.view.velocity -= backward_acceleration

    if p.player.view.velocity < min_velocity {
        p.player.view.velocity = min_velocity
    }

    p.moveBackwards(float64(p.player.view.velocity))
}

func (p *PlayerController) accelerateLeft() {
    p.player.view.horizontalVelocity += forward_acceleration

    if p.player.view.horizontalVelocity > max_velocity {
        p.player.view.horizontalVelocity = max_velocity
    }

    p.moveLeft(float64(p.player.view.horizontalVelocity))
}

func (p *PlayerController) accelerateRight() {
    p.player.view.horizontalVelocity += forward_acceleration

    if p.player.view.horizontalVelocity > max_velocity {
        p.player.view.horizontalVelocity = max_velocity
    }

    p.moveRight(float64(p.player.view.horizontalVelocity))
}

func (p *PlayerController) moveLeft(s float64) {
    mapData := p.player.game.mapData
    _, ok := interface{}(mapData).([][]int)

    if ok {
        if mapData[int(p.player.view.position.X-p.player.view.plane.X*s)][int(p.player.view.position.Y)] == 0 {
            p.player.view.position.X -= p.player.view.plane.X * s
        }

        if mapData[int(p.player.view.position.X)][int(p.player.view.position.Y-p.player.view.plane.Y*s)] == 0 {
            p.player.view.position.Y -= p.player.view.plane.Y * s
        }
    }
}

func (p *PlayerController) moveBackwards(s float64) {
    mapData := p.player.game.mapData
    _, ok := interface{}(mapData).([][]int)

    if ok {
        // Check if new position is within map bounds
        newX := int(p.player.view.position.X - p.player.view.direction.X*s)
        if newX >= 0 && newX < len(mapData) {
            // Check if new position is empty space
            if mapData[newX][int(p.player.view.position.Y)] == 0 {
                p.player.view.position.X -= p.player.view.direction.X * s
            }
        }

        // Check if new position is within map bounds
        newY := int(p.player.view.position.Y - p.player.view.direction.Y*s)
        if newY >= 0 && newY < len(mapData[0]) {
            // Check if new position is empty space
            if mapData[int(p.player.view.position.X)][newY] == 0 {
                p.player.view.position.Y -= p.player.view.direction.Y * s
            }
        }
    }
}

func (p *PlayerController) moveRight(s float64) {
    mapData := p.player.game.mapData
    _, ok := interface{}(mapData).([][]int)

    if ok {
        if mapData[int(p.player.view.position.X+p.player.view.plane.X*s)][int(p.player.view.position.Y)] == 0 {
            p.player.view.position.X += p.player.view.plane.X * s
        }

        if mapData[int(p.player.view.position.X)][int(p.player.view.position.Y+p.player.view.plane.Y*s)] == 0 {
            p.player.view.position.Y += p.player.view.plane.Y * s
        }
    }
}

func (p *PlayerController) turnRight(s float64) {
    oldDirX := p.player.view.direction.X

    p.player.view.direction.X = p.player.view.direction.X*math.Cos(-s) - p.player.view.direction.Y*math.Sin(-s)
    p.player.view.direction.Y = oldDirX*math.Sin(-s) + p.player.view.direction.Y*math.Cos(-s)

    oldPlaneX := p.player.view.plane.X

    p.player.view.plane.X = p.player.view.plane.X*math.Cos(-s) - p.player.view.plane.Y*math.Sin(-s)
    p.player.view.plane.Y = oldPlaneX*math.Sin(-s) + p.player.view.plane.Y*math.Cos(-s)
}

func (p *PlayerController) turnLeft(s float64) {
    oldDirX := p.player.view.direction.X

    p.player.view.direction.X = p.player.view.direction.X*math.Cos(s) - p.player.view.direction.Y*math.Sin(s)
    p.player.view.direction.Y = oldDirX*math.Sin(s) + p.player.view.direction.Y*math.Cos(s)

    oldPlaneX := p.player.view.plane.X

    p.player.view.plane.X = p.player.view.plane.X*math.Cos(s) - p.player.view.plane.Y*math.Sin(s)
    p.player.view.plane.Y = oldPlaneX*math.Sin(s) + p.player.view.plane.Y*math.Cos(s)
}

func (p *PlayerController) deaccelerateVelocity() {
    if p.player.view.velocity > 0 {
        p.player.view.velocity += backward_acceleration
        if p.player.view.velocity < 0 {
            p.player.view.velocity = 0
        }
    } else if p.player.view.velocity < 0 {
        p.player.view.velocity -= backward_acceleration
        if p.player.view.velocity > 0 {
            p.player.view.velocity = 0
        }
    }
}

func (p *PlayerController) deaccelerateHorizontalVelocity() {
    if p.player.view.horizontalVelocity > 0 {
        p.player.view.horizontalVelocity += backward_acceleration
        if p.player.view.horizontalVelocity < 0 {
            p.player.view.horizontalVelocity = 0
        }
    } else if p.player.view.horizontalVelocity < 0 {
        p.player.view.horizontalVelocity -= backward_acceleration
        if p.player.view.horizontalVelocity > 0 {
            p.player.view.horizontalVelocity = 0
        }
    }
}
