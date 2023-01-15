package game

import (
    "time"
)

type EnemyController struct {
    player              *Enemy
    lastDirectionChange time.Time
}

func (c EnemyController) update(delta float64) {

}
