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
    if win.Pressed(pixelgl.KeyUp) || win.Pressed(pixelgl.KeyW) {
        p.moveForward(3.5 * timeDelta)
    }

    if win.Pressed(pixelgl.KeyA) {
        p.moveLeft(3.5 * timeDelta)
    }

    if win.Pressed(pixelgl.KeyDown) || win.Pressed(pixelgl.KeyS) {
        p.moveBackwards(3.5 * timeDelta)
    }

    if win.Pressed(pixelgl.KeyD) {
        p.moveRight(3.5 * timeDelta)
    }

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
        if mapData[int(p.view.position.X-p.view.direction.X*s)][int(p.view.position.Y)] == 0 {
            p.view.position.X -= p.view.direction.X * s
        }

        if mapData[int(p.view.position.X)][int(p.view.position.Y-p.view.direction.Y*s)] == 0 {
            p.view.position.Y -= p.view.direction.Y * s
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
