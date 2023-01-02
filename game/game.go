package game

import (
    "bytes"
    "github.com/faiface/pixel"
    "github.com/faiface/pixel/pixelgl"
    "image"
    "image/color"
    "image/jpeg"
    "log"
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

    // Player ipc controllers
    player1Controller *PlayerController //runner
    player2Controller *PlayerController //chaser

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

    g.textureMap = LoadTextures()

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

    g.player1Controller = &PlayerController{}
    g.player2Controller = &PlayerController{}

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
        position:       pixel.V(5.0, 5.5),
        direction:      pixel.V(-1.0, 0.0),
        plane:          pixel.V(0.0, 0.66),
        renderListener: g.renderListener2}

    player2 := Player{
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

// Structs for reinforcement learning
// enum
type RLAction uint8

const (
    RLActionNone RLAction = iota
    RLActionMoveForward
    RLActionMoveBackward
    RLActionStrafeLeft
    RLActionStrafeRight
    RLActionTurnLeft
    RLActionTurnRight
)

type RLActionResult struct {
    Reward      float64
    Observation []uint8
    Done        bool
    Info        string
}

func (g *GameInstance) TakePlayer1Action(action_id RLAction) RLActionResult {

    if action_id == RLActionNone {
        p1Obs := g.GetPlayer1Observation()
        reward := g.player1Controller.player.getReward()
        return RLActionResult{Reward: reward, Observation: p1Obs, Done: false, Info: ""}
    } else if action_id == RLActionMoveForward {
        g.player1Controller.player.moveForward(0.1)
        p1Obs := g.GetPlayer1Observation()
        reward := g.player1Controller.player.getReward()
        return RLActionResult{Reward: reward, Observation: p1Obs, Done: false, Info: ""}
    } else if action_id == RLActionMoveBackward {
        g.player1Controller.player.moveBackwards(0.1)
        p1Obs := g.GetPlayer1Observation()
        reward := g.player1Controller.player.getReward()
        return RLActionResult{Reward: reward, Observation: p1Obs, Done: false, Info: ""}
    } else if action_id == RLActionStrafeLeft {
        g.player1Controller.player.moveLeft(0.1)
        p1Obs := g.GetPlayer1Observation()
        reward := g.player1Controller.player.getReward()
        return RLActionResult{Reward: reward, Observation: p1Obs, Done: false, Info: ""}
    } else if action_id == RLActionStrafeRight {
        g.player1Controller.player.moveRight(0.1)
        p1Obs := g.GetPlayer1Observation()
        reward := g.player1Controller.player.getReward()
        return RLActionResult{Reward: reward, Observation: p1Obs, Done: false, Info: ""}
    } else if action_id == RLActionTurnLeft {
        g.player1Controller.player.turnLeft(0.1)
        p1Obs := g.GetPlayer1Observation()
        reward := g.player1Controller.player.getReward()
        return RLActionResult{Reward: reward, Observation: p1Obs, Done: false, Info: ""}
    } else if action_id == RLActionTurnRight {
        g.player1Controller.player.turnRight(0.1)
        p1Obs := g.GetPlayer1Observation()
        reward := g.player1Controller.player.getReward()
        return RLActionResult{Reward: reward, Observation: p1Obs, Done: false, Info: ""}
    }

    log.Fatal("Unknown action type ", action_id)
    return RLActionResult{Reward: 0.0, Observation: nil, Done: false, Info: ""}
}

func (g *GameInstance) GetPlayer1Observation() []byte {
    // Compress the renderBuffer into JPEG
    var buf bytes.Buffer
    err := jpeg.Encode(&buf, g.renderListener.renderBuffer, &jpeg.Options{Quality: 30})
    if err != nil {
        log.Println("Error compressing observation: ", err)
        return nil
    } else {
        return buf.Bytes()
    }
}
