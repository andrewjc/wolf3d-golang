package game

import (
	"fmt"
	"github.com/fogleman/poissondisc"
	"math"
	"math/rand"
)

type Map struct {
	mapData [][]int
	visited [][]bool
	doors   [][]int
	rows    int
	cols    int
}

func (m *Map) GenerateMap() {
	// Initialize the map with all cells set to the wall cell type.
	m.mapData = make([][]int, m.rows)
	m.visited = make([][]bool, m.rows)
	for i := 0; i < m.rows; i++ {
		m.mapData[i] = make([]int, m.cols)
		m.visited[i] = make([]bool, m.cols)
		for j := 0; j < m.cols; j++ {
			m.mapData[i][j] = -1
		}
	}

	// Surround the map with the outside map boundary cell type.
	for i := 0; i < m.rows; i++ {
		m.mapData[i][0] = 1
		m.mapData[i][m.cols-1] = 1
	}
	for j := 0; j < m.cols; j++ {
		m.mapData[0][j] = 1
		m.mapData[m.rows-1][j] = 1
	}

	// Initialize the stack with the starting cell.
	//stack := []int{0, 0}
	m.visited[0][0] = true

	m.GenerateRooms(5)

	//m.GeneratePaths()

	// Replace all the -1 with 0
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			if m.mapData[i][j] == -1 {
				m.mapData[i][j] = 0
			}
		}
	}

	// Replace all the -1 with 0
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			if m.mapData[i][j] < 0 {
				m.mapData[i][j] = 0
			}
		}
	}

	//m.PrintMap()

}

func (m *Map) GenerateRooms(numRooms int) {
	maxRoomSize := 8
	minRoomSize := 6
	outerWallBoundary := 3

	// Generate the specified number of rooms.
	for i := 0; i < numRooms; i++ {

		roomGenerationSuccessful := false
		for !roomGenerationSuccessful {

			// Choose a random position and size for the room.
			x := rand.Intn(m.rows-outerWallBoundary) + outerWallBoundary
			y := rand.Intn(m.cols-outerWallBoundary) + outerWallBoundary
			w := rand.Intn(maxRoomSize) + minRoomSize
			h := rand.Intn(maxRoomSize) + minRoomSize

			// Check if any cell of the room overlaps with a cell of another room.
			overlaps := false
			for i := x; i < x+w; i++ {
				for j := y; j < y+h; j++ {
					if m.mapData[i][j] != -1 {
						overlaps = true
						break
					}
				}
				if overlaps {
					break
				}
			}

			// Skip room placement if the room overlaps with another room.
			if overlaps {
				continue
			}

			// Set the outer walls of the room to the room wall type.
			for i := x; i < x+w; i++ {
				m.mapData[i][y] = 4
				m.mapData[i][y+h-1] = 4
			}
			for j := y; j < y+h; j++ {
				m.mapData[x][j] = 4
				m.mapData[x+w-1][j] = 4
			}

			// Set the floor of the room to the walkable cell type.
			for i := x + 1; i < x+w-1; i++ {
				for j := y + 1; j < y+h-1; j++ {
					m.mapData[i][j] = -2 //This will be replace later, its used to mask the rooms
				}
			}

			// Choose a random wall of the room to be the door and set the door cell to the door type.
			doorPlaced := false
			for !doorPlaced {
				switch rand.Intn(4) {
				case 0:
					// Top wall.
					doorX := rand.Intn(w-2) + x + 1
					if m.mapData[doorX][y] == 4 {
						m.mapData[doorX][y] = 0
						m.doors = append(m.doors, []int{doorX, y})
						doorPlaced = true
					}
				case 1:
					// Bottom wall.
					doorX := rand.Intn(w-2) + x + 1
					if m.mapData[doorX][y+h-1] == 4 {
						m.mapData[doorX][y+h-1] = 0
						m.doors = append(m.doors, []int{doorX, y + h - 1})
						doorPlaced = true
					}
				case 2:
					// Left wall
					doorY := rand.Intn(h-2) + y + 1
					if m.mapData[x][doorY] == 4 {
						m.mapData[x][doorY] = 0
						m.doors = append(m.doors, []int{x, doorY})
						doorPlaced = true
					}
				case 3:
					// Right wall
					doorY := rand.Intn(h-2) + y + 1
					if m.mapData[x+w-1][doorY] == 4 {
						m.mapData[x+w-1][doorY] = 0
						m.doors = append(m.doors, []int{x + w - 1, doorY})
						doorPlaced = true
					}
				}
			}
			roomGenerationSuccessful = true
		}

	}
}

