package spawn

import (
	"math"
	"math/rand/v2"

	"github.com/ThatOtherAndrew/Hexecute/internal/models"
)

type App struct {
	app *models.App
}

func New(app *models.App) *App {
	return &App{app: app}
}

func (a *App) SpawnCursorSparkles(x, y float32) {
	for range 3 {
		angle := rand.Float64() * 2 * math.Pi
		speed := rand.Float32()*80 + 40
		a.app.Particles = append(a.app.Particles, models.Particle{
			X:       x + (rand.Float32()-0.5)*8,
			Y:       y + (rand.Float32()-0.5)*8,
			VX:      float32(math.Cos(angle)) * speed,
			VY:      float32(math.Sin(angle))*speed - 30,
			Life:    0.8,
			MaxLife: 0.8,
			Size:    rand.Float32()*8 + 6,
			Hue:     rand.Float32(),
		})
	}
}

func (a *App) SpawnExitWisps(x, y float32) {
	for range 8 {
		angle := rand.Float64() * 2 * math.Pi
		speed := rand.Float32()*150 + 80
		a.app.Particles = append(a.app.Particles, models.Particle{
			X:       x + (rand.Float32()-0.5)*30,
			Y:       y + (rand.Float32()-0.5)*30,
			VX:      float32(math.Cos(angle)) * speed,
			VY:      float32(math.Sin(angle)) * speed,
			Life:    1.2,
			MaxLife: 1.2,
			Size:    rand.Float32()*12 + 8,
			Hue:     rand.Float32(),
		})
	}
}
