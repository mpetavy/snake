package main

import (
	"fmt"
	"github.com/mpetavy/common"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"runtime"
	"sync"
	"time"
)

const (
	RasterCount   = 20
	PixelWidth    = 20
	StoneCount    = RasterCount / 2
	GameDelay     = time.Millisecond * 150
	DeadDelay     = GameDelay * time.Duration((RasterCount+2)*2)
	HungerDelay   = DeadDelay / 2
	TitleDuration = 2 * time.Second
)

var (
	SnakeColor  = [4]uint8{0, 255, 0, 0}
	FoodColor   = [4]uint8{255, 0, 0, 0}
	HungerColor = [4]uint8{255, 165, 0, 0}
	TailColor   = [4]uint8{255, 255, 255, 0}
	StoneColor  = [4]uint8{169, 169, 169, 0}
)

type Sprite struct {
	Position int
	Color    [4]uint8
}

func (s *Sprite) Paint() error {
	x := s.Position % RasterCount
	y := s.Position / RasterCount

	rect := sdl.Rect{int32(x * PixelWidth), int32(y * PixelWidth), PixelWidth, PixelWidth}

	err := renderer.SetDrawColor(s.Color[0], s.Color[1], s.Color[2], s.Color[3])
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

func (s *Sprite) Left() {
	common.DebugFunc(s.Position)

	x := contraint(s.Position%RasterCount, -1)
	y := s.Position / RasterCount

	s.Position = y*RasterCount + x
}

func (s *Sprite) Right() {
	common.DebugFunc(s.Position)

	x := contraint(s.Position%RasterCount, 1)
	y := s.Position / RasterCount

	s.Position = y*RasterCount + x
}

func (s *Sprite) Up() {
	common.DebugFunc(s.Position)

	y := contraint(s.Position/RasterCount, -1)
	x := s.Position % RasterCount

	s.Position = y*RasterCount + x
}

func (s *Sprite) Down() {
	common.DebugFunc(s.Position)

	y := contraint(s.Position/RasterCount, 1)
	x := s.Position % RasterCount

	s.Position = y*RasterCount + x
}

type Snake struct {
	Sprite
	LastMove func()
	Tail     []int
	Hunger   time.Duration
}

func NewSnake() *Snake {
	snake := &Snake{
		Sprite: Sprite{
			Position: RasterCount * RasterCount / 2,
			Color:    SnakeColor,
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
	Sprite
}

func NewFood() *Food {
	food := &Food{
		Sprite: Sprite{
			Position: findFreePosition(),
			Color:    FoodColor,
		},
	}

	return food
}

type Stone struct {
	Sprite
}

func NewStone() *Stone {
	stone := &Stone{
		Sprite: Sprite{
			Position: findFreePosition(),
			Color:    StoneColor,
		},
	}

	return stone
}

var (
	renderer *sdl.Renderer
	window   *sdl.Window
	snake    *Snake
	food     *Food
	gameOver *common.Notice
	stones   []int
)

func init() {
	common.Init(false, "1.0.0", "", "2020", "Snake game", "mpetavy", fmt.Sprintf("https://github.com/mpetavy/%s", common.Title()), common.APACHE, nil, nil, run, 0)
}

func findFreePosition() int {
	for {
		pos := common.Rnd(RasterCount * RasterCount)

		if !collides(pos, snake.Position) && !collides(pos, snake.Tail...) && !collides(pos, stones...) {
			return pos
		}
	}
}

func newScene() error {
	common.DebugFunc()

	err := renderer.SetDrawColor(0, 0, 0, 0)
	if common.Error(err) {
		return err
	}

	err = renderer.Clear()
	if common.Error(err) {
		return err
	}

	return nil
}

func paintScene() error {
	common.DebugFunc()

	err := newScene()
	if common.Error(err) {
		return err
	}

	err = food.Paint()
	if common.Error(err) {
		return err
	}

	tail := Sprite{
		Position: 0,
	}

	if snake.Hunger > HungerDelay {
		tail.Color = HungerColor
	} else {
		tail.Color = TailColor
	}

	for _, t := range snake.Tail {
		tail.Position = t
		err := tail.Paint()
		if common.Error(err) {
			return err
		}
	}

	stone := Sprite{
		Position: 0,
		Color:    StoneColor,
	}

	for _, s := range stones {
		stone.Position = s
		err := stone.Paint()
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

func drawTitle(r *sdl.Renderer, text string, size int) error {
	err := r.Clear()
	if common.Error(err) {
		return err
	}

	f, err := ttf.OpenFont("res/fonts/Flappy.ttf", size)
	if common.Error(err) {
		return err
	}
	defer f.Close()

	c := sdl.Color{R: 255, G: 100, B: 0, A: 255}
	s, err := f.RenderUTF8Solid(text, c)
	if common.Error(err) {
		return err
	}
	defer s.Free()

	t, err := r.CreateTextureFromSurface(s)
	if common.Error(err) {
		return err
	}
	defer func() {
		common.Error(t.Destroy())
	}()

	err = r.Copy(t, nil, nil)
	if common.Error(err) {
		return err
	}

	r.Present()

	return nil
}

func run() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	defer sdl.Quit()

	if err := ttf.Init(); err != nil {
		return fmt.Errorf("could not initialize TTF: %v", err)
	}
	defer ttf.Quit()

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

	snake = NewSnake()
	food = NewFood()
	gameOver = common.NewNotice()
	stones = make([]int, StoneCount)

	for i := 0; i < len(stones); i++ {
		stones[i] = findFreePosition()
	}

	err = drawTitle(renderer, common.Title(), 20)
	if common.Error(err) {
		return err
	}

	time.Sleep(TitleDuration)

	err = paintScene()
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

	defer func() {
		wg.Wait()

		for i := 0; i < 10; i++ {
			if i%2 == 0 {
				snake.Color = HungerColor
			} else {
				snake.Color = FoodColor
			}

			common.Error(paintScene())

			time.Sleep(200 * time.Millisecond)
		}

		common.Error(newScene())
		common.Error(drawTitle(renderer, "Game over!!", 10))

		time.Sleep(TitleDuration)

		common.Error(newScene())
		common.Error(drawTitle(renderer, fmt.Sprintf("Score: %d", (len(snake.Tail)-2)*10), 10))

		time.Sleep(TitleDuration)
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

		lastPos := snake.Position

		move()
		snake.LastMove = move

		if collides(snake.Position, food.Position) {
			snake.Tail = append(snake.Tail, lastPos)

			snake.Hunger = 0
			snake.Color = SnakeColor

			food.Position = findFreePosition()
		}

		if len(snake.Tail) > 1 && collides(snake.Position, snake.Tail[1:]...) {
			gameOver.Set()

			continue
		}

		if len(snake.Tail) > 0 {
			if len(snake.Tail) > 1 {
				copy(snake.Tail[1:], snake.Tail[0:len(snake.Tail)-1])
			}
			snake.Tail[0] = snake.Position
		}

		err = paintScene()
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
