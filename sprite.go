package main

import (
	"fmt"
	"github.com/mpetavy/common"
	"github.com/veandco/go-sdl2/sdl"
	"reflect"
)

type Sprite struct {
	position int
	color    [4]uint8
}

func (s *Sprite) Paint() error {
	x := s.position % RasterCount
	y := s.position / RasterCount

	rect := sdl.Rect{int32(x * PixelWidth), int32(y * PixelWidth), PixelWidth, PixelWidth}

	err := renderer.SetDrawColor(s.color[0], s.color[1], s.color[2], s.color[3])
	if common.Error(err) {
		return err
	}

	err = renderer.FillRect(&rect)
	if common.Error(err) {
		return err
	}

	return nil
}

func contraint(v int, s int) int {
	if v+s < 0 {
		return RasterCount + (v + s)
	} else {
		return (v + s) % RasterCount
	}
}

func (s *Sprite) Peek(move func()) int {
	position := s.position
	defer func() {
		s.position = position
	}()

	move()

	return s.position
}

func (s *Sprite) Left() {
	common.DebugFunc(s.position)

	x := contraint(s.position%RasterCount, -1)
	y := s.position / RasterCount

	s.position = y*RasterCount + x
}

func (s *Sprite) Right() {
	common.DebugFunc(s.position)

	x := contraint(s.position%RasterCount, 1)
	y := s.position / RasterCount

	s.position = y*RasterCount + x
}

func (s *Sprite) Up() {
	common.DebugFunc(s.position)

	y := contraint(s.position/RasterCount, -1)
	x := s.position % RasterCount

	s.position = y*RasterCount + x
}

func (s *Sprite) Down() {
	common.DebugFunc(s.position)

	y := contraint(s.position/RasterCount, 1)
	x := s.position % RasterCount

	s.position = y*RasterCount + x
}

type positioner interface {
	Position() int
}

func ToPositions(arr interface{}) []int {
	positions := make([]int, 0)

	varr := reflect.ValueOf(arr)

	if varr.Kind() != reflect.Slice {
		panic("not a slice")
	}

	for i := 0; i < varr.Len(); i++ {
		p, ok := varr.Index(i).Interface().(positioner)

		if ok {
			positions = append(positions, p.Position())
		} else {
			panic(fmt.Errorf("cannot get position: %+v", varr.Index(i)))
		}
	}

	return positions
}

func Collides(pos int, positions ...int) bool {
	for _, position := range positions {
		if position == pos {
			return true
		}
	}

	return false
}
