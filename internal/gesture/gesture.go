package gestures

import (
	"encoding/json"
	"math"
	"math/rand/v2"
	"os"
	"time"

	"github.com/ThatOtherAndrew/Hexecute/internal/config"
	"github.com/ThatOtherAndrew/Hexecute/internal/models"
)

type App struct {
	app *models.App
}

func New(app *models.App) *App {
	return &App{app: app}
}

func LoadGestures() ([]models.GestureConfig, error) {
	configFile, err := config.GetPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.GestureConfig{}, nil
		}
		return nil, err
	}

	var gestures []models.GestureConfig
	if err := json.Unmarshal(data, &gestures); err != nil {
		return nil, err
	}

	return gestures, nil
}

func SaveGesture(command string, templates [][]models.Point) error {
	configFile, err := config.GetPath()
	if err != nil {
		return err
	}

	var gestures []models.GestureConfig
	if data, err := os.ReadFile(configFile); err == nil {
		if err := json.Unmarshal(data, &gestures); err != nil {
			return err
		}
	}

	newGesture := models.GestureConfig{
		Command:   command,
		Templates: templates,
	}

	found := false
	for i, g := range gestures {
		if g.Command == command {
			gestures[i] = newGesture
			found = true
			break
		}
	}
	if !found {
		gestures = append(gestures, newGesture)
	}

	data, err := json.Marshal(gestures)
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
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
