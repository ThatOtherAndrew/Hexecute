package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	points    []Point
	isDrawing bool
}

type Point struct {
	X, Y float32
}

func (g *Game) Update() error {
	isMouseDown := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	if isMouseDown {
		if !g.isDrawing {
			log.Println("Start drawing")
			g.isDrawing = true
		}
		x, y := ebiten.CursorPosition()
		g.points = append(g.points, Point{X: float32(x), Y: float32(y)})
	} else if g.isDrawing {
		log.Println("Stop drawing")
		g.isDrawing = false
		g.points = []Point{}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for i := 0; i < len(g.points)-1; i++ {
		vector.StrokeLine(
			screen,
			g.points[i].X,
			g.points[i].Y,
			g.points[i+1].X,
			g.points[i+1].Y,
			5,
			color.RGBA{255, 0, 255, 255},
			false,
		)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	monitorWidth, monitorHeight := ebiten.Monitor().Size()

	ebiten.SetWindowSize(monitorWidth, monitorHeight)
	ebiten.SetWindowTitle("hexecute")
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)

	game := &Game{
		points:    []Point{},
		isDrawing: false,
	}

	gameOptions := &ebiten.RunGameOptions{ScreenTransparent: true}

	if err := ebiten.RunGameWithOptions(game, gameOptions); err != nil {
		log.Fatal(err)
	}
}
