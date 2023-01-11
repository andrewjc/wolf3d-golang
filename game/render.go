package game

import "C"
import (
	"github.com/faiface/pixel"
	"image"
	"image/color"
	"math"
)

const texSize = 64

type RenderView struct {
	parent         *Player
	renderListener *RenderListener

	renderWidth  int
	renderHeight int

	direction pixel.Vec
	position  pixel.Vec
	plane     pixel.Vec

	velocity           float32
	horizontalVelocity float32

	distanceToWall float64 // Calculated after a render cycle
}

type RenderListener struct {
	renderBuffer *image.RGBA
}

func (c *RenderView) render() *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, c.renderWidth, c.renderHeight))

	c.renderWalls(m)

	c.renderThings(m)

	// Add code here to render the enemy position as a floating sphere
	// based on the enemy's position and the player's position

	if c.renderListener != nil {
		c.renderListener.renderBuffer = m
	}
	return m
}

func (c *RenderView) renderWalls(m *image.RGBA) {
	for x := 0; x < c.renderWidth; x++ {
		var (
			step         image.Point
			sideDist     pixel.Vec
			perpWallDist float64
			hit, side    bool

			worldX, worldY = int(c.position.X), int(c.position.Y)

			cameraX = 2*float64(x)/float64(c.renderWidth) - 1

			rayDir = pixel.V(
				c.direction.X+c.plane.X*cameraX,
				c.direction.Y+c.plane.Y*cameraX,
			)

			deltaDist = pixel.V(
				math.Sqrt(1.0+(rayDir.Y*rayDir.Y)/(rayDir.X*rayDir.X)),
				math.Sqrt(1.0+(rayDir.X*rayDir.X)/(rayDir.Y*rayDir.Y)),
			)
		)

		if rayDir.X < 0 {
			step.X = -1
			sideDist.X = (c.position.X - float64(int(c.position.X))) * deltaDist.X
		} else {
			step.X = 1
			sideDist.X = (float64(int(c.position.X)) + 1.0 - c.position.X) * deltaDist.X
		}

		if rayDir.Y < 0 {
			step.Y = -1
			sideDist.Y = (c.position.Y - float64(int(c.position.Y))) * deltaDist.Y
		} else {
			step.Y = 1
			sideDist.Y = (float64(int(c.position.Y)) + 1.0 - c.position.Y) * deltaDist.Y
		}

		for !hit {
			if sideDist.X < sideDist.Y {
				sideDist.X += deltaDist.X
				worldX += step.X
				side = false
			} else {
				sideDist.Y += deltaDist.Y
				worldY += step.Y
				side = true
			}

			if c.parent.game.mapData[worldX][worldY] > 0 {
				hit = true
			}
			if c.parent.game.mapData[worldX][worldY] == 3 {
				hit = false
			}
		}

		var wallX float64

		if side {
			perpWallDist = (float64(worldY) - c.position.Y + (1-float64(step.Y))/2) / rayDir.Y
			wallX = c.position.X + perpWallDist*rayDir.X
		} else {
			perpWallDist = (float64(worldX) - c.position.X + (1-float64(step.X))/2) / rayDir.X
			wallX = c.position.Y + perpWallDist*rayDir.Y
		}

		if x == c.renderWidth/2 {
			c.distanceToWall = perpWallDist
		}

		wallX -= math.Floor(wallX)

		texX := int(wallX * float64(texSize))

		lineHeight := int(float64(c.renderHeight) / perpWallDist)

		if lineHeight < 1 {
			lineHeight = 1
		}

		drawStart := -lineHeight/2 + c.renderHeight/2
		if drawStart < 0 {
			drawStart = 0
		}

		drawEnd := lineHeight/2 + c.renderHeight/2
		if drawEnd >= c.renderHeight {
			drawEnd = c.renderHeight - 1
		}

		if !side && rayDir.X > 0 {
			texX = texSize - texX - 1
		}

		if side && rayDir.Y < 0 {
			texX = texSize - texX - 1
		}

		texNum := c.parent.game.getTexNum(worldX, worldY)

		for y := drawStart; y < drawEnd+1; y++ {
			d := y*256 - c.renderHeight*128 + lineHeight*128
			texY := ((d * texSize) / lineHeight) / 256

			c := c.parent.game.textureMap.RGBAAt(
				texX+texSize*(texNum),
				texY%texSize,
			)

			if side {
				c.R = c.R / 2
				c.G = c.G / 2
				c.B = c.B / 2
			}

			m.Set(x, y, c)
		}

		var floorWall pixel.Vec

		if !side && rayDir.X > 0 {
			floorWall.X = float64(worldX)
			floorWall.Y = float64(worldY) + wallX
		} else if !side && rayDir.X < 0 {
			floorWall.X = float64(worldX) + 1.0
			floorWall.Y = float64(worldY) + wallX
		} else if side && rayDir.Y > 0 {
			floorWall.X = float64(worldX) + wallX
			floorWall.Y = float64(worldY)
		} else {
			floorWall.X = float64(worldX) + wallX
			floorWall.Y = float64(worldY) + 1.0
		}

		distWall, distPlayer := perpWallDist, 0.0

		for y := drawEnd + 1; y < c.renderHeight; y++ {
			currentDist := float64(c.renderHeight) / (2.0*float64(y) - float64(c.renderHeight))

			weight := (currentDist - distPlayer) / (distWall - distPlayer)

			currentFloor := pixel.V(
				weight*floorWall.X+(1.0-weight)*c.position.X,
				weight*floorWall.Y+(1.0-weight)*c.position.Y,
			)

			fx := int(currentFloor.X*float64(texSize)) % texSize
			fy := int(currentFloor.Y*float64(texSize)) % texSize

			m.Set(x, y, c.parent.game.textureMap.At(fx, fy))

			m.Set(x, c.renderHeight-y-1, c.parent.game.textureMap.At(fx+(4*texSize), fy))
			m.Set(x, c.renderHeight-y, c.parent.game.textureMap.At(fx+(4*texSize), fy))
		}
	}
}

