package main

import (
	"encoding/json"
	"flag"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	"github.com/ThatOtherAndrew/Hexecute/internal/config"
	"github.com/ThatOtherAndrew/Hexecute/internal/models"
	"github.com/ThatOtherAndrew/Hexecute/internal/shaders"
	"github.com/ThatOtherAndrew/Hexecute/internal/stroke"
	"github.com/ThatOtherAndrew/Hexecute/pkg/wayland"
	"github.com/go-gl/gl/v4.1-core/gl"
)

func init() {
	runtime.LockOSThread()
}

type App struct {
	points            []models.Point
	particles         []models.Particle
	isDrawing         bool
	vao               uint32
	vbo               uint32
	program           uint32
	particleVAO       uint32
	particleVBO       uint32
	particleProgram   uint32
	bgVAO             uint32
	bgVBO             uint32
	bgProgram         uint32
	cursorGlowVAO     uint32
	cursorGlowVBO     uint32
	cursorGlowProgram uint32
	startTime         time.Time
	lastCursorX       float32
	lastCursorY       float32
	cursorVelocity    float32
	smoothVelocity    float32
	smoothRotation    float32
	smoothDrawing     float32
	isExiting         bool
	exitStartTime     time.Time
	learnMode         bool
	learnCommand      string
	learnGestures     [][]models.Point
	learnCount        int
	savedGestures     []models.GestureConfig
}

