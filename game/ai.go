package game

import (
	"bytes"
	"encoding/json"
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

const maxEpisodeLength = 15 * 60 * 1000

func (g *GameInstance) TakePlayer1Action(action_id RLAction) RLActionResult {
	var reward float32 = 0
	if action_id == RLActionNone {
		//g.player1Controller.deaccelerateVelocity()
		//g.player1Controller.deaccelerateHorizontalVelocity()
	} else if action_id == RLActionMoveForward {
		//g.player1Controller.deaccelerateHorizontalVelocity()
		//g.player1Controller.accelerateForward()
		g.player1Controller.moveForward(0.5)

	} else if action_id == RLActionMoveBackward {
		//g.player1Controller.deaccelerateHorizontalVelocity()
		//g.player1Controller.accelerateBackward()
		g.player1Controller.moveBackwards(0.5)
	} else if action_id == RLActionStrafeLeft {
		//g.player1Controller.deaccelerateVelocity()
		//g.player1Controller.accelerateLeft()
		g.player1Controller.moveLeft(0.5)
	} else if action_id == RLActionStrafeRight {
		//g.player1Controller.deaccelerateVelocity()
		//g.player1Controller.accelerateRight()
		g.player1Controller.moveRight(0.5)
	} else if action_id == RLActionTurnLeft {
		//g.player1Controller.deaccelerateVelocity()
		//g.player1Controller.deaccelerateHorizontalVelocity()
		g.player1Controller.turnLeft(0.1)
	} else if action_id == RLActionTurnRight {
		//g.player1Controller.deaccelerateVelocity()
		//g.player1Controller.deaccelerateHorizontalVelocity()
		g.player1Controller.turnRight(0.1)
	} else {
		log.Fatal("Unknown action type ", action_id)
	}

	// update and save the player's position to p.player.view.old_position every 1000 frames
	distTravelled := math.Sqrt(math.Pow(g.player1Controller.player.game.player1Controller.player.old_position.X-g.player1Controller.player.game.player1Controller.player.view.position.X, 2) + math.Pow(g.player1Controller.player.game.player1Controller.player.old_position.Y-g.player1Controller.player.game.player1Controller.player.view.position.Y, 2))
	if g.player1Controller.player.game.currentTick-g.player1Controller.player.game.episodeStartTick > 100 {
		if g.player1Controller.player.game.currentTick-g.player1Controller.player.game.lastPlayer1PositionUpdateTick > (3 * 1000) {

			//print("distTravelled: ", distTravelled, "\r\n")

			// set is_moving=True if the euclidian distance between old and new positions is greater than 1
			if distTravelled < 1 {
				g.player1Controller.player.game.player1Controller.player.is_moving = false
			} else {
				g.player1Controller.player.game.player1Controller.player.is_moving = true
			}
			g.player1Controller.player.game.player1Controller.player.old_position = g.player1Controller.player.game.player1Controller.player.view.position
			g.player1Controller.player.game.lastPlayer1PositionUpdateTick = g.player1Controller.player.game.currentTick
		}
	} else {
		g.player1Controller.player.game.player1Controller.player.is_moving = true
		if g.player1Controller.player.game.lastPlayer1PositionUpdateTick == 0 {
			g.player1Controller.player.game.lastPlayer1PositionUpdateTick = g.player1Controller.player.game.currentTick
		}
	}

	g.player1Controller.player.view.render()

	reward = g.player1Controller.player.getReward()

	p1Obs, p10img := g.GetPlayer1Observation()

	episodeLength := g.currentTick - g.episodeStartTick

	isNotMoving := !g.player1Controller.player.is_moving

	//touchingWall := g.player1Controller.player.view.distanceToWall < 0.5 //|| g.distToNearestWall(g.player1Controller.player.view.position, 0.5) < 1.5

	done := g.player1Controller.player.isDone() || episodeLength > maxEpisodeLength || isNotMoving

	if g.player1Controller.player.isDone() {
		print("Player is done", "\r\n")
	}
	if episodeLength > maxEpisodeLength {
		print("Episode length exceeded", "\r\n")
	}

	if isNotMoving {
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

	// lock and synchronise the renderBuffer
	g.renderListener.renderBufferMutex.Lock()
	defer g.renderListener.renderBufferMutex.Unlock()
	img := g.renderListener.renderBuffer
	if img == nil {
		return flatValues, nil
	}

	// Compress the renderBuffer into JPEG
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, g.renderListener.renderBuffer, &jpeg.Options{Quality: 60})
	if err != nil {
		log.Println("Error compressing observation: ", err)
		return nil, nil
	} else {

		return flatValues, buf.Bytes()

	}
}
