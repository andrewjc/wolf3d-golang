package game

import "C"
import (
	"fmt"
	"github.com/faiface/pixel"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"math"
)

const texSize = 64

var renderCeilingFloor = true

type RenderView struct {
	parent         interface{}
	renderListener *RenderListener

	renderWidth  int
	renderHeight int

	direction pixel.Vec
	position  pixel.Vec
	plane     pixel.Vec

	velocity           float32
	horizontalVelocity float32

	distanceToWall float64 // Calculated after a render cycle
	zBuffer        [][]float64
	old_position   pixel.Vec
	is_moving      bool
}

type RenderListener struct {
	renderBuffer *image.RGBA
}

func (c *RenderView) render() *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, c.renderWidth, c.renderHeight))

	// Initialize the zbuffer
	c.zBuffer = make([][]float64, c.renderWidth)
	for i := range c.zBuffer {
		c.zBuffer[i] = make([]float64, c.renderHeight)
	}

	c.renderWalls(m)

	c.renderThings(m)

	c.renderPosition(m)

	if c.renderListener != nil {
		c.renderListener.renderBuffer = m
	}
	return m
}

func (c *RenderView) renderWalls(m *image.RGBA) {

	for x := 0; x < c.renderWidth; x++ {
		var step image.Point

		worldX, worldY := int(c.position.X), int(c.position.Y)

		cameraX := 2*float64(x)/float64(c.renderWidth) - 1

		rayDir := pixel.V(
			c.direction.X+c.plane.X*cameraX,
			c.direction.Y+c.plane.Y*cameraX,
		)

		deltaDist := pixel.V(
			math.Sqrt(1.0+(rayDir.Y*rayDir.Y)/(rayDir.X*rayDir.X)),
			math.Sqrt(1.0+(rayDir.X*rayDir.X)/(rayDir.Y*rayDir.Y)),
		)

		var sideDist pixel.Vec
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

		var hit bool
		var side bool
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

			if c.parent.(*Player).game.mapData[worldX][worldY] > 0 {
				hit = true
			}
			if c.parent.(*Player).game.mapData[worldX][worldY] == 3 {
				hit = false
			}
		}

		var wallX float64
		var perpWallDist float64

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

		texNum := c.parent.(*Player).game.getTexNum(worldX, worldY)

		for y := drawStart; y < drawEnd+1; y++ {
			texY := (float64(y) - float64(c.renderHeight)/2 + float64(lineHeight)/2) * texSize / float64(lineHeight)

			col := c.parent.(*Player).game.textureMap.RGBAAt(
				texX+texSize*(texNum),
				int(texY)%texSize,
			)

			if side {
				col.R = col.R / 2
				col.G = col.G / 2
				col.B = col.B / 2
			}

			maxDistance := math.Max(float64(len(c.parent.(*Player).game.mapData)), float64(len(c.parent.(*Player).game.mapData[0])))
			percentage := perpWallDist / maxDistance
			// invert percentage
			percentage = 1.0 - percentage

			percentage = applyDistanceFalloff(percentage, perpWallDist)
			if percentage < 1e-6 {
				percentage = 1e-6
			}

			// scale the color by the percentage
			col.R = uint8(float64(col.R) * percentage)
			col.G = uint8(float64(col.G) * percentage)
			col.B = uint8(float64(col.B) * percentage)

			m.Set(x, y, col)

			// Calculate the zbuffer
			zBufferValue := perpWallDist
			// Store the zBuffer value in the zBuffer array
			c.zBuffer[x][y] = zBufferValue
		}

		if renderCeilingFloor {

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

				perpFloorDist := currentDist

				maxDistance := math.Max(float64(len(c.parent.(*Player).game.mapData)), float64(len(c.parent.(*Player).game.mapData[0])))
				percentage := perpFloorDist / maxDistance
				// invert percentage
				percentage = 1.0 - percentage

				percentage = applyDistanceFalloff(percentage, perpFloorDist)
				if percentage < 1e-6 {
					percentage = 1e-6
				}

				// scale the color by the percentage

				col := c.parent.(*Player).game.textureMap.RGBAAt(fx, fy)
				col.R = uint8(float64(col.R) * percentage)
				col.G = uint8(float64(col.G) * percentage)
				col.B = uint8(float64(col.B) * percentage)

				// Render floor
				m.Set(x, y, col)
				c.zBuffer[x][y] = perpFloorDist

				// Render roof
				col = c.parent.(*Player).game.textureMap.RGBAAt(fx+(4*texSize), fy)
				col.R = uint8(float64(col.R) * percentage)
				col.G = uint8(float64(col.G) * percentage)
				col.B = uint8(float64(col.B) * percentage)
				m.Set(x, c.renderHeight-y-1, col)
				m.Set(x, c.renderHeight-y, col)

				// Save this pixel to the z-buffer
				c.zBuffer[x][c.renderHeight-y-1] = perpFloorDist
				c.zBuffer[x][c.renderHeight-y] = perpFloorDist
			}

		}

	}
}