func (m *Map) PrintMap() {
	// Print the map.
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			switch m.mapData[i][j] {
			case 0:
				fmt.Print(" ")
			case 1:
				fmt.Print("#")
			case 2:
				fmt.Print("-")
			case 3:
				fmt.Print("D")
			case 4:
				fmt.Print("=")
			}
		}
		fmt.Println()
	}
}

func (m *Map) GeneratePaths() {
	var path [][]int
	for i := 0; i < len(m.doors)-1; i++ {
		start := m.doors[i]
		end := m.doors[i+1]
		x, y := start[0], start[1]
		endX, endY := end[0], end[1]
		for x != endX || y != endY {
			neighbors := getUnvisitedNeighbors(m, x, y)
			if len(neighbors) == 0 {
				break
			}
			next := m.getBestNeighbor(neighbors, endX, endY)
			path = append(path, next)
			x, y = next[0], next[1]
		}

		print("Path: ")
		for _, p := range path {
			if m.mapData[p[0]][p[1]] == -1 {
				m.mapData[p[0]][p[1]] = 2
			}
			print(p)
		}
	}

	//buildWalls(m, path)
}

func buildWalls(m *Map, path [][]int) {
	for _, pos := range path {
		x, y := pos[0], pos[1]
		if m.mapData[x][y] == 0 {
			addVerticalWall(m, x, y)
			addHorizontalWall(m, x, y)

			m.mapData[x][y] = 0
		}
	}
}

func addVerticalWall(m *Map, x, y int) {
	for i := x - 1; i <= x+1; i += 2 {
		if i >= 0 && i < m.rows && m.mapData[i][y] == 0 {
			m.mapData[i][y] = 2
		}
	}
}

func addHorizontalWall(m *Map, x, y int) {
	for i := y - 1; i <= y+1; i += 2 {
		if i >= 0 && i < m.cols && m.mapData[x][i] == 0 {
			m.mapData[x][i] = 2
		}
	}
}

func (m *Map) getBestNeighbor(neighbors [][]int, endX, endY int) []int {
	bestNeighbor := neighbors[0]
	bestScore := blueNoise(bestNeighbor[0], bestNeighbor[1], m.rows, m.cols)
	hScore := heuristic(bestNeighbor[0], bestNeighbor[1], endX, endY)
	bestScore += hScore
	for _, neighbor := range neighbors[1:] {
		score := blueNoise(neighbor[0], neighbor[1], m.rows, m.cols) + heuristic(neighbor[0], neighbor[1], endX, endY)
		if score < bestScore {
			bestNeighbor = neighbor
			bestScore = score
		}
	}
	return bestNeighbor
}

func blueNoise(x, y, max_x, max_y int) float64 {
	var x0, y0, x1, y1, r float64
	x0 = 0  // bbox min
	y0 = 0  // bbox min
	x1 = 2  // bbox max
	y1 = 2  // bbox max
	r = 10  // min distance between points
	k := 30 // max attempts to add neighboring point

	points := poissondisc.Sample(x0, y0, x1, y1, r, k, nil)

	// Choose random point
	randPoint := points[rand.Intn(len(points))]

	return randPoint.X
}

func heuristic(x, y, endX, endY int) float64 {
	// Calculate a heuristic value for (x, y) based on its distance to the end point (endX, endY)
	// For example, using the Euclidean distance:
	return math.Sqrt(math.Pow(float64(endX-x), 2) + math.Pow(float64(endY-y), 2))
}

func getUnvisitedNeighbors(m *Map, x, y int) [][]int {
	var neighbors [][]int
	if x > 0 && !m.visited[x-1][y] {
		neighbors = append(neighbors, []int{x - 1, y})
	}
	if x < m.rows-1 && !m.visited[x+1][y] {
		neighbors = append(neighbors, []int{x + 1, y})
	}
	if y > 0 && !m.visited[x][y-1] {
		neighbors = append(neighbors, []int{x, y - 1})
	}
	if y < m.cols-1 && !m.visited[x][y+1] {
		neighbors = append(neighbors, []int{x, y + 1})
	}
	return neighbors
}
