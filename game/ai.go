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
	Reward      float64
	Observation []uint8
	Done        bool
	Info        string
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
	var reward float64 = 0
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
	//if action_id == RLActionTurnLeft || action_id == RLActionTurnRight {
	//	reward = 0.01
	//}

	p1Obs := g.GetPlayer1Observation()
	episodeLength := g.currentTick - g.episodeStartTick

	isNotMoving := !g.player1Controller.player.view.is_moving

	//print("Episode length: ", episodeLength, "\r\n")
	//print("isNotMoving: ", isNotMoving, "\r\n")

	done := g.player1Controller.player.isDone() || episodeLength > maxEpisodeLength || isNotMoving

	return RLActionResult{Reward: reward, Observation: p1Obs, Done: done, Info: ""}
}

func (g *GameInstance) GetPlayer1Observation() []byte {
	// Compress the renderBuffer into JPEG
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, g.renderListener.renderBuffer, &jpeg.Options{Quality: 30})
	if err != nil {
		log.Println("Error compressing observation: ", err)
		return nil
	} else {
		return buf.Bytes()
	}
}
