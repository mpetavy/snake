package main

import (
	"fmt"
	"github.com/mpetavy/common"
	"github.com/veandco/go-sdl2/sdl"
	"runtime"
	"time"
)

const (
	RasterCount = 20
	PixelWidth  = 20
)

type Puppet struct {
	Pos      int
	LastMove func()
	Color    [4]uint8
}

func (p *Puppet) Paint() error {
	x := p.Pos % RasterCount
	y := p.Pos / RasterCount

	rect := sdl.Rect{int32(x * PixelWidth), int32(y * PixelWidth), PixelWidth, PixelWidth}

	renderer.SetDrawColor(p.Color[0],p.Color[1],p.Color[2],p.Color[3],)

	return renderer.FillRect(&rect)
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
	Tail []Puppet
}

func (s *Snake) Bites(ps ...Puppet) bool {
	for _, p := range ps {
		if p.Pos == s.Pos {
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
			Color: [4]uint8{255,255,255,0},
		},
	}
	food = Food{
		Puppet: Puppet{
			Pos:   common.Rnd(RasterCount * RasterCount),
			Color: [4]uint8{255,0,0,0},
		},
	}

	delay = time.Millisecond * 100
	nextTime = time.Now()
)

func init() {
	common.Init(false, "1.0.8", "", "2020", "Snake game", "mpetavy", fmt.Sprintf("https://github.com/mpetavy/%s", common.Title()), common.APACHE, nil, nil, run, 0)
}

func paint() {
	common.DebugFunc()

	renderer.SetDrawColor(0,0,0,0)
	renderer.Clear()

	food.Paint()
	snake.Paint()

	renderer.Present()
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

	defer window.Destroy()

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if common.Error(err) {
		return err
	}

	defer renderer.Destroy()

	runtime.LockOSThread()

	running := true

	snake.LastMove = []func(){snake.Left, snake.Right, snake.Up, snake.Down}[common.Rnd(4)]

	paint()

	for running {
		move := snake.LastMove

		event := sdl.PollEvent()
		if move == nil {
			if event != nil {
				switch t := event.(type) {
				case *sdl.QuitEvent:
					println("Quit")
					running = false
				case *sdl.KeyboardEvent:
					keyCode := t.Keysym.Sym

					if t.State == sdl.PRESSED {
						switch keyCode {
						case sdl.K_LEFT:
							move = snake.Left
						case sdl.K_RIGHT:
							move = snake.Right
						case sdl.K_UP:
							move = snake.Up
						case sdl.K_DOWN:
							move = snake.Down
						}
					}
				}
			}
		}

		if move == nil || time.Now().Before(nextTime){
			continue
		}

		nextTime = nextTime.Add(delay)

		move()
		snake.LastMove = move

		move = nil

		paint()

		//if len(snake.Tail) > 1 && snake.Bites(snake.Tail[1:]...) {
		//	common.Info("game over")
		//	os.Exit(0)
		//}

		if snake.Bites(food.Puppet) {
			tail := append(append(snake.Tail, food.Puppet), snake.Tail...)
			snake.Tail = tail
		}

		//sdl.Delay(uint32(delay))
	}

	return nil
}

func main() {
	defer common.Done()

	common.Run(nil)
}
