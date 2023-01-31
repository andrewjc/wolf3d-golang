package game

import (
	"github.com/faiface/pixel"
	"math"
)

// Functions associated with simulating sound propagation in the game world

func getIntensityValues(m [][]int, playerPos pixel.Vec) [][]float64 {
	var intensity [][]float64
	for i := 0; i < len(m); i++ {
		intensity = append(intensity, make([]float64, len(m[0])))
	}

	// Calculate intensity of sound at each grid cell based on distance from source and obstacles
	for i := 0; i < len(m); i++ {
		for j := 0; j < len(m[0]); j++ {
			intensity[i][j] = calcIntensity(i, j, playerPos.X, playerPos.Y, m)
		}
	}

	return intensity
}

func calcIntensity(x, y int, sourceX, sourceY float64, m [][]int) float64 {
	// Calculate intensity of sound at (x, y) based on distance from source and obstacles
	// For example, using the inverse square law:
	distance := math.Sqrt(math.Pow(float64(x-int(sourceX)), 2) + math.Pow(float64(y-int(sourceY)), 2))
	intensity := 1.0 / math.Pow(distance, 2)

	// Adjust intensity based on obstacles in the way
	//for i := x - 1; i <= x+1; i++ {
	//	for j := y - 1; j <= y+1; j++ {
	//		if i >= 0 && i < len(m) && j >= 0 && j < len(m[0]) && m[i][j] == 2 {
	//			intensity *= 0.5
	//		}
	//	}
	//}

	return intensity
}