func (c *RenderView) renderThings(m *image.RGBA) {
	for _, t := range c.parent.game.gameObjects {
		//if t == c.parent.game.player {
		//    continue
		// }

		x := t.getPosition().X - c.position.X
		y := t.getPosition().Y - c.position.Y

		invDet := 1.0 / (c.plane.X*c.direction.Y - c.direction.X*c.plane.Y)

		transformX := invDet * (c.direction.Y*x - c.direction.X*y)
		transformY := invDet * (-c.plane.Y*x + c.plane.X*y)

		spriteScreenX := int((float64(c.renderWidth) / 2) * (1 + transformX/transformY))

		spriteHeight := int(math.Abs(float64(c.renderHeight) / transformY))
		drawStartY := -spriteHeight/2 + c.renderHeight/2
		if drawStartY < 0 {
			drawStartY = 0
		}
		drawEndY := spriteHeight/2 + c.renderHeight/2
		if drawEndY >= c.renderHeight {
			drawEndY = c.renderHeight - 1
		}

		spriteWidth := int(math.Abs(float64(c.renderHeight) / transformY))
		drawStartX := -spriteWidth/2 + spriteScreenX
		if drawStartX < 0 {
			drawStartX = 0
		}
		drawEndX := spriteWidth/2 + spriteScreenX
		if drawEndX >= c.renderWidth {
			drawEndX = c.renderWidth - 1
		}

		for stripe := drawStartX; stripe < drawEndX; stripe++ {
			//texX := int(256*(stripe-(drawStartX-spriteWidth/2))*texSize/spriteWidth) / 256
			if transformY > 0 && stripe > 0 && stripe < c.renderWidth { // && transformY < c.zBuffer[stripe] {
				for y := drawStartY; y < drawEndY; y++ {
					//d := y*256 - c.renderHeight*128 + spriteHeight*128
					//texY := ((d * texSize) / spriteHeight) / 256
					//c := c.parent.game.textureMap.RGBAAt(texX+texSize*4, texY%texSize)
					c := color.RGBA{129, 127, 124, 0}
					if c.R != 0 {
						m.Set(stripe, y, c)
					}
				}
			}
		}
	}
}
