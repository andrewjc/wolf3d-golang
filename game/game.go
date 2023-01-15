package game

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"image"
	"image/color"
	"math/rand"
	"time"
)

type GameInstance struct {
	win         *pixelgl.Window
	cfg         pixelgl.WindowConfig
	mapData     [][]int
	lights      []LightSource
	textureData []byte
	textureMap  *image.RGBA
	normalMap   *image.RGBA
	dispMap     *image.RGBA

	center pixel.Vec

	gameObjects     []GameObject
	renderListener  *RenderListener
	renderListener2 *RenderListener

	// Runner ipc controllers
	player1Controller *PlayerController //runner
	player2Controller *EnemyController  //chaser

	// Params
	RenderWidth      int
	RenderHeight     int
	RenderScale      float64
	RenderFullscreen bool

	currentTick      int64
	episodeStartTick int64

	previousEucDistance float64
}

func (g *GameInstance) GameLoop() {

	rand.Seed(time.Now().UnixNano())

	g.episodeStartTick = time.Now().UnixMilli()
	g.previousEucDistance = 0

	g.gameInit()

	g.addGameObjects()

	last := time.Now()

	for !g.win.Closed() {
		if g.win.JustPressed(pixelgl.KeyEscape) || g.win.JustPressed(pixelgl.KeyQ) {
			return
		}

		g.win.Clear(color.Black)

		dt := time.Since(last).Seconds()
		last = time.Now()

		g.currentTick = last.UnixMilli()

		g.updateGameEntities(dt)

		// Process player input
		g.player1Controller.processInput(g.win, dt)

		// Render player1's view
		g.player1Controller.player.view.render()

		p := pixel.PictureDataFromImage(g.renderListener.renderBuffer)

		pixel.NewSprite(p, p.Bounds()).
			Draw(g.win, pixel.IM.Moved(g.center).Scaled(g.center, g.RenderScale))

		g.win.Update()
	}
}

func (g *GameInstance) gameInit() {

	mapGen := Map{rows: 48, cols: 48}
	mapGen.GenerateMap()
	g.mapData = mapGen.mapData
	g.lights = mapGen.lights

	g.textureMap = textureToImage("assets/texture.png")
	g.normalMap = textureToImage("assets/normal.png")
	g.dispMap = textureToImage("assets/disp.png")

	cfg := pixelgl.WindowConfig{
		Bounds:      pixel.R(0, 0, float64(g.RenderWidth)*g.RenderScale, float64(g.RenderHeight)*g.RenderScale),
		VSync:       true,
		Undecorated: false,
		Resizable:   true,
	}

	if g.RenderFullscreen {
		cfg.Monitor = pixelgl.PrimaryMonitor()
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	g.center = win.Bounds().Center()

	g.win = win
	g.cfg = cfg

	g.renderListener = &RenderListener{}
	g.renderListener2 = &RenderListener{}
}

func (g *GameInstance) updateGameEntities(timeDelta float64) {
	for _, e := range g.gameObjects {
		e.update(timeDelta)
	}
}

func (g *GameInstance) getTexNum(x, y int) int {
	return g.mapData[x][y]
}

func (g *GameInstance) addGameObjects() {

	player1Camera := RenderView{
		renderWidth:    640,
		renderHeight:   480,
		position:       getRandomStartPosition(&g.mapData),
		direction:      pixel.V(-1.0, 0.0),
		plane:          pixel.V(0.0, 0.66),
		renderListener: g.renderListener}

	g.player1Controller = &PlayerController{}

	player1 := Player{
		game:       g,
		view:       &player1Camera,
		controller: g.player1Controller,
	}

	g.player1Controller.player = &player1
	player1Camera.parent = &player1

	g.gameObjects = append(g.gameObjects,
		&player1,
	)

	player2Camera := RenderView{
		renderWidth:    320,
		renderHeight:   240,
		position:       getRandomStartPosition(&g.mapData),
		direction:      pixel.V(-1.0, 0.0),
		plane:          pixel.V(0.0, 0.66),
		renderListener: g.renderListener2}

	g.player2Controller = &EnemyController{}
	player2 := Enemy{
		game:       g,
		view:       &player2Camera,
		controller: g.player2Controller,
	}
	g.player2Controller.player = &player2

	player2Camera.parent = &player2

	g.gameObjects = append(g.gameObjects,
		&player2,
	)
}

func getRandomStartPosition(mapData *[][]int) pixel.Vec {
	var x, y int
	for {
		x = rand.Intn(len(*mapData))
		y = rand.Intn(len((*mapData)[0]))
		if (*mapData)[x][y] == 0 && emptyWithin(mapData, x, y, 2) {
			break
		}
	}

	return pixel.V(float64(x), float64(y))
}

func emptyWithin(data *[][]int, x int, y int, radius int) bool {
	maxX := len((*data)) - 1
	maxY := len((*data)[0]) - 1

	for i := -radius; i <= radius; i++ {
		for j := -radius; j <= radius; j++ {
			if (x+i) < 0 || (x+i) > maxX || (y+j) < 0 || (y+j) > maxY {
				continue
			}
			if (i*i+j*j) <= radius*radius && (*data)[x+i][y+j] != 0 {
				return false
			}
		}
	}
	return true
}
