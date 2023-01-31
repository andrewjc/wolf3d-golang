package game

import (
	"github.com/faiface/pixel"
	"math"
)

type point struct {
	x, y int
	//parent *point
	path []pixel.Vec
}

// implement == and != operators on point
func (p *point) Equal(p2 *point) bool {
	return p.x == p2.x && p.y == p2.y
}

// THIS DOESN'T WORK GRUMBLE GRUMBLE FIX ME

func (g *GameInstance) autoPlan(controller *PlayerController, controller2 *EnemyController, dt float64) {

	if g.planPath == nil {
		g.planPath = make([]pixel.Vec, 0)
		startPos := controller.player.getPosition()
		destination := controller2.player.getPosition()

		g.planPath = plotPathToDestination(g.mapData, startPos, destination)
	}

	if len(g.planPath) == 0 {
		return
	}
	// get the next point in the path
	next := g.planPath[0]
	// check if the player has reached the next point in the path
	nPos := next.Sub(controller.player.getPosition()).Len()
	if nPos < 2 {
		g.planPath = g.planPath[1:]
		return
	}
	// calculate the direction vector towards the next point in the path
	///direction := next.Sub(controller.player.getPosition()).Normal()

	// rotate the player towards the direction vector
	targetVec := next.Sub(controller.player.getPosition())
	// calculate the angle between the direction vector and the target vector
	angle := math.Acos(controller.player.getRotation().Dot(targetVec) / (controller.player.getRotation().Len() * targetVec.Len()))

	// convert radians to degrees
	angle = angle * 180 / math.Pi

	print(angle, "\r\n")

	// rotate the player towards the next point in the path
	if angle > 1 {
		controller.turnLeft(0.005)
		return
	}

	if angle < 0 {
		controller.turnRight(0.005)
		return
	}

	controller.accelerateForward()

}

func plotPathToDestination(mapData [][]int, startPos pixel.Vec, destination pixel.Vec) []pixel.Vec {
	// A* pathfinding algorithm
	start := point{int(startPos.X), int(startPos.Y), []pixel.Vec{startPos}}
	dest := point{int(destination.X), int(destination.Y), nil}
	openSet := map[*point]bool{&start: true}
	closedSet := map[*point]bool{}
	for len(openSet) > 0 {
		// find the point in openSet with the smallest distance to the destination
		current := point{-1, -1, nil}
		for p := range openSet {
			if current.x == -1 || p.distToDestination(pixel.Vec{X: float64(dest.x), Y: float64(dest.y)}) < current.distToDestination(pixel.V(float64(dest.x), float64(dest.y))) {
				current = *p
			}
		}

		if current.Equal(&dest) {
			path := current.path
			path = append(path, pixel.Vec{X: float64(current.x), Y: float64(current.y)})
			return path
		}
		//delete(openSet, &current)
		closedSet[&current] = true
		current.addNeighbors(mapData, openSet, closedSet)
	}
	return []pixel.Vec{}
}

func (p *point) addNeighbors(mapData [][]int, openSet, closedSet map[*point]bool) {
	// check all 8 possible neighbors, including diagonals
	neighbors := []*point{&point{p.x - 1, p.y, nil}, &point{p.x + 1, p.y, nil}, &point{p.x, p.y - 1, nil}, &point{p.x, p.y + 1, nil},
		&point{p.x - 1, p.y - 1, nil}, &point{p.x - 1, p.y + 1, nil}, &point{p.x + 1, p.y - 1, nil}, &point{p.x + 1, p.y + 1, nil}}
	for _, neighbor := range neighbors {
		if neighbor.x < 0 || neighbor.y < 0 || neighbor.x >= len(mapData) || neighbor.y >= len(mapData[0]) ||
			mapData[neighbor.x][neighbor.y] != 0 || closedSet[neighbor] {
			continue
		}
		if _, ok := closedSet[neighbor]; !ok && mapData[neighbor.x][neighbor.y] == 0 {
			neighbor.path = append(p.path, pixel.Vec{X: float64(p.x), Y: float64(p.y)})
			openSet[neighbor] = true
		}
	}
}

func (p *point) distToDestination(destination pixel.Vec) float64 {
	// Manhattan distance heuristic
	dist := math.Abs(float64(p.x)-destination.X) + math.Abs(float64(p.y)-destination.Y)
	return dist
}
