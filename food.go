package main

var (
	FoodColor = [4]uint8{255, 0, 0, 0}
)

type Food struct {
	Sprite
}

func NewFood(position int) *Food {
	food := &Food{
		Sprite: Sprite{
			position: position,
			color:    FoodColor,
		},
	}

	return food
}

func (f *Food) Position() int {
	return f.position
}
