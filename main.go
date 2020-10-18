package main

import (
	"fmt"
	"github.com/mpetavy/common"
	"github.com/veandco/go-sdl2/sdl"
	"runtime"
	"sync"
	"time"
)

const (
	RasterCount = 20
	PixelWidth  = 20
	GameDelay   = time.Millisecond * 150
	HungerDelay = time.Duration(2000) * time.Millisecond
	DeadDelay   = time.Duration(4000) * time.Millisecond
)

var (
	SnakeColor  = [4]uint8{0, 255, 0, 0}
	FoodColor   = [4]uint8{255, 0, 0, 0}
	HungerColor = [4]uint8{255, 165, 0, 0}
	TailColor   = [4]uint8{255, 255, 255, 0}
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
	Hunger   time.Duration
}

func NewSnake() *Snake {
	snake := &Snake{
		Puppet: Puppet{
			Pos:   RasterCount * RasterCount / 2,
			Color: SnakeColor,
		},
		Tail: []int{RasterCount * RasterCount / 2, (RasterCount * RasterCount / 2) - 1},
	}

	snake.LastMove = []func(){snake.Right, snake.Up, snake.Down}[common.Rnd(3)]

	return snake
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

func NewFood() *Food {
	food := &Food{
		Puppet: Puppet{
			Pos:   findNewFoodPos(),
			Color: FoodColor,
		},
	}

	return food
}

var (
	renderer *sdl.Renderer
	window   *sdl.Window
	snake    = NewSnake()
	food     = NewFood()
	gameOver = common.NewNotice()
)

func init() {
	common.Init(false, "1.0.0", "", "2020", "Snake game", "mpetavy", fmt.Sprintf("https://github.com/mpetavy/%s", common.Title()), common.APACHE, nil, nil, run, 0)
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
		Pos: 0,
	}

	if snake.Hunger > HungerDelay {
		tail.Color = HungerColor
	} else {
		tail.Color = TailColor
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
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

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

	err = paint()
	if common.Error(err) {
		return err
	}

	moves := make(chan func())
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for !gameOver.IsSet() {
			event := sdl.PollEvent()
			if event == nil {
				continue
			}

			switch t := event.(type) {
			case *sdl.QuitEvent:
				gameOver.Set()
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

	for !gameOver.IsSet() {
		var move func()

		select {
		case <-time.After(GameDelay):
			move = snake.LastMove

			snake.Hunger += GameDelay
			if snake.Hunger > DeadDelay {
				gameOver.Set()
				continue
			}

		case move = <-moves:
			if move == nil {
				gameOver.Set()
				continue
			}
		}

		lastPos := snake.Pos

		move()
		snake.LastMove = move

		if collides(snake.Pos, food.Pos) {
			snake.Tail = append(snake.Tail, lastPos)

			snake.Hunger = 0
			snake.Color = SnakeColor

			food.Pos = findNewFoodPos()
		}

		if len(snake.Tail) > 1 && collides(snake.Pos, snake.Tail[1:]...) {
			return nil
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

	wg.Wait()

	common.Info("Game over!!")

	return nil
}

func main() {
	defer common.Done()

	common.Run(nil)
}
