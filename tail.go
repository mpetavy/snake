package main

var (
	HungerColor = [4]uint8{255, 165, 0, 0}
	TailColor   = [4]uint8{255, 255, 255, 0}
)

type Tail struct {
	Sprite
}

func NewTail(position int) *Tail {
	tail := &Tail{
		Sprite: Sprite{
			position: position,
			color:    TailColor,
		},
	}

	return tail
}

func (t *Tail) Position() int {
	return t.position
}
