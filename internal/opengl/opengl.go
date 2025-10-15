package opengl

import (
	"log"

	"github.com/ThatOtherAndrew/Hexecute/internal/models"
	"github.com/ThatOtherAndrew/Hexecute/internal/shaders"
	"github.com/go-gl/gl/v4.1-core/gl"
)

type App struct {
	app *models.App
}

func New(app *models.App) *App {
	return &App{app: app}
}

func (a *App) InitGL() error {
	if err := gl.Init(); err != nil {
		return err
	}

	vertShader, err := shaders.CompileShaderFromSource(shaders.LineVertex, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}
	fragShader, err := shaders.CompileShaderFromSource(shaders.LineFragment, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	a.app.Program = gl.CreateProgram()
	gl.AttachShader(a.app.Program, vertShader)
	gl.AttachShader(a.app.Program, fragShader)
	gl.LinkProgram(a.app.Program)

	var status int32
	gl.GetProgramiv(a.app.Program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(a.app.Program, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetProgramInfoLog(a.app.Program, logLength, nil, &logMsg[0])
		log.Fatalf("Failed to link program: %s", logMsg)
	}

	gl.DeleteShader(vertShader)
	gl.DeleteShader(fragShader)

	particleVertShader, err := shaders.CompileShaderFromSource(
		shaders.ParticleVertex,
		gl.VERTEX_SHADER,
	)
	if err != nil {
		return err
	}
	particleFragShader, err := shaders.CompileShaderFromSource(
		shaders.ParticleFragment,
		gl.FRAGMENT_SHADER,
	)
	if err != nil {
		return err
	}

	a.app.ParticleProgram = gl.CreateProgram()
	gl.AttachShader(a.app.ParticleProgram, particleVertShader)
	gl.AttachShader(a.app.ParticleProgram, particleFragShader)
	gl.LinkProgram(a.app.ParticleProgram)

	gl.GetProgramiv(a.app.ParticleProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(a.app.ParticleProgram, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetProgramInfoLog(a.app.ParticleProgram, logLength, nil, &logMsg[0])
		log.Fatalf("Failed to link particle program: %s", logMsg)
	}

	gl.DeleteShader(particleVertShader)
	gl.DeleteShader(particleFragShader)

	gl.GenVertexArrays(1, &a.app.Vao)
	gl.GenBuffers(1, &a.app.Vbo)

	gl.BindVertexArray(a.app.Vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, a.app.Vbo)

	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 5*4, nil)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, 5*4, 2*4)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointerWithOffset(2, 1, gl.FLOAT, false, 5*4, 4*4)
	gl.EnableVertexAttribArray(2)

	gl.BindVertexArray(0)

	gl.GenVertexArrays(1, &a.app.ParticleVAO)
	gl.GenBuffers(1, &a.app.ParticleVBO)

	gl.BindVertexArray(a.app.ParticleVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, a.app.ParticleVBO)

	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 6*4, nil)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(1, 1, gl.FLOAT, false, 6*4, 2*4)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointerWithOffset(2, 1, gl.FLOAT, false, 6*4, 3*4)
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointerWithOffset(3, 1, gl.FLOAT, false, 6*4, 4*4)
	gl.EnableVertexAttribArray(3)
	gl.VertexAttribPointerWithOffset(4, 1, gl.FLOAT, false, 6*4, 5*4)
	gl.EnableVertexAttribArray(4)

	gl.BindVertexArray(0)

	bgVertShader, err := shaders.CompileShaderFromSource(
		shaders.BackgroundVertex,
		gl.VERTEX_SHADER,
	)
	if err != nil {
		return err
	}
	bgFragShader, err := shaders.CompileShaderFromSource(
		shaders.BackgroundFragment,
		gl.FRAGMENT_SHADER,
	)
	if err != nil {
		return err
	}

	a.app.BgProgram = gl.CreateProgram()
	gl.AttachShader(a.app.BgProgram, bgVertShader)
	gl.AttachShader(a.app.BgProgram, bgFragShader)
	gl.LinkProgram(a.app.BgProgram)

	gl.GetProgramiv(a.app.BgProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(a.app.BgProgram, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetProgramInfoLog(a.app.BgProgram, logLength, nil, &logMsg[0])
		log.Fatalf("Failed to link background program: %s", logMsg)
	}

	gl.DeleteShader(bgVertShader)
	gl.DeleteShader(bgFragShader)

	gl.GenVertexArrays(1, &a.app.BgVAO)
	gl.GenBuffers(1, &a.app.BgVBO)

	gl.BindVertexArray(a.app.BgVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, a.app.BgVBO)

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

	cursorGlowVertShader, err := shaders.CompileShaderFromSource(
		shaders.CursorGlowVertex,
		gl.VERTEX_SHADER,
	)
	if err != nil {
		return err
	}
	cursorGlowFragShader, err := shaders.CompileShaderFromSource(
		shaders.CursorGlowFragment,
		gl.FRAGMENT_SHADER,
	)
	if err != nil {
		return err
	}

	a.app.CursorGlowProgram = gl.CreateProgram()
	gl.AttachShader(a.app.CursorGlowProgram, cursorGlowVertShader)
	gl.AttachShader(a.app.CursorGlowProgram, cursorGlowFragShader)
	gl.LinkProgram(a.app.CursorGlowProgram)

	gl.GetProgramiv(a.app.CursorGlowProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(a.app.CursorGlowProgram, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetProgramInfoLog(a.app.CursorGlowProgram, logLength, nil, &logMsg[0])
		log.Fatalf("Failed to link cursor glow program: %s", logMsg)
	}

	gl.DeleteShader(cursorGlowVertShader)
	gl.DeleteShader(cursorGlowFragShader)

	gl.GenVertexArrays(1, &a.app.CursorGlowVAO)
	gl.GenBuffers(1, &a.app.CursorGlowVBO)

	gl.BindVertexArray(a.app.CursorGlowVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, a.app.CursorGlowVBO)

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
