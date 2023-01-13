package game

import (
	"github.com/faiface/pixel"
	"math"
	"math/rand"
	"time"
)

type EnemyController struct {
	player              *Enemy
	lastDirectionChange time.Time
}

func (c EnemyController) update(delta float64) {
	x := int(c.player.getPosition().X)
	y := int(c.player.getPosition().Y)

	if c.player.game.mapData[y][x] != 0 {
		// If it is a wall, randomly choose a new direction
		c.player.view.direction = pixel.V(float64(rand.Float32()*2-1), float64(rand.Float32()*2-1))
	}

	// Generate a random noise value between 0 and 1
	noise := rand.Float32()

	// Update the velocity based on the noise value
	if noise > 0.9 {
		c.player.view.velocity += 0.1
		print("Speed up")
	}

	// Clamp the velocity to a minimum and maximum value
	c.player.view.velocity = float32(math.Max(0.1, math.Min(float64(c.player.view.velocity), 1.0)))
	print("Velocity: ", c.player.view.velocity)

	// Update the position based on the direction and velocity
	c.player.view.position = c.player.view.position.Add(c.player.view.direction.Scaled(float64(c.player.view.velocity * float32(delta))))

	// Check if the new position is a wall
	x = int(c.player.view.position.X)
	y = int(c.player.view.position.Y)
	if c.player.game.mapData[y][x] != 0 {
		// If it is a wall, set the position back to the previous position
		c.player.view.position = c.player.view.position.Sub(c.player.view.direction.Scaled(float64(c.player.view.velocity * float32(delta))))
	}
}
