package main

import (
	"log"
	"math"
	"os"
	"os/exec"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type Point struct {
	X, Y float32
}

type Game struct {
	points    []Point
	isDrawing bool
	vao       uint32
	vbo       uint32
	program   uint32
}

const vertexShader = `
#version 410 core
layout (location = 0) in vec2 position;
layout (location = 1) in vec2 offset;

uniform vec2 resolution;
uniform float thickness;

void main() {
	vec2 pos = position + offset * thickness;
	vec2 normalized = (pos / resolution) * 2.0 - 1.0;
	normalized.y = -normalized.y; // Flip Y coordinate
	gl_Position = vec4(normalized, 0.0, 1.0);
}
` + "\x00"

const fragmentShader = `
#version 410 core
out vec4 FragColor;

void main() {
	FragColor = vec4(1.0, 0.0, 1.0, 1.0); // Magenta color
}
` + "\x00"

func init() {
	runtime.LockOSThread()
}

func (g *Game) initGL() error {
	if err := gl.Init(); err != nil {
		return err
	}

	// Compile vertex shader
	vertShader, err := compileShader(vertexShader, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	// Compile fragment shader
	fragShader, err := compileShader(fragmentShader, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	// Link program
	g.program = gl.CreateProgram()
	gl.AttachShader(g.program, vertShader)
	gl.AttachShader(g.program, fragShader)
	gl.LinkProgram(g.program)

	// Check linking status
	var status int32
	gl.GetProgramiv(g.program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(g.program, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetProgramInfoLog(g.program, logLength, nil, &logMsg[0])
		log.Fatalf("Failed to link program: %s", logMsg)
	}

	// Clean up shaders
	gl.DeleteShader(vertShader)
	gl.DeleteShader(fragShader)

	// Create VAO and VBO
	gl.GenVertexArrays(1, &g.vao)
	gl.GenBuffers(1, &g.vbo)

	gl.BindVertexArray(g.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, g.vbo)
	// Position attribute (location 0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, nil)
	gl.EnableVertexAttribArray(0)
	// Offset attribute (location 1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)

	// Enable blending for smooth lines
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

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
	newPoint := Point{X: x, Y: y}

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

	const MAX_POINTS = 2048

	if shouldAdd {
		g.points = append(g.points, newPoint)
		if len(g.points) > MAX_POINTS {
			g.points = g.points[len(g.points)-MAX_POINTS:]
		}
	}
}

func (g *Game) draw(window *glfw.Window) {
	gl.Clear(gl.COLOR_BUFFER_BIT)

	if len(g.points) < 2 {
		return
	}

	// Build triangle strip vertices with perpendicular offsets
	vertices := make([]float32, 0, len(g.points)*8) // 4 floats per vertex, 2 vertices per point

	for i := range g.points {
		var perpX, perpY float32

		if i == 0 {
			// First point: use direction to next point
			dx := g.points[i+1].X - g.points[i].X
			dy := g.points[i+1].Y - g.points[i].Y
			length := float32(1.0) / float32(math.Sqrt(float64(dx*dx+dy*dy)))
			perpX = -dy * length
			perpY = dx * length
		} else if i == len(g.points)-1 {
			// Last point: use direction from previous point
			dx := g.points[i].X - g.points[i-1].X
			dy := g.points[i].Y - g.points[i-1].Y
			length := float32(1.0) / float32(math.Sqrt(float64(dx*dx+dy*dy)))
			perpX = -dy * length
			perpY = dx * length
		} else {
			// Middle points: average of two directions
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

		// Add two vertices per point (one on each side)
		// Vertex 1: position + offset on one side
		vertices = append(vertices, g.points[i].X, g.points[i].Y, perpX, perpY)
		// Vertex 2: position + offset on other side
		vertices = append(vertices, g.points[i].X, g.points[i].Y, -perpX, -perpY)
	}

	// Update VBO
	gl.BindBuffer(gl.ARRAY_BUFFER, g.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.DYNAMIC_DRAW)

	// Get window size for shader uniform
	width, height := window.GetSize()

	// Draw thick lines as triangle strip
	gl.UseProgram(g.program)
	resolutionLoc := gl.GetUniformLocation(g.program, gl.Str("resolution\x00"))
	gl.Uniform2f(resolutionLoc, float32(width), float32(height))
	thicknessLoc := gl.GetUniformLocation(g.program, gl.Str("thickness\x00"))
	gl.Uniform1f(thicknessLoc, 10.0) // Line thickness in pixels

	gl.BindVertexArray(g.vao)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(g.points)*2))
	gl.BindVertexArray(0)
}

func main() {
	// Make Hyprland play nice
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		exec.Command("hyprctl", "keyword", "windowrulev2", "center,pin,noborder,noanim,noblur,title:^(hexecute)$").Run()
	}

	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		log.Fatal("Failed to initialize GLFW:", err)
	}
	defer glfw.Terminate()

	// Get monitor size
	monitor := glfw.GetPrimaryMonitor()
	mode := monitor.GetVideoMode()

	// Configure window
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Decorated, glfw.False)
	glfw.WindowHint(glfw.Floating, glfw.True)
	glfw.WindowHint(glfw.TransparentFramebuffer, glfw.True)

	// Create window
	window, err := glfw.CreateWindow(mode.Width, mode.Height, "hexecute", nil, nil)
	if err != nil {
		log.Fatal("Failed to create window:", err)
	}
	window.MakeContextCurrent()
	glfw.SwapInterval(1)

	// Initialize game and OpenGL
	game := &Game{
		points:    []Point{},
		isDrawing: false,
	}

	if err := game.initGL(); err != nil {
		log.Fatal("Failed to initialize OpenGL:", err)
	}

	// Set clear color to transparent
	gl.ClearColor(0, 0, 0, 0)

	// Mouse button callback
	window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		if button == glfw.MouseButtonLeft {
			switch action {
			case glfw.Press:
				if !game.isDrawing {
					log.Println("Start drawing")
					game.isDrawing = true
				}
			case glfw.Release:
				if game.isDrawing {
					log.Println("Stop drawing")
					game.isDrawing = false
					game.points = []Point{}
				}
			}
		}
	})

	// Main loop
	for !window.ShouldClose() {
		// Update
		if game.isDrawing {
			x, y := window.GetCursorPos()
			game.addPoint(float32(x), float32(y))
		}

		// Draw
		game.draw(window)

		// Swap and poll
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
