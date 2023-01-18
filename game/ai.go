package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"math"
)

type RLAction uint8

const (
	RLActionNone RLAction = iota
	RLActionMoveForward
	RLActionMoveBackward
	RLActionStrafeLeft
	RLActionStrafeRight
	RLActionTurnLeft
	RLActionTurnRight
)

type RLActionResult struct {
	Reward          float32
	Observation     []uint8
	Observation_Pos []float32
	Done            bool
	Info            string
}

func (r *RLActionResult) ToJson() *string {
	// Implement the json serializer method
	b, err := json.Marshal(r)
	if err != nil {
		log.Println("Error serializing RLActionResult: ", err)
		return nil
	} else {
		s := string(b)
		return &s
	}
}

const maxEpisodeLength = 60 * 1000

func (g *GameInstance) TakePlayer1Action(action_id RLAction) RLActionResult {
	var reward float32 = 0
	if action_id == RLActionNone {
		g.player1Controller.deaccelerateVelocity()
		g.player1Controller.deaccelerateHorizontalVelocity()
	} else if action_id == RLActionMoveForward {
		g.player1Controller.deaccelerateHorizontalVelocity()
		g.player1Controller.accelerateForward()

	} else if action_id == RLActionMoveBackward {
		g.player1Controller.deaccelerateHorizontalVelocity()
		g.player1Controller.accelerateBackward()
	} else if action_id == RLActionStrafeLeft {
		g.player1Controller.deaccelerateVelocity()
		g.player1Controller.accelerateLeft()
	} else if action_id == RLActionStrafeRight {
		g.player1Controller.deaccelerateVelocity()
		g.player1Controller.accelerateRight()
	} else if action_id == RLActionTurnLeft {
		g.player1Controller.deaccelerateVelocity()
		g.player1Controller.deaccelerateHorizontalVelocity()
		g.player1Controller.turnLeft(0.1)
	} else if action_id == RLActionTurnRight {
		g.player1Controller.deaccelerateVelocity()
		g.player1Controller.deaccelerateHorizontalVelocity()
		g.player1Controller.turnRight(0.1)
	} else {
		log.Fatal("Unknown action type ", action_id)
	}

	// update and save the player's position to p.player.view.old_position every 1000 frames
	if g.currentTick-g.episodeStartTick > 1000 {
		if g.currentTick-g.lastPlayer1PositionUpdateTick > (15 * 1000) {
			g.player1Controller.player.view.old_position = g.player1Controller.player.view.position
			g.lastPlayer1PositionUpdateTick = g.currentTick

			// set is_moving=True if the euclidian distance between old and new positions is greater than 1
			distTravelled := math.Sqrt(math.Pow(g.player1Controller.player.view.old_position.X-g.player1Controller.player.view.position.X, 2) + math.Pow(g.player1Controller.player.view.old_position.Y-g.player1Controller.player.view.position.Y, 2))
			if distTravelled < 1e-5 {
				g.player1Controller.player.view.is_moving = false
				//print("Player is not moving", distTravelled, "\r\n")
			}
		}
	} else {
		if g.lastPlayer1PositionUpdateTick == 0 {
			g.lastPlayer1PositionUpdateTick = g.currentTick
		}
		distTravelled := math.Sqrt(math.Pow(g.player1Controller.player.view.old_position.X-g.player1Controller.player.view.position.X, 2) + math.Pow(g.player1Controller.player.view.old_position.Y-g.player1Controller.player.view.position.Y, 2))
		if distTravelled > 1e-5 {
			g.player1Controller.player.view.is_moving = true
			//print("Player is moving", distTravelled, "\r\n")
		}
	}

	reward = g.player1Controller.player.getReward()

	p1Obs, p10img := g.GetPlayer1Observation()
	episodeLength := g.currentTick - g.episodeStartTick

	isNotMoving := !g.player1Controller.player.view.is_moving

	done := g.player1Controller.player.isDone() || episodeLength > maxEpisodeLength || isNotMoving

	return RLActionResult{Reward: reward, Observation: p10img, Observation_Pos: p1Obs, Done: done, Info: ""}
}

func (g *GameInstance) GetPlayer1Observation() ([]float32, []byte) {
	values := g.player1Controller.player.getIntensityValuesAroundPlayer()

	// Flatten the 2d array of values
	flatValues := make([]float32, len(values)*len(values[0]))
	for i := 0; i < len(values); i++ {
		for j := 0; j < len(values[i]); j++ {
			flatValues[i*len(values)+j] = float32((values[i][j]))

			// Cap the value between 0 and 1
			if flatValues[i*len(values)+j] > 1 {
				flatValues[i*len(values)+j] = 1
			}

			if flatValues[i*len(values)+j] < 0 {
				flatValues[i*len(values)+j] = 0
			}
		}
	}

	// Convert g.renderListener.renderBuffer into grayscale
	img := image.NewGray(g.renderListener.renderBuffer.Bounds())
	for i := 0; i < img.Rect.Dx(); i++ {
		for j := 0; j < img.Rect.Dy(); j++ {
			img.Set(i, j, color.Gray16Model.Convert(g.renderListener.renderBuffer.At(i, j)))
		}
	}

	// Draw text onto the image in the top left with a black background
	draw.Draw(img, img.Bounds(), image.NewUniform(color.Black), image.ZP, draw.Src)
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.ZP, draw.Over)
	f, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal(err)
	}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: truetype.NewFace(f, &truetype.Options{Size: 12}),
		Dot:  fixed.P(0, 12),
	}
	d.DrawString(fmt.Sprintf("Player 1: %v", g.player1Controller.player.view.position))

	// Compress the renderBuffer into JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 30})
	if err != nil {
		log.Println("Error compressing observation: ", err)
		return nil, nil
	} else {
		return flatValues, buf.Bytes()
	}
}
