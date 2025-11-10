package draw

import (
	"math"
	"time"

	"github.com/ThatOtherAndrew/Hexecute/internal/models"
	"github.com/ThatOtherAndrew/Hexecute/pkg/wayland"
	"github.com/go-gl/gl/v4.1-core/gl"
)

type App struct {
	app *models.App
}

func New(app *models.App) *App {
	return &App{app: app}
}

func (a *App) Draw(window *wayland.WaylandWindow) {
	gl.Clear(gl.COLOR_BUFFER_BIT)

	currentTime := float32(time.Since(a.app.StartTime).Seconds())

	a.drawBackground(currentTime, window)

	x, y := window.GetCursorPos()
	a.drawCursorGlow(window, float32(x), float32(y), currentTime)

	for pass := range 3 {
		thickness := float32(7 + pass*4)
		alpha := float32(0.7 - float32(pass)*0.15)
		a.drawLine(window, thickness, alpha, currentTime)
	}

	a.drawParticles(window)
}

func (a *App) drawLine(
	window *wayland.WaylandWindow,
	baseThickness, baseAlpha, currentTime float32,
) {
	if len(a.app.Points) < 2 {
		return
	}

	vertices := make([]float32, 0, len(a.app.Points)*10)

	for i := range a.app.Points {
		age := float32(time.Since(a.app.Points[i].BornTime).Seconds())
		fade := 1.0 - (age / 1.5)
		if fade < 0 {
			fade = 0
		}
		alpha := fade * baseAlpha

		var perpX, perpY float32

		if i == 0 {
			dx := a.app.Points[i+1].X - a.app.Points[i].X
			dy := a.app.Points[i+1].Y - a.app.Points[i].Y
			length := float32(1.0) / float32(math.Sqrt(float64(dx*dx+dy*dy)))
			perpX = -dy * length
			perpY = dx * length
		} else if i == len(a.app.Points)-1 {
			dx := a.app.Points[i].X - a.app.Points[i-1].X
			dy := a.app.Points[i].Y - a.app.Points[i-1].Y
			length := float32(1.0) / float32(math.Sqrt(float64(dx*dx+dy*dy)))
			perpX = -dy * length
			perpY = dx * length
		} else {
			dx1 := a.app.Points[i].X - a.app.Points[i-1].X
			dy1 := a.app.Points[i].Y - a.app.Points[i-1].Y
			len1 := float32(math.Sqrt(float64(dx1*dx1 + dy1*dy1)))
			if len1 > 0 {
				dx1 /= len1
				dy1 /= len1
			}

			dx2 := a.app.Points[i+1].X - a.app.Points[i].X
			dy2 := a.app.Points[i+1].Y - a.app.Points[i].Y
			len2 := float32(math.Sqrt(float64(dx2*dx2 + dy2*dy2)))
			if len2 > 0 {
				dx2 /= len2
				dy2 /= len2
			}

			avgDx := (dx1 + dx2) * 0.5
			avgDy := (dy1 + dy2) * 0.5
			avgLen := float32(math.Sqrt(float64(avgDx*avgDx + avgDy*avgDy)))
			if avgLen > 0 {
				avgDx /= avgLen
				avgDy /= avgLen
			}

			perpX = -avgDy
			perpY = avgDx
		}

		vertices = append(vertices, a.app.Points[i].X, a.app.Points[i].Y, perpX, perpY, alpha)
		vertices = append(vertices, a.app.Points[i].X, a.app.Points[i].Y, -perpX, -perpY, alpha)
	}

	cutoff := time.Now().Add(-1500 * time.Millisecond)
	for len(a.app.Points) > 0 && a.app.Points[0].BornTime.Before(cutoff) {
		a.app.Points = a.app.Points[1:]
	}

	if len(vertices) == 0 {
		return
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, a.app.Vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.DYNAMIC_DRAW)

	width, height := window.GetSize()

	gl.UseProgram(a.app.Program)
	resolutionLoc := gl.GetUniformLocation(a.app.Program, gl.Str("resolution\x00"))
	gl.Uniform2f(resolutionLoc, float32(width), float32(height))
	thicknessLoc := gl.GetUniformLocation(a.app.Program, gl.Str("thickness\x00"))
	gl.Uniform1f(thicknessLoc, baseThickness)
	timeLoc := gl.GetUniformLocation(a.app.Program, gl.Str("time\x00"))
	gl.Uniform1f(timeLoc, currentTime)

	gl.BindVertexArray(a.app.Vao)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(a.app.Points)*2))
	gl.BindVertexArray(0)
}

func (a *App) drawParticles(window *wayland.WaylandWindow) {
	if len(a.app.Particles) == 0 {
		return
	}

	vertices := make([]float32, 0, len(a.app.Particles)*6)
	for _, p := range a.app.Particles {
		vertices = append(vertices, p.X, p.Y, p.Life, p.MaxLife, p.Size, p.Hue)
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, a.app.ParticleVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.DYNAMIC_DRAW)

	width, height := window.GetSize()

	gl.UseProgram(a.app.ParticleProgram)
	resolutionLoc := gl.GetUniformLocation(a.app.ParticleProgram, gl.Str("resolution\x00"))
	gl.Uniform2f(resolutionLoc, float32(width), float32(height))

	gl.BindVertexArray(a.app.ParticleVAO)
	gl.DrawArrays(gl.POINTS, 0, int32(len(a.app.Particles)))
	gl.BindVertexArray(0)
}

