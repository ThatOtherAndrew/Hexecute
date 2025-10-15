package fornow

import (
	"log"
	"math"
	"math/rand/v2"
	"time"

	"github.com/ThatOtherAndrew/Hexecute/internal/execute"
	"github.com/ThatOtherAndrew/Hexecute/internal/models"
	"github.com/ThatOtherAndrew/Hexecute/internal/spawn"
	"github.com/ThatOtherAndrew/Hexecute/internal/stroke"
	"github.com/ThatOtherAndrew/Hexecute/pkg/wayland"
)

type App struct {
	app *models.App
}

func New(app *models.App) *App {
	return &App{app: app}
}

func (a *App) AddPoint(x, y float32) {
	newPoint := models.Point{X: x, Y: y, BornTime: time.Now()}

	shouldAdd := false
	if len(a.app.Points) == 0 {
		shouldAdd = true
	} else {
		lastPoint := a.app.Points[len(a.app.Points)-1]
		dx := newPoint.X - lastPoint.X
		dy := newPoint.Y - lastPoint.Y
		if dx*dx+dy*dy > 4 {
			shouldAdd = true

			for range 3 {
				angle := rand.Float64() * 2 * math.Pi
				speed := rand.Float32()*50 + 20
				a.app.Particles = append(a.app.Particles, models.Particle{
					X:       x + (rand.Float32()-0.5)*10,
					Y:       y + (rand.Float32()-0.5)*10,
					VX:      float32(math.Cos(angle)) * speed,
					VY:      float32(math.Sin(angle)) * speed,
					Life:    1.0,
					MaxLife: 1.0,
					Size:    rand.Float32()*15 + 10,
					Hue:     rand.Float32(),
				})
			}
		}
	}

	const MAX_POINTS = 2048

	if shouldAdd {
		a.app.Points = append(a.app.Points, newPoint)
		if len(a.app.Points) > MAX_POINTS {
			a.app.Points = a.app.Points[len(a.app.Points)-MAX_POINTS:]
		}
	}
}

func (a *App) RecognizeAndExecute(window *wayland.WaylandWindow, x, y float32) {
	if len(a.app.Points) < 5 {
		log.Println("Gesture too short, ignoring")
		return
	}

	processed := stroke.ProcessStroke(a.app.Points)

	bestMatch := -1
	bestScore := 0.0

	for i, gesture := range a.app.SavedGestures {
		match, score := stroke.UnistrokeRecognise(processed, gesture.Templates)
		log.Printf("Gesture %d (%s): template %d, score %.3f", i, gesture.Command, match, score)

		if score > bestScore {
			bestScore = score
			bestMatch = i
		}
	}

	if bestMatch >= 0 && bestScore > 0.6 {
		command := a.app.SavedGestures[bestMatch].Command
		log.Printf("Matched gesture: %s (score: %.3f)", command, bestScore)

		if err := execute.Command(command); err != nil {
			log.Printf("Failed to execute command: %v", err)
		} else {
			log.Printf("Executed: %s", command)
		}

		a.app.IsExiting = true
		a.app.ExitStartTime = time.Now()
		window.DisableInput()
		spawn := spawn.New(a.app)
		spawn.SpawnExitWisps(x, y)
	} else {
		log.Printf("No confident match (best score: %.3f)", bestScore)
	}
}
