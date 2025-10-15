package models

import "time"

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