var (
	falloffPercentages = map[float64]float64{}
	lightest           = 0.80
	darkest            = 0.0
	darkestDistance    = 10.0
)

func init() {
	light := lightest
	for i := 0.0; i <= darkestDistance; i += 0.25 {
		falloffPercentages[i] = light
		if light > darkest {
			light -= 0.005
		}
	}
	darkest = light
}

func applyDistanceFalloff(percentage float64, dist float64) float64 {
	if dist > darkestDistance {
		return percentage * darkest
	}
	return percentage * falloffPercentages[roundDownToClosest(dist)]
}

// this will round down to the closest multiple of .25
func roundDownToClosest(f float64) float64 {
	rounded := math.Floor(f*4 + 0.25)
	if math.Floor(rounded) == rounded {
		return rounded / 4
	}
	return math.Floor(f)
}

func (r *RenderView) renderThings(m *image.RGBA) {
	for _, t := range r.parent.(*Player).game.gameObjects {

		x := t.getPosition().X - r.position.X
		y := t.getPosition().Y - r.position.Y

		invDet := 1.0 / (r.plane.X*r.direction.Y - r.direction.X*r.plane.Y)

		transformX := invDet * (r.direction.Y*x - r.direction.X*y)
		transformY := invDet * (-r.plane.Y*x + r.plane.X*y)

		objectPerpDist := transformY

		spriteScreenX := int((float64(r.renderWidth) / 2) * (1 + transformX/transformY))

		spriteHeight := int(math.Abs(float64(r.renderHeight) / transformY))
		drawStartY := -spriteHeight/2 + r.renderHeight/2
		if drawStartY < 0 {
			drawStartY = 0
		}
		drawEndY := spriteHeight/2 + r.renderHeight/2
		if drawEndY >= r.renderHeight {
			drawEndY = r.renderHeight - 1
		}

		spriteWidth := int(math.Abs(float64(r.renderHeight) / transformY))
		drawStartX := -spriteWidth/2 + spriteScreenX
		if drawStartX < 0 {
			drawStartX = 0
		}
		drawEndX := spriteWidth/2 + spriteScreenX
		if drawEndX >= r.renderWidth {
			drawEndX = r.renderWidth - 1
		}

		for xx := drawStartX; xx < drawEndX; xx++ {
			texX := int(256*(xx-(drawStartX-spriteWidth/2))*texSize/spriteWidth) / 256
			if transformY > 0 && xx > 0 && xx < r.renderWidth {
				for y := drawStartY; y < drawEndY; y++ {
					d := y*256 - r.renderHeight*128 + spriteHeight*128
					texY := ((d * texSize) / spriteHeight) / 256
					c := r.parent.(*Player).game.textureMap.RGBAAt(texX+texSize*2.5, texY%texSize)

					if c.R != 0 {
						if r.zBuffer[xx][y] > objectPerpDist {
							m.Set(xx, y, c)
						}
					}
				}
			}
		}
	}
}

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{0, 0, 0, 255}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
	}

	d.Dot = fixed.Point26_6{
		X: (fixed.I(img.Rect.Max.X) - d.MeasureString(label)) / 2,
		Y: fixed.I(y),
	}

	d.DrawString(label)
}

func (c *RenderView) renderPosition(img *image.RGBA) {

	addLabel(img, 10, 150, fmt.Sprintf("X: %f", c.parent.(*Player).getIntensityValuesAroundPlayer()[0]))

}
