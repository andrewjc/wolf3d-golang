package game

import (
	"encoding/json"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
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

	timeBonusStartTick int64

	previousEucDistance           float64
	episodeCount                  int
	lastPlayer1PositionUpdateTick int64
	lastPlayer1Obs                []float64
	planPath                      []pixel.Vec
}

func (g *GameInstance) GameLoop() {

	g.gameInit()

	g.addGameObjects()

	g.Reset()

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

		//g.autoPlan(g.player1Controller, g.player2Controller, dt)

		// Render player1's view
		if g.renderListener.renderBuffer != nil {
			g.renderListener.renderBufferMutex.Lock()
			p := pixel.PictureDataFromImage(g.renderListener.renderBuffer)
			pixel.NewSprite(p, p.Bounds()).
				Draw(g.win, pixel.IM.Moved(g.center).Scaled(g.center, g.RenderScale))
			g.renderListener.renderBufferMutex.Unlock()
		} else {
			g.player1Controller.player.view.render()
		}

		g.win.Update()
	}
}

func (g *GameInstance) Reset() {

	g.episodeCount += 1
	print("Reset! Episode: ", g.episodeCount, "\n")

	rand.Seed(time.Now().UnixNano())

	g.episodeStartTick = time.Now().UnixMilli()
	g.previousEucDistance = 0
	g.lastPlayer1Obs = nil

	mapGen := Map{rows: 48, cols: 48}
	mapGen.GenerateMap()
	g.mapData = mapGen.mapData
	g.lights = mapGen.lights

	g.player1Controller.player.view.position = getRandomStartPosition(&g.mapData)
	g.player2Controller.player.view.position = getRandomStartPosition(&g.mapData)

	g.player1Controller.distanceStack = []float64{}
}

func (g *GameInstance) gameInit() {

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
		position:       pixel.V(0.0, 0.0),
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
		position:       pixel.V(0.0, 0.0),
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

	g.player1Controller.player.is_moving = true

	g.gameObjects = append(g.gameObjects,
		&player2,
	)
}

func (g *GameInstance) distToNearestWall(position pixel.Vec, threshold float32) float64 {
	mapData := g.mapData

	// Get the position of the player in the map
	mapX := int(position.X)
	mapY := int(position.Y)

	// Scan in concentric circles around the player, looking for a wall
	// if no wall is found, expand the circle by 1 cell
	for i := 0; i < 100; i++ {
		for y := -i; y <= i; y++ {
			for x := -i; x <= i; x++ {
				if mapData[mapX+x][mapY+y] > 0 {

					// Calculate the distance to the wall
					dist := math.Sqrt(math.Pow(float64(x), 2) + math.Pow(float64(y), 2))

					return dist
				}
			}
		}
	}

	return 9999
}

type ObsRecordSet struct {
	Obs1   []float64
	Obs2   []byte
	Reward float32
	Done   bool
	Action int
}

func (g *GameInstance) recordPlayerSet(action int, reward float32, done bool) {
	textObservations, imgObservations := g.GetPlayer1Observation()

	if textObservations[0] == 0 && textObservations[1] == 0 && textObservations[2] == 0 {
		return
	}

	recordSet := ObsRecordSet{
		Obs1:   textObservations,
		Obs2:   imgObservations,
		Reward: reward,
		Done:   done,
		Action: action,
	}

	// Implement the json serializer method
	b, err := json.Marshal(recordSet)

	// Open log file and write observation
	f, err := os.OpenFile("train_data.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if _, err := f.WriteString(string(b) + "\r\n"); err != nil {
		log.Fatal(err)
	}

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
