package update

import (
	"math"

	"github.com/ThatOtherAndrew/Hexecute/internal/models"
	"github.com/ThatOtherAndrew/Hexecute/pkg/wayland"
)

type App struct {
	app *models.App
}

func New(app *models.App) *App {
	return &App{app: app}
}

func (a *App) UpdateParticles(dt float32) {
	for i := 0; i < len(a.app.Particles); i++ {
		p := &a.app.Particles[i]
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.VY += 100 * dt
		p.Life -= dt

		if p.Life <= 0 {
			a.app.Particles[i] = a.app.Particles[len(a.app.Particles)-1]
			a.app.Particles = a.app.Particles[:len(a.app.Particles)-1]
			i--
		}
	}
}

func (a *App) UpdateCursor(window *wayland.WaylandWindow) {
	x, y := window.GetCursorPos()
	fx, fy := float32(x), float32(y)

	dx := fx - a.app.LastCursorX
	dy := fy - a.app.LastCursorY
	a.app.CursorVelocity = float32(math.Sqrt(float64(dx*dx + dy*dy)))

	velocityDiff := a.app.CursorVelocity - a.app.SmoothVelocity
	a.app.SmoothVelocity += velocityDiff * 0.2

	if a.app.CursorVelocity > 0.1 {
		targetRotation := float32(math.Atan2(float64(dy), float64(dx)))

		angleDiff := targetRotation - a.app.SmoothRotation
		if angleDiff > math.Pi {
			angleDiff -= 2 * math.Pi
		} else if angleDiff < -math.Pi {
			angleDiff += 2 * math.Pi
		}

		velocityFactor := float32(math.Min(float64(a.app.SmoothVelocity/5.0), 1.0))
		smoothFactor := 0.03 + velocityFactor*0.08
		a.app.SmoothRotation += angleDiff * smoothFactor
	}

	var targetDrawing float32
	if a.app.IsDrawing {
		targetDrawing = 1.0
	} else {
		targetDrawing = 0.0
	}

	diff := targetDrawing - a.app.SmoothDrawing
	a.app.SmoothDrawing += diff * 0.0375

	a.app.LastCursorX = fx
	a.app.LastCursorY = fy
}
