package main

import (
	"github.com/mpetavy/common"
	"time"
)

var (
	SnakeColor = [4]uint8{0, 255, 0, 0}
)

type Snake struct {
	Sprite
	LastMove func()
	Tails    []*Tail
	Hunger   time.Duration
}

func NewSnake() *Snake {
	position := RasterCount * RasterCount / 2

	snake := &Snake{
		Sprite: Sprite{
			position: position,
			color:    SnakeColor,
		},
		Tails: []*Tail{NewTail(position), NewTail(position - 1)},
	}

	snake.LastMove = []func(){snake.Right, snake.Up, snake.Down}[common.Rnd(3)]

	return snake
}

func (s *Snake) Position() int {
	return s.position
}