func (a *App) initGL() error {
	if err := gl.Init(); err != nil {
		return err
	}

	vertShader, err := shaders.CompileShaderFromFile(shaders.LineVertexPath, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}
	fragShader, err := shaders.CompileShaderFromFile(shaders.LineFragmentPath, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	a.program = gl.CreateProgram()
	gl.AttachShader(a.program, vertShader)
	gl.AttachShader(a.program, fragShader)
	gl.LinkProgram(a.program)

	var status int32
	gl.GetProgramiv(a.program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(a.program, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetProgramInfoLog(a.program, logLength, nil, &logMsg[0])
		log.Fatalf("Failed to link program: %s", logMsg)
	}

	gl.DeleteShader(vertShader)
	gl.DeleteShader(fragShader)

	particleVertShader, err := shaders.CompileShaderFromFile(
		shaders.ParticleVertexPath,
		gl.VERTEX_SHADER,
	)
	if err != nil {
		return err
	}
	particleFragShader, err := shaders.CompileShaderFromFile(
		shaders.ParticleFragmentPath,
		gl.FRAGMENT_SHADER,
	)
	if err != nil {
		return err
	}

	a.particleProgram = gl.CreateProgram()
	gl.AttachShader(a.particleProgram, particleVertShader)
	gl.AttachShader(a.particleProgram, particleFragShader)
	gl.LinkProgram(a.particleProgram)

	gl.GetProgramiv(a.particleProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(a.particleProgram, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetProgramInfoLog(a.particleProgram, logLength, nil, &logMsg[0])
		log.Fatalf("Failed to link particle program: %s", logMsg)
	}

	gl.DeleteShader(particleVertShader)
	gl.DeleteShader(particleFragShader)

	gl.GenVertexArrays(1, &a.vao)
	gl.GenBuffers(1, &a.vbo)

	gl.BindVertexArray(a.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, a.vbo)

	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 5*4, nil)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(2, 1, gl.FLOAT, false, 5*4, gl.PtrOffset(4*4))
	gl.EnableVertexAttribArray(2)

	gl.BindVertexArray(0)

	gl.GenVertexArrays(1, &a.particleVAO)
	gl.GenBuffers(1, &a.particleVBO)

	gl.BindVertexArray(a.particleVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, a.particleVBO)

	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 6*4, nil)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 1, gl.FLOAT, false, 6*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(2, 1, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(3, 1, gl.FLOAT, false, 6*4, gl.PtrOffset(4*4))
	gl.EnableVertexAttribArray(3)
	gl.VertexAttribPointer(4, 1, gl.FLOAT, false, 6*4, gl.PtrOffset(5*4))
	gl.EnableVertexAttribArray(4)

	gl.BindVertexArray(0)

	bgVertShader, err := shaders.CompileShaderFromFile(
		shaders.BackgroundVertexPath,
		gl.VERTEX_SHADER,
	)
	if err != nil {
		return err
	}
	bgFragShader, err := shaders.CompileShaderFromFile(
		shaders.BackgroundFragmentPath,
		gl.FRAGMENT_SHADER,
	)
	if err != nil {
		return err
	}

	a.bgProgram = gl.CreateProgram()
	gl.AttachShader(a.bgProgram, bgVertShader)
	gl.AttachShader(a.bgProgram, bgFragShader)
	gl.LinkProgram(a.bgProgram)

	gl.GetProgramiv(a.bgProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(a.bgProgram, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetProgramInfoLog(a.bgProgram, logLength, nil, &logMsg[0])
		log.Fatalf("Failed to link background program: %s", logMsg)
	}

	gl.DeleteShader(bgVertShader)
	gl.DeleteShader(bgFragShader)

	gl.GenVertexArrays(1, &a.bgVAO)
	gl.GenBuffers(1, &a.bgVBO)

	gl.BindVertexArray(a.bgVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, a.bgVBO)

	quadVertices := []float32{
		-1.0, -1.0,
		1.0, -1.0,
		-1.0, 1.0,
		1.0, 1.0,
	}
	gl.BufferData(gl.ARRAY_BUFFER, len(quadVertices)*4, gl.Ptr(quadVertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 2*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.BindVertexArray(0)

	cursorGlowVertShader, err := shaders.CompileShaderFromFile(
		shaders.CursorGlowVertexPath,
		gl.VERTEX_SHADER,
	)
	if err != nil {
		return err
	}
	cursorGlowFragShader, err := shaders.CompileShaderFromFile(
		shaders.CursorGlowFragmentPath,
		gl.FRAGMENT_SHADER,
	)
	if err != nil {
		return err
	}

	a.cursorGlowProgram = gl.CreateProgram()
	gl.AttachShader(a.cursorGlowProgram, cursorGlowVertShader)
	gl.AttachShader(a.cursorGlowProgram, cursorGlowFragShader)
	gl.LinkProgram(a.cursorGlowProgram)

	gl.GetProgramiv(a.cursorGlowProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(a.cursorGlowProgram, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetProgramInfoLog(a.cursorGlowProgram, logLength, nil, &logMsg[0])
		log.Fatalf("Failed to link cursor glow program: %s", logMsg)
	}

	gl.DeleteShader(cursorGlowVertShader)
	gl.DeleteShader(cursorGlowFragShader)

	gl.GenVertexArrays(1, &a.cursorGlowVAO)
	gl.GenBuffers(1, &a.cursorGlowVBO)

	gl.BindVertexArray(a.cursorGlowVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, a.cursorGlowVBO)

	glowQuadVertices := []float32{
		-1.0, -1.0,
		1.0, -1.0,
		-1.0, 1.0,
		1.0, 1.0,
	}
	gl.BufferData(
		gl.ARRAY_BUFFER,
		len(glowQuadVertices)*4,
		gl.Ptr(glowQuadVertices),
		gl.STATIC_DRAW,
	)

	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 2*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.BindVertexArray(0)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
	gl.Enable(gl.PROGRAM_POINT_SIZE)

	return nil
}

func (a *App) addPoint(x, y float32) {
	newPoint := models.Point{X: x, Y: y, BornTime: time.Now()}

	shouldAdd := false
	if len(a.points) == 0 {
		shouldAdd = true
	} else {
		lastPoint := a.points[len(a.points)-1]
		dx := newPoint.X - lastPoint.X
		dy := newPoint.Y - lastPoint.Y
		if dx*dx+dy*dy > 4 {
			shouldAdd = true

			for range 3 {
				angle := rand.Float64() * 2 * math.Pi
				speed := rand.Float32()*50 + 20
				a.particles = append(a.particles, models.Particle{
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
		a.points = append(a.points, newPoint)
		if len(a.points) > MAX_POINTS {
			a.points = a.points[len(a.points)-MAX_POINTS:]
		}
	}
}

func (a *App) spawnCursorSparkles(x, y float32) {
	for range 3 {
		angle := rand.Float64() * 2 * math.Pi
		speed := rand.Float32()*80 + 40
		a.particles = append(a.particles, models.Particle{
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

func (a *App) spawnExitWisps(x, y float32) {
	for range 8 {
		angle := rand.Float64() * 2 * math.Pi
		speed := rand.Float32()*150 + 80
		a.particles = append(a.particles, models.Particle{
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

func (a *App) updateParticles(dt float32) {
	for i := 0; i < len(a.particles); i++ {
		p := &a.particles[i]
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.VY += 100 * dt
		p.Life -= dt

		if p.Life <= 0 {
			a.particles[i] = a.particles[len(a.particles)-1]
			a.particles = a.particles[:len(a.particles)-1]
			i--
		}
	}
}

func (a *App) updateCursor(window *wayland.WaylandWindow) {
	x, y := window.GetCursorPos()
	fx, fy := float32(x), float32(y)

	dx := fx - a.lastCursorX
	dy := fy - a.lastCursorY
	a.cursorVelocity = float32(math.Sqrt(float64(dx*dx + dy*dy)))

	velocityDiff := a.cursorVelocity - a.smoothVelocity
	a.smoothVelocity += velocityDiff * 0.2

	if a.cursorVelocity > 0.1 {
		targetRotation := float32(math.Atan2(float64(dy), float64(dx)))

		angleDiff := targetRotation - a.smoothRotation
		if angleDiff > math.Pi {
			angleDiff -= 2 * math.Pi
		} else if angleDiff < -math.Pi {
			angleDiff += 2 * math.Pi
		}

		velocityFactor := float32(math.Min(float64(a.smoothVelocity/5.0), 1.0))
		smoothFactor := 0.03 + velocityFactor*0.08
		a.smoothRotation += angleDiff * smoothFactor
	}

	var targetDrawing float32
	if a.isDrawing {
		targetDrawing = 1.0
	} else {
		targetDrawing = 0.0
	}

	diff := targetDrawing - a.smoothDrawing
	a.smoothDrawing += diff * 0.0375

	a.lastCursorX = fx
	a.lastCursorY = fy
}

func (a *App) draw(window *wayland.WaylandWindow) {
	gl.Clear(gl.COLOR_BUFFER_BIT)

	currentTime := float32(time.Since(a.startTime).Seconds())

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
	if len(a.points) < 2 {
		return
	}

	vertices := make([]float32, 0, len(a.points)*10)

	for i := range a.points {
		age := float32(time.Since(a.points[i].BornTime).Seconds())
		fade := 1.0 - (age / 1.5)
		if fade < 0 {
			fade = 0
		}
		alpha := fade * baseAlpha

		var perpX, perpY float32

		if i == 0 {
			dx := a.points[i+1].X - a.points[i].X
			dy := a.points[i+1].Y - a.points[i].Y
			length := float32(1.0) / float32(math.Sqrt(float64(dx*dx+dy*dy)))
			perpX = -dy * length
			perpY = dx * length
		} else if i == len(a.points)-1 {
			dx := a.points[i].X - a.points[i-1].X
			dy := a.points[i].Y - a.points[i-1].Y
			length := float32(1.0) / float32(math.Sqrt(float64(dx*dx+dy*dy)))
			perpX = -dy * length
			perpY = dx * length
		} else {
			dx1 := a.points[i].X - a.points[i-1].X
			dy1 := a.points[i].Y - a.points[i-1].Y
			len1 := float32(math.Sqrt(float64(dx1*dx1 + dy1*dy1)))
			if len1 > 0 {
				dx1 /= len1
				dy1 /= len1
			}

			dx2 := a.points[i+1].X - a.points[i].X
			dy2 := a.points[i+1].Y - a.points[i].Y
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

		vertices = append(vertices, a.points[i].X, a.points[i].Y, perpX, perpY, alpha)
		vertices = append(vertices, a.points[i].X, a.points[i].Y, -perpX, -perpY, alpha)
	}

	cutoff := time.Now().Add(-1500 * time.Millisecond)
	for len(a.points) > 0 && a.points[0].BornTime.Before(cutoff) {
		a.points = a.points[1:]
	}

	if len(vertices) == 0 {
		return
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, a.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.DYNAMIC_DRAW)

	width, height := window.GetSize()

	gl.UseProgram(a.program)
	resolutionLoc := gl.GetUniformLocation(a.program, gl.Str("resolution\x00"))
	gl.Uniform2f(resolutionLoc, float32(width), float32(height))
	thicknessLoc := gl.GetUniformLocation(a.program, gl.Str("thickness\x00"))
	gl.Uniform1f(thicknessLoc, baseThickness)
	timeLoc := gl.GetUniformLocation(a.program, gl.Str("time\x00"))
	gl.Uniform1f(timeLoc, currentTime)

	gl.BindVertexArray(a.vao)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(a.points)*2))
	gl.BindVertexArray(0)
}

func (a *App) drawParticles(window *wayland.WaylandWindow) {
	if len(a.particles) == 0 {
		return
	}

	vertices := make([]float32, 0, len(a.particles)*6)
	for _, p := range a.particles {
		vertices = append(vertices, p.X, p.Y, p.Life, p.MaxLife, p.Size, p.Hue)
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, a.particleVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.DYNAMIC_DRAW)

	width, height := window.GetSize()

	gl.UseProgram(a.particleProgram)
	resolutionLoc := gl.GetUniformLocation(a.particleProgram, gl.Str("resolution\x00"))
	gl.Uniform2f(resolutionLoc, float32(width), float32(height))

	gl.BindVertexArray(a.particleVAO)
	gl.DrawArrays(gl.POINTS, 0, int32(len(a.particles)))
	gl.BindVertexArray(0)
}

func (a *App) drawBackground(currentTime float32, window *wayland.WaylandWindow) {
	fadeDuration := float32(1.0)
	targetAlpha := float32(0.75)

	var alpha float32
	if currentTime < fadeDuration {
		progress := currentTime / fadeDuration
		easedProgress := 1.0 - (1.0-progress)*(1.0-progress)*(1.0-progress)*(1.0-progress)*(1.0-progress)
		alpha = easedProgress * targetAlpha
	} else {
		alpha = targetAlpha
	}

	if a.isExiting {
		exitDuration := float32(0.8)
		elapsed := float32(time.Since(a.exitStartTime).Seconds())
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

	gl.UseProgram(a.bgProgram)

	alphaLoc := gl.GetUniformLocation(a.bgProgram, gl.Str("alpha\x00"))
	gl.Uniform1f(alphaLoc, alpha)

	cursorPosLoc := gl.GetUniformLocation(a.bgProgram, gl.Str("cursorPos\x00"))
	gl.Uniform2f(cursorPosLoc, float32(x), float32(float64(height)-y))

	resolutionLoc := gl.GetUniformLocation(a.bgProgram, gl.Str("resolution\x00"))
	gl.Uniform2f(resolutionLoc, float32(width), float32(height))

	gl.BindVertexArray(a.bgVAO)
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
	if a.isExiting {
		exitDuration := float32(0.8)
		elapsed := float32(time.Since(a.exitStartTime).Seconds())
		if elapsed < exitDuration {
			t := elapsed / exitDuration
			exitProgress = t * t * t
			scale *= (1.0 - exitProgress)
		} else {
			exitProgress = 1.0
			scale = 0
		}
	}

	gl.UseProgram(a.cursorGlowProgram)

	cursorPosLoc := gl.GetUniformLocation(a.cursorGlowProgram, gl.Str("cursorPos\x00"))
	gl.Uniform2f(cursorPosLoc, cursorX, cursorY)

	resolutionLoc := gl.GetUniformLocation(a.cursorGlowProgram, gl.Str("resolution\x00"))
	gl.Uniform2f(resolutionLoc, float32(width), float32(height))

	glowSizeLoc := gl.GetUniformLocation(a.cursorGlowProgram, gl.Str("glowSize\x00"))
	gl.Uniform1f(glowSizeLoc, 80.0*scale)

	timeLoc := gl.GetUniformLocation(a.cursorGlowProgram, gl.Str("time\x00"))
	gl.Uniform1f(timeLoc, currentTime)

	velocityLoc := gl.GetUniformLocation(a.cursorGlowProgram, gl.Str("velocity\x00"))
	gl.Uniform1f(velocityLoc, a.smoothVelocity)

	rotationLoc := gl.GetUniformLocation(a.cursorGlowProgram, gl.Str("rotation\x00"))
	gl.Uniform1f(rotationLoc, a.smoothRotation)

	isDrawingLoc := gl.GetUniformLocation(a.cursorGlowProgram, gl.Str("isDrawing\x00"))
	gl.Uniform1f(isDrawingLoc, a.smoothDrawing)

	exitProgressLoc := gl.GetUniformLocation(a.cursorGlowProgram, gl.Str("exitProgress\x00"))
	gl.Uniform1f(exitProgressLoc, exitProgress)

	gl.BindVertexArray(a.cursorGlowVAO)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	gl.BindVertexArray(0)
}

func loadGestures() ([]models.GestureConfig, error) {
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

func saveGesture(command string, templates [][]models.Point) error {
	configFile, err := config.GetPath()
	if err != nil {
		return err
	}

	var gestures []models.GestureConfig
	if data, err := os.ReadFile(configFile); err == nil {
		json.Unmarshal(data, &gestures)
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

func executeCommand(command string) error {
	if command == "" {
		return nil
	}

	cmd := exec.Command("sh", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	return cmd.Start()
}

func (a *App) recognizeAndExecute(window *wayland.WaylandWindow, x, y float32) {
	if len(a.points) < 5 {
		log.Println("Gesture too short, ignoring")
		return
	}

	processed := stroke.ProcessStroke(a.points)

	bestMatch := -1
	bestScore := 0.0

	for i, gesture := range a.savedGestures {
		match, score := stroke.UnistrokeRecognise(processed, gesture.Templates)
		log.Printf("Gesture %d (%s): template %d, score %.3f", i, gesture.Command, match, score)

		if score > bestScore {
			bestScore = score
			bestMatch = i
		}
	}

	if bestMatch >= 0 && bestScore > 0.6 {
		command := a.savedGestures[bestMatch].Command
		log.Printf("Matched gesture: %s (score: %.3f)", command, bestScore)

		if err := executeCommand(command); err != nil {
			log.Printf("Failed to execute command: %v", err)
		} else {
			log.Printf("Executed: %s", command)
		}

		a.isExiting = true
		a.exitStartTime = time.Now()
		window.DisableInput()
		a.spawnExitWisps(x, y)
	} else {
		log.Printf("No confident match (best score: %.3f)", bestScore)
	}
}

func main() {
	learnCommand := flag.String("learn", "", "Learn a new gesture for the specified command")
	listGestures := flag.Bool("list", false, "List all registered gestures")
	removeGesture := flag.String("remove", "", "Remove a gesture by command name")
	flag.Parse()

	if flag.NArg() > 0 {
		log.Fatalf("Unknown arguments: %v", flag.Args())
	}

	if *listGestures {
		gestures, err := loadGestures()
		if err != nil {
			log.Fatal("Failed to load gestures:", err)
		}
		if len(gestures) == 0 {
			println("No gestures registered")
		} else {
			println("Registered gestures:")
			for _, g := range gestures {
				println("  ", g.Command)
			}
		}
		return
	}

	if *removeGesture != "" {
		gestures, err := loadGestures()
		if err != nil {
			log.Fatal("Failed to load gestures:", err)
		}

		found := false
		for i, g := range gestures {
			if g.Command == *removeGesture {
				gestures = append(gestures[:i], gestures[i+1:]...)
				found = true
				break
			}
		}

		if !found {
			log.Fatalf("Gesture not found: %s", *removeGesture)
		}

		configFile, err := config.GetPath()
		if err != nil {
			log.Fatal("Failed to get config path:", err)
		}

		data, err := json.Marshal(gestures)
		if err != nil {
			log.Fatal("Failed to marshal gestures:", err)
		}

		if err := os.WriteFile(configFile, data, 0644); err != nil {
			log.Fatal("Failed to save gestures:", err)
		}

		println("Removed gesture:", *removeGesture)
		return
	}

	window, err := wayland.NewWaylandWindow()
	if err != nil {
		log.Fatal("Failed to create Wayland window:", err)
	}
	defer window.Destroy()

	app := &App{startTime: time.Now()}

	if *learnCommand != "" {
		app.learnMode = true
		app.learnCommand = *learnCommand
		log.Printf("Learn mode: Draw the gesture 3 times for command '%s'", *learnCommand)
	} else {
		gestures, err := loadGestures()
		if err != nil {
			log.Fatal("Failed to load gestures:", err)
		}
		app.savedGestures = gestures
		log.Printf("Loaded %d gesture(s)", len(gestures))
	}

	if err := app.initGL(); err != nil {
		log.Fatal("Failed to initialize OpenGL:", err)
	}

	gl.ClearColor(0, 0, 0, 0)

	for i := 0; i < 5; i++ {
		window.PollEvents()
		gl.Clear(gl.COLOR_BUFFER_BIT)
		window.SwapBuffers()
	}

	x, y := window.GetCursorPos()
	app.lastCursorX = float32(x)
	app.lastCursorY = float32(y)

	lastTime := time.Now()
	var wasPressed bool

	for !window.ShouldClose() {
		now := time.Now()
		dt := float32(now.Sub(lastTime).Seconds())
		lastTime = now

		window.PollEvents()
		app.updateCursor(window)

		if key, state, hasKey := window.GetLastKey(); hasKey {
			if state == 1 && key == 0xff1b {
				if !app.isExiting {
					app.isExiting = true
					app.exitStartTime = time.Now()
					window.DisableInput()
					x, y := window.GetCursorPos()
					app.spawnExitWisps(float32(x), float32(y))
				}
			}
			window.ClearLastKey()
		}

		if app.isExiting {
			if time.Since(app.exitStartTime).Seconds() > 0.8 {
				break
			}
		}
		isPressed := window.GetMouseButton()
		if isPressed && !wasPressed {
			app.isDrawing = true
		} else if !isPressed && wasPressed {
			app.isDrawing = false

			if app.learnMode && len(app.points) > 0 {
				processed := stroke.ProcessStroke(app.points)
				app.learnGestures = append(app.learnGestures, processed)
				app.learnCount++
				log.Printf("Captured gesture %d/3", app.learnCount)

				app.points = nil

				if app.learnCount >= 3 {
					if err := saveGesture(app.learnCommand, app.learnGestures); err != nil {
						log.Fatal("Failed to save gesture:", err)
					}
					log.Printf("Gesture saved for command: %s", app.learnCommand)

					app.isExiting = true
					app.exitStartTime = time.Now()
					window.DisableInput()
					x, y := window.GetCursorPos()
					app.spawnExitWisps(float32(x), float32(y))
				}
			} else if !app.learnMode && len(app.points) > 0 {
				x, y := window.GetCursorPos()
				app.recognizeAndExecute(window, float32(x), float32(y))
				app.points = nil
			}
		}
		wasPressed = isPressed

		if app.isDrawing {
			x, y := window.GetCursorPos()
			app.addPoint(float32(x), float32(y))
			app.spawnCursorSparkles(float32(x), float32(y))
		}

		app.updateParticles(dt)
		app.draw(window)
		window.SwapBuffers()
	}
}
