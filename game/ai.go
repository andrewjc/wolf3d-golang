package game

import (
	"bytes"
	"encoding/json"
	"image/jpeg"
	"log"
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
		reward = 0.0
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

	reward = g.player1Controller.player.getReward()
	p1Obs := g.GetPlayer1Observation()
	episodeLength := g.currentTick - g.episodeStartTick

	print("Episode length: ", episodeLength)

	done := g.player1Controller.player.isDone()

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
