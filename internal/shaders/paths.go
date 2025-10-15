package shaders

import _ "embed"

// TODO: select one to choose embed the shaders or place like system files.
const BackgroundFragmentPath = "internal/shaders/background.frag.glsl"
const BackgroundVertexPath = "internal/shaders/background.vert.glsl"
const CursorGlowFragmentPath = "internal/shaders/cursorGlow.frag.glsl"
const CursorGlowVertexPath = "internal/shaders/cursorGlow.vert.glsl"
const LineFragmentPath = "internal/shaders/line.frag.glsl"
const LineVertexPath = "internal/shaders/line.vert.glsl"
const ParticleVertexPath = "internal/shaders/particle.vert.glsl"
const ParticleFragmentPath = "internal/shaders/particle.frag.glsl"

// Vertex shaders
//
//go:embed background.vert.glsl
var BackgroundVertex string

//go:embed cursorGlow.vert.glsl
var CursorGlowVertex string

//go:embed line.vert.glsl
var LineVertex string

//go:embed particle.vert.glsl
var ParticleVertex string

// Fragment shaders
//
//go:embed background.frag.glsl
var BackgroundFragment string

//go:embed cursorGlow.frag.glsl
var CursorGlowFragment string

//go:embed line.frag.glsl
var LineFragment string

//go:embed particle.frag.glsl
var ParticleFragment string
