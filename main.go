package main

import (
	"embed"
	"fmt"
	"github.com/mpetavy/common"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const (
	RasterCount   = 20
	PixelWidth    = 20
	GameDelay     = time.Millisecond * 150
	DeadDelay     = GameDelay * time.Duration((RasterCount+2)*2)
	HungerDelay   = DeadDelay / 2
	TitleDuration = 2 * time.Second
)

var (
	filename string
	renderer *sdl.Renderer
	window   *sdl.Window
	running  *common.Notice
	ranking  *Ranking
	snake    *Snake
	food     *Food
	stones   []*Stone

	//go:embed res
	res embed.FS
)

//go:embed go.mod
var resources embed.FS

func init() {
	common.Init("", "", "", "", "Snake game", "", "", "", &resources, nil, nil, run, 0)
}

func findFreePosition() int {
	for {
		pos := common.Rnd(RasterCount * RasterCount)

		if !Collides(pos, snake.Position()) && (snake.Tails == nil || !Collides(pos, ToPositions(snake.Tails)...)) && (stones == nil || !Collides(pos, ToPositions(stones)...)) {
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

	for _, tail := range snake.Tails {
		if snake.Hunger > HungerDelay {
			tail.color = HungerColor
		} else {
			tail.color = TailColor
		}

		err := tail.Paint()
		if common.Error(err) {
			return err
		}
	}

	for _, stone := range stones {
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

func paintTitle(r *sdl.Renderer, text string, size int) error {
	err := r.Clear()
	if common.Error(err) {
		return err
	}

	ba, err := res.ReadFile("res/fonts/Flappy.ttf")
	if common.Error(err) {
		return err
	}

	tempf, err := common.CreateTempFile()
	if common.Error(err) {
		return err
	}

	err = os.WriteFile(tempf.Name(), ba, os.ModePerm)
	if common.Error(err) {
		return err
	}

	f, err := ttf.OpenFont(tempf.Name(), size)
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
	var err error

	filename, err = os.UserHomeDir()
	if common.Error(err) {
		return err
	}

	filename = filepath.Join(filename, common.AppFilename("-ranking.json"))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ranking = LoadlRanking()

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	defer sdl.Quit()

	if err := ttf.Init(); err != nil {
		return fmt.Errorf("could not initialize TTF: %v", err)
	}
	defer ttf.Quit()

	window, err = sdl.CreateWindow(common.TitleVersion(true, true, true), sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
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
	food = NewFood(findFreePosition())
	running = common.NewNotice()
	stones = make([]*Stone, 0)

	stones = append(stones, NewStone(findFreePosition()))

	err = paintTitle(renderer, common.Title(), 20)
	if common.Error(err) {
		return err
	}

	common.Sleep(TitleDuration)

	err = paintScene()
	if common.Error(err) {
		return err
	}

	moves := make(chan func())
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer common.UnregisterGoRoutine(common.RegisterGoRoutine(1))

		defer wg.Done()

		for running.IsSet() {
			event := sdl.PollEvent()
			if event == nil {
				continue
			}

			switch t := event.(type) {
			case *sdl.QuitEvent:
				running.Unset()
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

		for i := 0; i < 5; i++ {
			if i%2 == 0 {
				snake.color = HungerColor
			} else {
				snake.color = FoodColor
			}

			common.Error(paintScene())

			common.Sleep(200 * time.Millisecond)
		}

		common.Error(newScene())
		common.Error(paintTitle(renderer, "Game over!!", 10))

		common.Sleep(TitleDuration)

		points := (len(snake.Tails) - 2) * 10

		err, highscore := ranking.Score(points)
		common.Error(err)

		if highscore {
			common.Error(newScene())
			common.Error(paintTitle(renderer, "Highscore!!", 10))

			common.Sleep(TitleDuration)
		}

		common.Error(newScene())
		common.Error(paintTitle(renderer, fmt.Sprintf("%d", points), 10))

		common.Sleep(TitleDuration)
	}()

	for running.IsSet() {
		var move func()

		select {
		case <-time.After(GameDelay):
			move = snake.LastMove

			snake.Hunger += GameDelay
			if snake.Hunger > DeadDelay {
				running.Unset()
				continue
			}
		case move = <-moves:
		}

		if move == nil {
			continue
		}

		if Collides(snake.Peek(move), ToPositions(stones)...) {
			snake.LastMove = nil

			continue
		}

		lastPos := snake.Position()

		move()
		snake.LastMove = move

		if Collides(snake.position, food.position) {
			snake.Tails = append(snake.Tails, NewTail(lastPos))

			snake.Hunger = 0
			snake.color = SnakeColor

			food.position = findFreePosition()

			stones = append(stones, NewStone(findFreePosition()))
		}

		if len(snake.Tails) > 1 && Collides(snake.position, ToPositions(snake.Tails[1:])...) {
			running.Unset()

			continue
		}

		if len(snake.Tails) > 0 {
			if len(snake.Tails) > 1 {
				copy(snake.Tails[1:], snake.Tails[0:len(snake.Tails)-1])
			}
			snake.Tails[0] = NewTail(snake.Position())
		}

		err = paintScene()
		if common.Error(err) {
			return err
		}
	}

	return nil
}

func main() {
	common.Run(nil)
}
