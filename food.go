package main

import (
	"math/rand"
)

type Food struct {
	Position Point
}

func NewFood(snake *Snake, winWidth, winHeight float32) *Food {
	food := &Food{}

	for {
		food.Position.X = float32(rand.Intn(gridSize)) * (winWidth / gridSize)
		food.Position.Y = float32(rand.Intn(gridSize)) * (winHeight / gridSize)

		valid := true
		for _, point := range snake.Body {
			if point.X == food.Position.X && point.Y == food.Position.Y {
				valid = false
				break
			}
		}

		if valid {
			break
		}
	}

	return food
}

func (f *Food) Draw(program, vao uint32, winWidth, winHeight float32) {
	drawSquare(program, vao, indices, f.Position.X, f.Position.Y, winWidth, winHeight)
}
