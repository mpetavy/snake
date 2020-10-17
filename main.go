package main

import (
	"fmt"
	"github.com/mpetavy/common"
	"github.com/veandco/go-sdl2/sdl"
	"os"
	"runtime"
	"time"
)

const (
	RasterCount = 20
	PixelWidth  = 20
)

type Puppet struct {
	Pos   int
	Color [4]uint8
}

func (p *Puppet) Paint() error {
	x := p.Pos % RasterCount
	y := p.Pos / RasterCount

	rect := sdl.Rect{int32(x * PixelWidth), int32(y * PixelWidth), PixelWidth, PixelWidth}

	err := renderer.SetDrawColor(p.Color[0], p.Color[1], p.Color[2], p.Color[3])
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

func (p *Puppet) Left() {
	common.DebugFunc(p.Pos)

	x := contraint(p.Pos%RasterCount, -1)
	y := p.Pos / RasterCount

	p.Pos = y*RasterCount + x
}

func (p *Puppet) Right() {
	common.DebugFunc(p.Pos)

	x := contraint(p.Pos%RasterCount, 1)
	y := p.Pos / RasterCount

	p.Pos = y*RasterCount + x
}

func (p *Puppet) Up() {
	common.DebugFunc(p.Pos)

	y := contraint(p.Pos/RasterCount, -1)
	x := p.Pos % RasterCount

	p.Pos = y*RasterCount + x
}

func (p *Puppet) Down() {
	common.DebugFunc(p.Pos)

	y := contraint(p.Pos/RasterCount, 1)
	x := p.Pos % RasterCount

	p.Pos = y*RasterCount + x
}

type Snake struct {
	Puppet
	LastMove func()
	Tail     []int
}

func collides(pos int, ps ...int) bool {
	for _, v := range ps {
		if v == pos {
			common.DebugFunc()
			return true
		}
	}

	return false
}

type Food struct {
	Puppet
}

var (
	renderer *sdl.Renderer
	window   *sdl.Window
	snake    = Snake{
		Puppet: Puppet{
			Pos:   0,
			Color: [4]uint8{0, 255, 0, 0},
		},
		Tail: make([]int, 1),
	}
	food = Food{
		Puppet: Puppet{
			Pos:   findNewFoodPos(),
			Color: [4]uint8{255, 0, 0, 0},
		},
	}
)

func init() {
	common.Init(false, "1.0.8", "", "2020", "Snake game", "mpetavy", fmt.Sprintf("https://github.com/mpetavy/%s", common.Title()), common.APACHE, nil, nil, run, 0)
}

func findNewFoodPos() int {
	for {
		pos := common.Rnd(RasterCount * RasterCount)

		if !collides(pos, snake.Pos) && !collides(pos, snake.Tail...) {
			return pos
		}
	}
}

func paint() error {
	common.DebugFunc()

	err := renderer.SetDrawColor(0, 0, 0, 0)
	if common.Error(err) {
		return err
	}

	err = renderer.Clear()
	if common.Error(err) {
		return err
	}

	err = food.Paint()
	if common.Error(err) {
		return err
	}

	tail := Puppet{
		Pos:   0,
		Color: [4]uint8{255, 255, 255, 0},
	}

	for _, t := range snake.Tail {
		tail.Pos = t
		err := tail.Paint()
		if common.Error(err) {
			return err
		}
	}

	err = snake.Paint()
	if common.Error(err) {
		return err
	}

	renderer.Present()

	return nil
}

func run() error {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	defer sdl.Quit()

	var err error

	window, err = sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		RasterCount*PixelWidth, RasterCount*PixelWidth, sdl.WINDOW_SHOWN)
	if common.Error(err) {
		return err
	}

	defer func() {
		common.Error(window.Destroy())
	}()

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if common.Error(err) {
		return err
	}

	defer func() {
		common.Error(renderer.Destroy())
	}()

	runtime.LockOSThread()

	snake.LastMove = []func(){snake.Left, snake.Right, snake.Up, snake.Down}[common.Rnd(4)]

	err = paint()
	if common.Error(err) {
		return err
	}

	moves := make(chan func())

	go func() {
		running := true

		for running {
			event := sdl.PollEvent()
			if event == nil {
				continue
			}

			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
				continue
			case *sdl.KeyboardEvent:
				keyCode := t.Keysym.Sym

				if t.State == sdl.PRESSED {
					switch keyCode {
					case sdl.K_LEFT:
						moves <- snake.Left
					case sdl.K_RIGHT:
						moves <- snake.Right
					case sdl.K_UP:
						moves <- snake.Up
					case sdl.K_DOWN:
						moves <- snake.Down
					}
				}
			}
		}

		close(moves)
	}()

	running := true
	for running {
		var move func()

		select {
		case <-time.After(time.Millisecond * 200):
			move = snake.LastMove
		case move = <-moves:
			if move == nil {
				running = false
				continue
			}
		}

		lastPos := snake.Pos

		move()
		snake.LastMove = move

		if collides(snake.Pos, food.Pos) {
			snake.Tail = append(snake.Tail, lastPos)

			food.Pos = findNewFoodPos()
		}

		if len(snake.Tail) > 1 && collides(snake.Pos, snake.Tail[1:]...) {
			common.Info("game over")
			os.Exit(0)
		}

		if len(snake.Tail) > 0 {
			if len(snake.Tail) > 1 {
				copy(snake.Tail[1:], snake.Tail[0:len(snake.Tail)-1])
			}
			snake.Tail[0] = snake.Pos
		}

		err = paint()
		if common.Error(err) {
			return err
		}

	}

	return nil
}

func main() {
	defer common.Done()

	common.Run(nil)
}