func (a *App) drawBackground(currentTime float32, window *wayland.WaylandWindow) {
	fadeDuration := float32(1.0)
	targetAlpha := a.app.Settings.OverlayAlpha

	var alpha float32
	if currentTime < fadeDuration {
		progress := currentTime / fadeDuration
		easedProgress := 1.0 - (1.0-progress)*(1.0-progress)*(1.0-progress)*(1.0-progress)*(1.0-progress)
		alpha = easedProgress * targetAlpha
	} else {
		alpha = targetAlpha
	}

	if a.app.IsExiting {
		exitDuration := float32(0.8)
		elapsed := float32(time.Since(a.app.ExitStartTime).Seconds())
		if elapsed < exitDuration {
			progress := elapsed / exitDuration
			easedProgress := 1.0 - (1.0-progress)*(1.0-progress)*(1.0-progress)*(1.0-progress)*(1.0-progress)
			alpha *= (1.0 - easedProgress)
		} else {
			alpha = 0
		}
	}

	x, y := window.GetCursorPos()
	width, height := window.GetSize()

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.UseProgram(a.app.BgProgram)

	alphaLoc := gl.GetUniformLocation(a.app.BgProgram, gl.Str("alpha\x00"))
	gl.Uniform1f(alphaLoc, alpha)

	cursorPosLoc := gl.GetUniformLocation(a.app.BgProgram, gl.Str("cursorPos\x00"))
	gl.Uniform2f(cursorPosLoc, float32(x), float32(float64(height)-y))

	resolutionLoc := gl.GetUniformLocation(a.app.BgProgram, gl.Str("resolution\x00"))
	gl.Uniform2f(resolutionLoc, float32(width), float32(height))

	gl.BindVertexArray(a.app.BgVAO)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	gl.BindVertexArray(0)

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
}

func (a *App) drawCursorGlow(window *wayland.WaylandWindow, cursorX, cursorY, currentTime float32) {
	width, height := window.GetSize()

	growDuration := float32(1.2)
	var scale float32
	if currentTime < growDuration {
		t := currentTime / growDuration
		c4 := (2.0 * math.Pi) / 3.0
		if t == 0 {
			scale = 0
		} else if t >= 1 {
			scale = 1
		} else {
			scale = float32(math.Pow(2, -10*float64(t))*math.Sin((float64(t)*10-0.75)*c4) + 1)
		}
	} else {
		scale = 1.0
	}

	var exitProgress float32
	if a.app.IsExiting {
		exitDuration := float32(0.8)
		elapsed := float32(time.Since(a.app.ExitStartTime).Seconds())
		if elapsed < exitDuration {
			t := elapsed / exitDuration
			exitProgress = t * t * t
			scale *= (1.0 - exitProgress)
		} else {
			exitProgress = 1.0
			scale = 0
		}
	}

	gl.UseProgram(a.app.CursorGlowProgram)

	cursorPosLoc := gl.GetUniformLocation(a.app.CursorGlowProgram, gl.Str("cursorPos\x00"))
	gl.Uniform2f(cursorPosLoc, cursorX, cursorY)

	resolutionLoc := gl.GetUniformLocation(a.app.CursorGlowProgram, gl.Str("resolution\x00"))
	gl.Uniform2f(resolutionLoc, float32(width), float32(height))

	glowSizeLoc := gl.GetUniformLocation(a.app.CursorGlowProgram, gl.Str("glowSize\x00"))
	gl.Uniform1f(glowSizeLoc, 80.0*scale)

	timeLoc := gl.GetUniformLocation(a.app.CursorGlowProgram, gl.Str("time\x00"))
	gl.Uniform1f(timeLoc, currentTime)

	velocityLoc := gl.GetUniformLocation(a.app.CursorGlowProgram, gl.Str("velocity\x00"))
	gl.Uniform1f(velocityLoc, a.app.SmoothVelocity)

	rotationLoc := gl.GetUniformLocation(a.app.CursorGlowProgram, gl.Str("rotation\x00"))
	gl.Uniform1f(rotationLoc, a.app.SmoothRotation)

	isDrawingLoc := gl.GetUniformLocation(a.app.CursorGlowProgram, gl.Str("isDrawing\x00"))
	gl.Uniform1f(isDrawingLoc, a.app.SmoothDrawing)

	exitProgressLoc := gl.GetUniformLocation(a.app.CursorGlowProgram, gl.Str("exitProgress\x00"))
	gl.Uniform1f(exitProgressLoc, exitProgress)

	gl.BindVertexArray(a.app.CursorGlowVAO)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	gl.BindVertexArray(0)
}
