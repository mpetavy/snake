package main

var (
	StoneColor = [4]uint8{169, 169, 169, 0}
)

type Stone struct {
	Sprite
}

func NewStone(position int) *Stone {
	stone := &Stone{
		Sprite: Sprite{
			position: position,
			color:    StoneColor,
		},
	}

	return stone
}

func (s *Stone) Position() int {
	return s.position
}
