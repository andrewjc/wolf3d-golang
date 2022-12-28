package game

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"image"
	"image/color"
	"time"
)

type GameInstance struct {
	win         *pixelgl.Window
	cfg         pixelgl.WindowConfig
	mapData     [][]int
	textureData []byte
	textureMap  *image.RGBA

	center pixel.Vec

	gameObjects     []GameObject
	renderListener  *RenderListener
	renderListener2 *RenderListener

	// Params
	RenderWidth      int
	RenderHeight     int
	RenderScale      float64
	RenderFullscreen bool
}

func (g *GameInstance) GameLoop() {

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

		g.updateGameEntities(dt)

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

	// Print out the mapData to the console
	for _, row := range g.mapData {
		for _, col := range row {
			fmt.Printf("%d", col)
		}
		fmt.Println()
	}

	g.textureMap = LoadTextures()

	cfg := pixelgl.WindowConfig{
		Bounds:      pixel.R(0, 0, float64(g.RenderWidth)*g.RenderScale, float64(g.RenderHeight)*g.RenderScale),
		VSync:       true,
		Undecorated: true,
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
		e.processInput(g.win, timeDelta)
		e.update(timeDelta)
		e.render(timeDelta)
	}
}

func (g *GameInstance) getTexNum(x, y int) int {
	return g.mapData[x][y]
}

func (g *GameInstance) addGameObjects() {

	player1Camera := RenderView{
		renderWidth:    640,
		renderHeight:   480,
		position:       pixel.V(12.0, 14.5),
		direction:      pixel.V(-1.0, 0.0),
		plane:          pixel.V(0.0, 0.66),
		renderListener: g.renderListener}

	player1 := Player{
		game:       g,
		view:       &player1Camera,
		controller: &PlayerController{},
	}

	player1Camera.parent = &player1

	g.gameObjects = append(g.gameObjects,
		&player1,
	)

	player2Camera := RenderView{
		renderWidth:    320,
		renderHeight:   240,
		position:       pixel.V(5.0, 5.5),
		direction:      pixel.V(-1.0, 0.0),
		plane:          pixel.V(0.0, 0.66),
		renderListener: g.renderListener2}

	player2 := Player{
		game:       g,
		view:       &player2Camera,
		controller: &PlayerController{},
	}

	player2Camera.parent = &player2

	g.gameObjects = append(g.gameObjects,
		&player2,
	)
}