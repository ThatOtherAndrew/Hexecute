package main

import (
	"log"
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type Point struct {
	X, Y     float32
	BornTime time.Time
}

type Particle struct {
	X, Y    float32
	VX, VY  float32
	Life    float32
	MaxLife float32
	Size    float32
	Hue     float32
}

type Game struct {
	points          []Point
	particles       []Particle
	isDrawing       bool
	vao             uint32
	vbo             uint32
	program         uint32
	particleVAO     uint32
	particleVBO     uint32
	particleProgram uint32
	flashVAO        uint32
	flashVBO        uint32
	flashProgram    uint32
	startTime       time.Time
}

const vertexShader = `
#version 410 core
layout (location = 0) in vec2 position;
layout (location = 1) in vec2 offset;
layout (location = 2) in float alpha;

uniform vec2 resolution;
uniform float thickness;

out float vAlpha;
out vec2 vPosition;

void main() {
	vec2 pos = position + offset * thickness;
	vec2 normalized = (pos / resolution) * 2.0 - 1.0;
	normalized.y = -normalized.y;
	gl_Position = vec4(normalized, 0.0, 1.0);
	vAlpha = alpha;
	vPosition = pos;
}
` + "\x00"

const fragmentShader = `
#version 410 core
in float vAlpha;
in vec2 vPosition;
out vec4 FragColor;

uniform float time;

vec3 hsv2rgb(vec3 c) {
	vec4 K = vec4(1.0, 2.0 / 3.0, 1.0 / 3.0, 3.0);
	vec3 p = abs(fract(c.xxx + K.xyz) * 6.0 - K.www);
	return c.z * mix(K.xxx, clamp(p - K.xxx, 0.0, 1.0), c.y);
}

void main() {
	// Animated rainbow gradient with shimmer
	float hue = mod(vPosition.x * 0.001 + vPosition.y * 0.001 + time * 0.5, 1.0);
	vec3 color = hsv2rgb(vec3(hue, 0.8, 1.0));

	// Add sparkle effect
	float sparkle = sin(vPosition.x * 0.1 + time * 3.0) * sin(vPosition.y * 0.1 + time * 2.0);
	sparkle = smoothstep(0.7, 1.0, sparkle) * 0.5;

	color += vec3(sparkle);

	// Glow effect - brighter in center
	float glow = 1.0 + sparkle * 2.0;

	FragColor = vec4(color * glow, vAlpha);
}
` + "\x00"

const particleVertexShader = `
#version 410 core
layout (location = 0) in vec2 position;
layout (location = 1) in float life;
layout (location = 2) in float maxLife;
layout (location = 3) in float size;
layout (location = 4) in float hue;

uniform vec2 resolution;

out float vLife;
out float vHue;

void main() {
	vec2 normalized = (position / resolution) * 2.0 - 1.0;
	normalized.y = -normalized.y;
	gl_Position = vec4(normalized, 0.0, 1.0);
	gl_PointSize = size * (life / maxLife);
	vLife = life / maxLife;
	vHue = hue;
}
` + "\x00"

const particleFragmentShader = `
#version 410 core
in float vLife;
in float vHue;
out vec4 FragColor;

vec3 hsv2rgb(vec3 c) {
	vec4 K = vec4(1.0, 2.0 / 3.0, 1.0 / 3.0, 3.0);
	vec3 p = abs(fract(c.xxx + K.xyz) * 6.0 - K.www);
	return c.z * mix(K.xxx, clamp(p - K.xxx, 0.0, 1.0), c.y);
}

void main() {
	// Circular particle shape
	vec2 coord = gl_PointCoord - vec2(0.5);
	float dist = length(coord);
	if (dist > 0.5) discard;

	// Soft edge
	float alpha = smoothstep(0.5, 0.2, dist) * vLife;

	// Rainbow color
	vec3 color = hsv2rgb(vec3(vHue, 0.9, 1.0));

	// Glow
	color *= (1.0 + (1.0 - dist * 2.0) * 2.0);

	FragColor = vec4(color, alpha * 0.8);
}
` + "\x00"

const flashVertexShader = `
#version 410 core
layout (location = 0) in vec2 position;

void main() {
	gl_Position = vec4(position, 0.0, 1.0);
}
` + "\x00"

const flashFragmentShader = `
#version 410 core
out vec4 FragColor;

uniform float alpha;

void main() {
	FragColor = vec4(255.0, 255.0, 255.0, alpha);
}
` + "\x00"

func init() {
	runtime.LockOSThread()
}

func (g *Game) initGL() error {
	if err := gl.Init(); err != nil {
		return err
	}

	// Compile line shaders
	vertShader, err := compileShader(vertexShader, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}
	fragShader, err := compileShader(fragmentShader, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	g.program = gl.CreateProgram()
	gl.AttachShader(g.program, vertShader)
	gl.AttachShader(g.program, fragShader)
	gl.LinkProgram(g.program)

	var status int32
	gl.GetProgramiv(g.program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(g.program, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetProgramInfoLog(g.program, logLength, nil, &logMsg[0])
		log.Fatalf("Failed to link program: %s", logMsg)
	}

	gl.DeleteShader(vertShader)
	gl.DeleteShader(fragShader)

	// Compile particle shaders
	particleVertShader, err := compileShader(particleVertexShader, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}
	particleFragShader, err := compileShader(particleFragmentShader, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	g.particleProgram = gl.CreateProgram()
	gl.AttachShader(g.particleProgram, particleVertShader)
	gl.AttachShader(g.particleProgram, particleFragShader)
	gl.LinkProgram(g.particleProgram)

	gl.GetProgramiv(g.particleProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(g.particleProgram, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetProgramInfoLog(g.particleProgram, logLength, nil, &logMsg[0])
		log.Fatalf("Failed to link particle program: %s", logMsg)
	}

	gl.DeleteShader(particleVertShader)
	gl.DeleteShader(particleFragShader)

	// Create VAO and VBO for lines
	gl.GenVertexArrays(1, &g.vao)
	gl.GenBuffers(1, &g.vbo)

	gl.BindVertexArray(g.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, g.vbo)

	// Position (location 0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 5*4, nil)
	gl.EnableVertexAttribArray(0)
	// Offset (location 1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)
	// Alpha (location 2)
	gl.VertexAttribPointer(2, 1, gl.FLOAT, false, 5*4, gl.PtrOffset(4*4))
	gl.EnableVertexAttribArray(2)

	gl.BindVertexArray(0)

	// Create VAO and VBO for particles
	gl.GenVertexArrays(1, &g.particleVAO)
	gl.GenBuffers(1, &g.particleVBO)

	gl.BindVertexArray(g.particleVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, g.particleVBO)

	// Position (location 0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 6*4, nil)
	gl.EnableVertexAttribArray(0)
	// Life (location 1)
	gl.VertexAttribPointer(1, 1, gl.FLOAT, false, 6*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)
	// MaxLife (location 2)
	gl.VertexAttribPointer(2, 1, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(2)
	// Size (location 3)
	gl.VertexAttribPointer(3, 1, gl.FLOAT, false, 6*4, gl.PtrOffset(4*4))
	gl.EnableVertexAttribArray(3)
	// Hue (location 4)
	gl.VertexAttribPointer(4, 1, gl.FLOAT, false, 6*4, gl.PtrOffset(5*4))
	gl.EnableVertexAttribArray(4)

	gl.BindVertexArray(0)

	// Compile flash shaders
	flashVertShader, err := compileShader(flashVertexShader, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}
	flashFragShader, err := compileShader(flashFragmentShader, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	g.flashProgram = gl.CreateProgram()
	gl.AttachShader(g.flashProgram, flashVertShader)
	gl.AttachShader(g.flashProgram, flashFragShader)
	gl.LinkProgram(g.flashProgram)

	gl.GetProgramiv(g.flashProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(g.flashProgram, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetProgramInfoLog(g.flashProgram, logLength, nil, &logMsg[0])
		log.Fatalf("Failed to link flash program: %s", logMsg)
	}

	gl.DeleteShader(flashVertShader)
	gl.DeleteShader(flashFragShader)

	// Create VAO and VBO for flash overlay (fullscreen quad)
	gl.GenVertexArrays(1, &g.flashVAO)
	gl.GenBuffers(1, &g.flashVBO)

	gl.BindVertexArray(g.flashVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, g.flashVBO)

	// Fullscreen quad vertices
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

	// Enable blending
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE) // Additive blending for glow
	gl.Enable(gl.PROGRAM_POINT_SIZE)

	return nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetShaderInfoLog(shader, logLength, nil, &logMsg[0])
		log.Fatalf("Failed to compile shader: %s", logMsg)
	}

	return shader, nil
}

func (g *Game) addPoint(x, y float32) {
	newPoint := Point{X: x, Y: y, BornTime: time.Now()}

	shouldAdd := false
	if len(g.points) == 0 {
		shouldAdd = true
	} else {
		lastPoint := g.points[len(g.points)-1]
		dx := newPoint.X - lastPoint.X
		dy := newPoint.Y - lastPoint.Y
		if dx*dx+dy*dy > 4 {
			shouldAdd = true

			// Spawn particles along the line
			for range 3 {
				angle := rand.Float64() * 2 * math.Pi
				speed := rand.Float32()*50 + 20
				g.particles = append(g.particles, Particle{
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
		g.points = append(g.points, newPoint)
		if len(g.points) > MAX_POINTS {
			g.points = g.points[len(g.points)-MAX_POINTS:]
		}
	}
}

func (g *Game) updateParticles(dt float32) {
	// Update particles
	for i := 0; i < len(g.particles); i++ {
		p := &g.particles[i]
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.VY += 100 * dt // Gravity
		p.Life -= dt

		if p.Life <= 0 {
			// Remove dead particle
			g.particles[i] = g.particles[len(g.particles)-1]
			g.particles = g.particles[:len(g.particles)-1]
			i--
		}
	}
}

func (g *Game) draw(window *WaylandWindow) {
	gl.Clear(gl.COLOR_BUFFER_BIT)

	currentTime := float32(time.Since(g.startTime).Seconds())

	// Draw main line with glow (multiple passes)
	for pass := range 3 {
		thickness := float32(15 + pass*8)         // Increasing thickness for glow layers
		alpha := float32(0.4 - float32(pass)*0.1) // Decreasing alpha for outer glow

		g.drawLine(window, thickness, alpha, currentTime)
	}

	// Draw particles
	g.drawParticles(window)

	// Draw flash overlay
	g.drawFlash(currentTime)
}

func (g *Game) drawLine(window *WaylandWindow, baseThickness, baseAlpha, currentTime float32) {
	if len(g.points) < 2 {
		return
	}

	// Build triangle strip vertices with fade
	vertices := make([]float32, 0, len(g.points)*10)

	for i := range g.points {
		// Calculate age-based fade
		age := float32(time.Since(g.points[i].BornTime).Seconds())
		fadeDuration := float32(1.5)
		fade := 1.0 - (age / fadeDuration)
		if fade < 0 {
			fade = 0
		}
		alpha := fade * baseAlpha

		// Calculate perpendicular direction
		var perpX, perpY float32

		if i == 0 {
			dx := g.points[i+1].X - g.points[i].X
			dy := g.points[i+1].Y - g.points[i].Y
			length := float32(1.0) / float32(math.Sqrt(float64(dx*dx+dy*dy)))
			perpX = -dy * length
			perpY = dx * length
		} else if i == len(g.points)-1 {
			dx := g.points[i].X - g.points[i-1].X
			dy := g.points[i].Y - g.points[i-1].Y
			length := float32(1.0) / float32(math.Sqrt(float64(dx*dx+dy*dy)))
			perpX = -dy * length
			perpY = dx * length
		} else {
			dx1 := g.points[i].X - g.points[i-1].X
			dy1 := g.points[i].Y - g.points[i-1].Y
			len1 := float32(math.Sqrt(float64(dx1*dx1 + dy1*dy1)))
			if len1 > 0 {
				dx1 /= len1
				dy1 /= len1
			}

			dx2 := g.points[i+1].X - g.points[i].X
			dy2 := g.points[i+1].Y - g.points[i].Y
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

		// Add two vertices per point
		vertices = append(vertices, g.points[i].X, g.points[i].Y, perpX, perpY, alpha)
		vertices = append(vertices, g.points[i].X, g.points[i].Y, -perpX, -perpY, alpha)
	}

	// Remove old faded points
	cutoff := time.Now().Add(-1500 * time.Millisecond)
	for len(g.points) > 0 && g.points[0].BornTime.Before(cutoff) {
		g.points = g.points[1:]
	}

	if len(vertices) == 0 {
		return
	}

	// Update VBO
	gl.BindBuffer(gl.ARRAY_BUFFER, g.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.DYNAMIC_DRAW)

	width, height := window.GetSize()

	// Draw
	gl.UseProgram(g.program)
	resolutionLoc := gl.GetUniformLocation(g.program, gl.Str("resolution\x00"))
	gl.Uniform2f(resolutionLoc, float32(width), float32(height))
	thicknessLoc := gl.GetUniformLocation(g.program, gl.Str("thickness\x00"))
	gl.Uniform1f(thicknessLoc, baseThickness)
	timeLoc := gl.GetUniformLocation(g.program, gl.Str("time\x00"))
	gl.Uniform1f(timeLoc, currentTime)

	gl.BindVertexArray(g.vao)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(g.points)*2))
	gl.BindVertexArray(0)
}

func (g *Game) drawParticles(window *WaylandWindow) {
	if len(g.particles) == 0 {
		return
	}

	vertices := make([]float32, 0, len(g.particles)*6)
	for _, p := range g.particles {
		vertices = append(vertices, p.X, p.Y, p.Life, p.MaxLife, p.Size, p.Hue)
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, g.particleVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.DYNAMIC_DRAW)

	width, height := window.GetSize()

	gl.UseProgram(g.particleProgram)
	resolutionLoc := gl.GetUniformLocation(g.particleProgram, gl.Str("resolution\x00"))
	gl.Uniform2f(resolutionLoc, float32(width), float32(height))

	gl.BindVertexArray(g.particleVAO)
	gl.DrawArrays(gl.POINTS, 0, int32(len(g.particles)))
	gl.BindVertexArray(0)
}

func (g *Game) drawFlash(currentTime float32) {
	// Flash duration and fade
	flashDuration := float32(3.0) // 1.0 seconds
	if currentTime > flashDuration {
		return
	}

	// Ease-out quadratic easing (more gradual)
	progress := currentTime / flashDuration
	easedProgress := 1.0 - (1.0-progress)*(1.0-progress)
	alpha := 1.0 - easedProgress

	if alpha <= 0 {
		return
	}

	// Switch to alpha blending for the flash overlay
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.UseProgram(g.flashProgram)
	alphaLoc := gl.GetUniformLocation(g.flashProgram, gl.Str("alpha\x00"))
	gl.Uniform1f(alphaLoc, alpha)

	gl.BindVertexArray(g.flashVAO)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	gl.BindVertexArray(0)

	// Restore additive blending for subsequent draws
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
}

func main() {
	// Create Wayland window with layer shell
	window, err := NewWaylandWindow()
	if err != nil {
		log.Fatal("Failed to create Wayland window:", err)
	}
	defer window.Destroy()

	// Initialize game and OpenGL
	game := &Game{
		points:    []Point{},
		particles: []Particle{},
		isDrawing: false,
		startTime: time.Now(),
	}

	if err := game.initGL(); err != nil {
		log.Fatal("Failed to initialize OpenGL:", err)
	}

	// Set clear color to transparent
	gl.ClearColor(0, 0, 0, 0)

	lastTime := time.Now()
	var wasPressed bool

	// Give the compositor time to give us focus by doing a few initial frames
	for i := 0; i < 5; i++ {
		window.PollEvents()
		gl.Clear(gl.COLOR_BUFFER_BIT)
		window.SwapBuffers()
	}

	// Main loop
	for !window.ShouldClose() {
		// Calculate delta time
		now := time.Now()
		dt := float32(now.Sub(lastTime).Seconds())
		lastTime = now

		// Poll events
		window.PollEvents()

		// Handle keyboard
		if key, state, hasKey := window.GetLastKey(); hasKey {
			if state == 1 && key == 1 { // Escape key pressed
				break
			}
			window.ClearLastKey()
		}

		// Handle mouse button
		isPressed := window.GetMouseButton()
		if isPressed && !wasPressed {
			game.isDrawing = true
		} else if !isPressed && wasPressed {
			game.isDrawing = false
		}
		wasPressed = isPressed

		// Update
		if game.isDrawing {
			x, y := window.GetCursorPos()
			game.addPoint(float32(x), float32(y))
		}

		game.updateParticles(dt)

		// Draw
		game.draw(window)

		// Swap buffers
		window.SwapBuffers()
	}
}
