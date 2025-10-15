package models

import (
	"time"
)

type Point struct {
	X, Y     float32
	BornTime time.Time `json:"-"`
}

type Particle struct {
	X, Y    float32
	VX, VY  float32
	Life    float32
	MaxLife float32
	Size    float32
	Hue     float32
}

type GestureConfig struct {
	Command   string    `json:"command"`
	Templates [][]Point `json:"templates"`
}

type App struct {
	Points            []Point
	Particles         []Particle
	IsDrawing         bool
	Vao               uint32
	Vbo               uint32
	Program           uint32
	ParticleVAO       uint32
	ParticleVBO       uint32
	ParticleProgram   uint32
	BgVAO             uint32
	BgVBO             uint32
	BgProgram         uint32
	CursorGlowVAO     uint32
	CursorGlowVBO     uint32
	CursorGlowProgram uint32
	StartTime         time.Time
	LastCursorX       float32
	LastCursorY       float32
	CursorVelocity    float32
	SmoothVelocity    float32
	SmoothRotation    float32
	SmoothDrawing     float32
	IsExiting         bool
	ExitStartTime     time.Time
	LearnMode         bool
	LearnCommand      string
	LearnGestures     [][]Point
	LearnCount        int
	SavedGestures     []GestureConfig
}
