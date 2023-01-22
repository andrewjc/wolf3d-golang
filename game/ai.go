package game

import (
	"bytes"
	"encoding/json"
	"image/color"
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
	Observation_Pos []float64
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

const maxEpisodeLength = 5 * 60 * 1000

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

	//isNotMoving := !g.player1Controller.player.view.is_moving

	touchingWall := g.player1Controller.player.view.distanceToWall < 0.5 //|| g.distToNearestWall(g.player1Controller.player.view.position, 0.5) < 1.5

	done := g.player1Controller.player.isDone() || episodeLength > maxEpisodeLength || touchingWall

	if g.player1Controller.player.isDone() {
		print("Player is done", "\r\n")
	}
	if episodeLength > maxEpisodeLength {
		print("Episode length exceeded", "\r\n")
	}

	if touchingWall {
		print("Player is touching wall", "\r\n")
		reward = -2
	}

	return RLActionResult{Reward: reward, Observation: p10img, Observation_Pos: p1Obs, Done: done, Info: ""}
}

func (g *GameInstance) GetPlayer1Observation() ([]float64, []byte) {
	values := g.player1Controller.player.getIntensityValuesAroundPlayer()

	// Flatten the 2d array of values
	flatValues := make([]float64, len(values)*len(values[0]))
	for i := 0; i < len(values); i++ {
		for j := 0; j < len(values[i]); j++ {
			flatValues[i*len(values)+j] = float64(values[i][j])

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
	img := g.renderListener.renderBuffer
	if img == nil {
		return flatValues, nil
	}
	for i := 0; i < img.Rect.Dx(); i++ {
		for j := 0; j < img.Rect.Dy(); j++ {
			img.Set(i, j, color.Gray16Model.Convert(g.renderListener.renderBuffer.At(i, j)))
		}
	}

	// Compress the renderBuffer into JPEG
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 30})
	if err != nil {
		log.Println("Error compressing observation: ", err)
		return nil, nil
	} else {

		// Diff the p10bs observations with the previous ones
		// and return the diff as the observation
		var p1ObsDiff []float64
		if g.lastPlayer1Obs != nil {
			p1ObsDiff = make([]float64, len(flatValues))
			for i := range flatValues {
				// Calculate the percentage difference between the current and previous observation
				p1ObsDiff[i] = (flatValues[i] - g.lastPlayer1Obs[i]) / g.lastPlayer1Obs[i]

				if p1ObsDiff[0] > 0 {
					p1ObsDiff[0] = 0
				}
			}
		} else {
			p1ObsDiff = flatValues
		}
		g.lastPlayer1Obs = flatValues

		flatValues = p1ObsDiff

		return flatValues, buf.Bytes()

	}
}
