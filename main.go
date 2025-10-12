package main

import (
	"image/color"
	"log"
	"os"
	"os/exec"

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
		newPoint := Point{X: float32(x), Y: float32(y)}

		// Only add if cursor moved enough (> 2 pixels) or it's the first point
		shouldAdd := false
		if len(g.points) == 0 {
			shouldAdd = true
		} else {
			lastPoint := g.points[len(g.points)-1]
			dx := newPoint.X - lastPoint.X
			dy := newPoint.Y - lastPoint.Y
			if dx*dx+dy*dy > 4 { // distance > 2 pixels
				shouldAdd = true
			}
		}

		if shouldAdd {
			g.points = append(g.points, newPoint)
			if len(g.points) > 512 {
				g.points = g.points[len(g.points)-512:]
			}
		}
	} else if g.isDrawing {
		log.Println("Stop drawing")
		g.isDrawing = false
		g.points = []Point{}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if len(g.points) == 0 {
		return
	}

	var path vector.Path
	path.MoveTo(g.points[0].X, g.points[0].Y)
	for i := 1; i < len(g.points); i++ {
		path.LineTo(g.points[i].X, g.points[i].Y)
	}

	strokeOptions := &vector.StrokeOptions{
		Width:    5,
		LineCap:  vector.LineCapRound,
		LineJoin: vector.LineJoinRound,
	}
	drawOptions := &vector.DrawPathOptions{
		// AntiAlias: true,
	}
	drawOptions.ColorScale.ScaleWithColor(color.RGBA{255, 0, 255, 255})

	vector.StrokePath(screen, &path, strokeOptions, drawOptions)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	// make hyprland play nice
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		exec.Command("hyprctl", "keyword", "windowrulev2", "center,pin,noborder,noanim,noblur,title:^(hexecute)$").Run()
	}

	monitorWidth, monitorHeight := ebiten.Monitor().Size()

	ebiten.SetWindowSize(monitorWidth, monitorHeight)
	ebiten.SetWindowTitle("hexecute")
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	ebiten.SetTPS(ebiten.SyncWithFPS)

	game := &Game{
		points:    []Point{},
		isDrawing: false,
	}

	gameOptions := &ebiten.RunGameOptions{ScreenTransparent: true}

	if err := ebiten.RunGameWithOptions(game, gameOptions); err != nil {
		log.Fatal(err)
	}
}
