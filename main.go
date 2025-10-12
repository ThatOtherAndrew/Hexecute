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

type Launcher struct {
	points            []Point
	particles         []Particle
	isDrawing         bool
	vao               uint32
	vbo               uint32
	program           uint32
	particleVAO       uint32
	particleVBO       uint32
	particleProgram   uint32
	flashVAO          uint32
	flashVBO          uint32
	flashProgram      uint32
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

const backgroundDimFragmentShader = `
#version 410 core
out vec4 FragColor;

uniform float alpha;
uniform vec2 cursorPos;
uniform vec2 resolution;

void main() {
	// Calculate distance from cursor in screen space
	vec2 fragCoord = gl_FragCoord.xy;
	float dist = length(fragCoord - cursorPos);

	// Create circular glow around cursor (flashlight effect)
	float glowRadius = 300.0; // Radius of the glow
	float glowFalloff = smoothstep(0.0, glowRadius, dist);

	// Reduce alpha near cursor (more transparent = see through more)
	float cursorTransparency = mix(0.3, 1.0, glowFalloff);

	// Dimmed background with flashlight effect
	FragColor = vec4(0., 0., 0., alpha * cursorTransparency);
}
` + "\x00"

const cursorGlowVertexShader = `
#version 410 core
layout (location = 0) in vec2 position;

uniform vec2 cursorPos;
uniform vec2 resolution;
uniform float glowSize;
uniform float rotation;

out vec2 vTexCoord;

void main() {
	// Apply rotation to position
	float c = cos(rotation);
	float s = sin(rotation);
	vec2 rotatedPos = vec2(
		position.x * c - position.y * s,
		position.x * s + position.y * c
	);

	// Scale quad around cursor position
	vec2 worldPos = cursorPos + rotatedPos * glowSize;
	vec2 normalized = (worldPos / resolution) * 2.0 - 1.0;
	normalized.y = -normalized.y;
	gl_Position = vec4(normalized, 0.0, 1.0);
	// Convert -1..1 to 0..1 for proper UV coordinates
	vTexCoord = rotatedPos * 0.5 + 0.5;
}
` + "\x00"

const cursorGlowFragmentShader = `
#version 410 core
in vec2 vTexCoord;
out vec4 FragColor;

uniform float time;
uniform float velocity;
uniform float isDrawing;

vec3 hsv2rgb(vec3 c) {
	vec4 K = vec4(1.0, 2.0 / 3.0, 1.0 / 3.0, 3.0);
	vec3 p = abs(fract(c.xxx + K.xyz) * 6.0 - K.www);
	return c.z * mix(K.xxx, clamp(p - K.xxx, 0.0, 1.0), c.y);
}

// Smooth minimum for metaball blending
float smin(float a, float b, float k) {
	float h = clamp(0.5 + 0.5 * (b - a) / k, 0.0, 1.0);
	return mix(b, a, h) - k * h * (1.0 - h);
}

// Simple hash for pseudo-random noise
float hash(vec2 p) {
	p = fract(p * vec2(123.34, 456.21));
	p += dot(p, p + 45.32);
	return fract(p.x * p.y);
}

// Smooth noise
float noise(vec2 p) {
	vec2 i = floor(p);
	vec2 f = fract(p);
	f = f * f * (3.0 - 2.0 * f);

	float a = hash(i);
	float b = hash(i + vec2(1.0, 0.0));
	float c = hash(i + vec2(0.0, 1.0));
	float d = hash(i + vec2(1.0, 1.0));

	return mix(mix(a, b, f.x), mix(c, d, f.x), f.y);
}

// Fractal noise
float fbm(vec2 p) {
	float value = 0.0;
	float amplitude = 0.5;
	float frequency = 1.0;

	for(int i = 0; i < 4; i++) {
		value += amplitude * noise(p * frequency);
		frequency *= 2.0;
		amplitude *= 0.5;
	}
	return value;
}

void main() {
	// Center coordinates from -1 to 1
	vec2 coord = vTexCoord * 2.0 - 1.0;

	// Velocity-based parameters (0 = static, 1 = fast)
	float velocityNorm = clamp(velocity * 0.01, 0.0, 1.0);

	// Drawing boost: extra energy when actively drawing
	float drawBoost = isDrawing * 0.7;

	// Energy level: high when moving or drawing, low when static
	float energy = mix(0.3, 1.0, velocityNorm) + drawBoost;

	// Create multiple animated orbs for metaball effect
	float sdf = 1000.0; // Start with large distance

	// Central orb pulses more aggressively when drawing
	float centralSize = mix(0.15, 0.35, velocityNorm) + isDrawing * 0.05;
	float pulseSpeed = (3.0 + velocityNorm * 2.0) * (1.0 + isDrawing * 0.75);
	float pulseAmount = (0.1 * energy + isDrawing * 0.075);
	float pulse = sin(time * pulseSpeed) * pulseAmount + 0.9;
	float centralDist = length(coord) - centralSize * pulse;
	sdf = centralDist;

	// More blobs appear when drawing (smooth float to avoid discrete jumps)
	float numBlobsFloat = mix(5.0, 9.0, velocityNorm) + isDrawing * 1.0;
	int numBlobs = int(numBlobsFloat);
	float blobFade = fract(numBlobsFloat); // Fade in/out the last blob

	for(int i = 0; i < 10; i++) {
		if(i > numBlobs) break; // One extra for fade
		if(i == numBlobs && blobFade < 0.01) break; // Skip if fade is negligible

		// Rotation speed (unchanged by drawing state)
		float baseRotation = time * 0.8;
		float velocityRotation = time * velocityNorm * 0.4;
		float angle = float(i) * 6.28 / float(numBlobs) + baseRotation + velocityRotation;

		// Radius expands outward when drawing
		float baseRadius = mix(0.2, 0.5, velocityNorm) + isDrawing * 0.075;
		float radiusVariation = sin(time * (1.5 + isDrawing * 0.5) + float(i) * 0.8) * mix(0.05, 0.15, velocityNorm);
		// Additional chaotic movement when drawing (very subtle)
		float chaoticRadius = sin(time * 4.0 + float(i) * 2.1) * cos(time * 3.5 + float(i) * 1.7) * 0.003 * isDrawing;
		float radius = baseRadius + radiusVariation + chaoticRadius;
		vec2 orbPos = vec2(cos(angle), sin(angle)) * radius;

		// Blobs grow and shrink more dramatically when drawing
		float baseBlobSize = mix(0.08, 0.18, velocityNorm) + isDrawing * 0.04;
		float sizeVariation = sin(time * (2.5 + isDrawing * 1.0) + float(i) * 0.6) * mix(0.02, 0.05, velocityNorm);
		float drawingGrowth = sin(time * 5.0 + float(i) * 1.3) * 0.03 * isDrawing;
		float blobSize = baseBlobSize + sizeVariation + drawingGrowth;
		float blobDist = length(coord - orbPos) - blobSize;

		// Fade out the last blob smoothly
		if(i == numBlobs) {
			blobDist += (1.0 - blobFade) * 0.5; // Make it smaller/further away as it fades
		}

		// More fluid blending when drawing
		float blendAmount = mix(0.15, 0.3, velocityNorm) + isDrawing * 0.075;
		sdf = smin(sdf, blobDist, blendAmount);
	}

	// Swirling tendrils with zoom effect when drawing
	// Zoom in by scaling noise coordinates when drawing
	float noiseZoom = 3.0 + isDrawing * 0.5;
	vec2 noiseCoord = coord * noiseZoom;
	noiseCoord += vec2(time * 0.3, time * 0.2);
	float swirl = fbm(noiseCoord) * 2.0 - 1.0;

	// Distort SDF with noise (less distortion when static)
	sdf += swirl * 0.1 * energy;

	// Convert SDF to intensity
	float intensity = exp(-max(sdf, 0.0) * 4.0);
	float outerGlow = exp(-max(sdf, 0.0) * 1.5) * 0.4 * energy;
	float innerGlow = exp(-max(sdf, 0.0) * 8.0) * 0.8;

	float totalIntensity = intensity + outerGlow + innerGlow;

	// Fade out at edges to eliminate square boundary
	float edgeDist = max(abs(coord.x), abs(coord.y));
	float edgeFade = smoothstep(1.0, 0.7, edgeDist);
	totalIntensity *= edgeFade;

	// Dynamic color based on position, time, and velocity
	float hueSpeed = mix(0.2, 0.6, velocityNorm);
	float hue = mod(time * hueSpeed + atan(coord.y, coord.x) / 6.28 + swirl * 0.3, 1.0);
	// Lower saturation to reduce rainbow halo intensity
	float saturation = mix(0.7, 0.75, velocityNorm);
	vec3 mainColor = hsv2rgb(vec3(hue, saturation, 1.0));

	// Secondary color for variation (also less saturated)
	vec3 accentColor = hsv2rgb(vec3(mod(hue + 0.5, 1.0), 0.75, 1.2));

	// Mix colors based on intensity layers
	vec3 finalColor = mainColor * intensity;
	finalColor += accentColor * innerGlow;
	finalColor += mainColor * 0.5 * outerGlow;

	// Energy sparkles (more when moving)
	float sparkle = noise(coord * 20.0 + time * 5.0 * energy);
	sparkle = smoothstep(0.85, 1.0, sparkle) * totalIntensity * velocityNorm;
	finalColor += vec3(1.0) * sparkle;

	// Edge highlighting for more definition
	float edge = smoothstep(0.05, -0.05, sdf) - smoothstep(0.15, 0.05, sdf);
	finalColor += accentColor * edge * energy;

	// Global pulse (more pronounced when drawing)
	float pulseIntensity = (0.1 + velocityNorm * 0.1 + isDrawing * 0.075);
	float globalPulse = sin(time * (2.5 + isDrawing * 0.75)) * pulseIntensity + 0.9;
	finalColor *= globalPulse;

	// Alpha based on total intensity and energy
	float alpha = clamp(totalIntensity * mix(0.8, 1.3, velocityNorm), 0.0, 1.0);

	FragColor = vec4(finalColor, alpha * 0.95);
}
` + "\x00"

func init() {
	runtime.LockOSThread()
}

func (g *Launcher) initGL() error {
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
	bgDimFragShader, err := compileShader(backgroundDimFragmentShader, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	g.flashProgram = gl.CreateProgram()
	gl.AttachShader(g.flashProgram, flashVertShader)
	gl.AttachShader(g.flashProgram, bgDimFragShader)
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
	gl.DeleteShader(bgDimFragShader)

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

	// Compile cursor glow shaders
	cursorGlowVertShader, err := compileShader(cursorGlowVertexShader, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}
	cursorGlowFragShader, err := compileShader(cursorGlowFragmentShader, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	g.cursorGlowProgram = gl.CreateProgram()
	gl.AttachShader(g.cursorGlowProgram, cursorGlowVertShader)
	gl.AttachShader(g.cursorGlowProgram, cursorGlowFragShader)
	gl.LinkProgram(g.cursorGlowProgram)

	gl.GetProgramiv(g.cursorGlowProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(g.cursorGlowProgram, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength)
		gl.GetProgramInfoLog(g.cursorGlowProgram, logLength, nil, &logMsg[0])
		log.Fatalf("Failed to link cursor glow program: %s", logMsg)
	}

	gl.DeleteShader(cursorGlowVertShader)
	gl.DeleteShader(cursorGlowFragShader)

	// Create VAO and VBO for cursor glow (quad)
	gl.GenVertexArrays(1, &g.cursorGlowVAO)
	gl.GenBuffers(1, &g.cursorGlowVBO)

	gl.BindVertexArray(g.cursorGlowVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, g.cursorGlowVBO)

	// Quad vertices centered at origin (will be positioned by shader)
	glowQuadVertices := []float32{
		-1.0, -1.0,
		1.0, -1.0,
		-1.0, 1.0,
		1.0, 1.0,
	}
	gl.BufferData(gl.ARRAY_BUFFER, len(glowQuadVertices)*4, gl.Ptr(glowQuadVertices), gl.STATIC_DRAW)

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

func (g *Launcher) addPoint(x, y float32) {
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

func (g *Launcher) spawnCursorSparkles(x, y float32) {
	// Spawn magical spark particles around cursor
	for range 3 {
		angle := rand.Float64() * 2 * math.Pi
		speed := rand.Float32()*80 + 40
		g.particles = append(g.particles, Particle{
			X:       x + (rand.Float32()-0.5)*8,
			Y:       y + (rand.Float32()-0.5)*8,
			VX:      float32(math.Cos(angle)) * speed,
			VY:      float32(math.Sin(angle))*speed - 30, // Slight upward bias
			Life:    0.8,
			MaxLife: 0.8,
			Size:    rand.Float32()*8 + 6,
			Hue:     rand.Float32(),
		})
	}
}

func (g *Launcher) updateParticles(dt float32) {
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

func (g *Launcher) updateCursor(dt float32, window *WaylandWindow) {
	x, y := window.GetCursorPos()
	fx, fy := float32(x), float32(y)

	// Calculate instantaneous velocity
	dx := fx - g.lastCursorX
	dy := fy - g.lastCursorY
	g.cursorVelocity = float32(math.Sqrt(float64(dx*dx + dy*dy)))

	// Smooth velocity to prevent jittering
	velocityDiff := g.cursorVelocity - g.smoothVelocity
	g.smoothVelocity += velocityDiff * 0.2 // Smooth velocity transitions

	// Calculate rotation angle from movement direction
	// Always update rotation, but with influence based on velocity
	if g.cursorVelocity > 0.1 { // Only update when there's meaningful movement
		targetRotation := float32(math.Atan2(float64(dy), float64(dx)))

		// Smooth the rotation with lerp, handling angle wrapping
		angleDiff := targetRotation - g.smoothRotation
		// Normalize angle difference to [-pi, pi]
		if angleDiff > math.Pi {
			angleDiff -= 2 * math.Pi
		} else if angleDiff < -math.Pi {
			angleDiff += 2 * math.Pi
		}

		// Smooth factor varies with velocity - slower movement = slower rotation update
		velocityFactor := float32(math.Min(float64(g.smoothVelocity/5.0), 1.0))
		smoothFactor := 0.03 + velocityFactor*0.08 // 0.03 to 0.11 based on velocity
		g.smoothRotation += angleDiff * smoothFactor
	}

	// Smooth drawing state with ease-out
	var targetDrawing float32
	if g.isDrawing {
		targetDrawing = 1.0
	} else {
		targetDrawing = 0.0
	}

	// Ease out cubic: 1 - (1-t)^3
	diff := targetDrawing - g.smoothDrawing
	t := float32(0.0375) // Transition speed (4x slower)
	g.smoothDrawing += diff * t

	g.lastCursorX = fx
	g.lastCursorY = fy
}

func (g *Launcher) draw(window *WaylandWindow) {
	gl.Clear(gl.COLOR_BUFFER_BIT)

	currentTime := float32(time.Since(g.startTime).Seconds())

	// Draw background first (underneath everything)
	g.drawBackground(currentTime, window)

	// Draw cursor glow (behind trails and particles) - instant position
	x, y := window.GetCursorPos()
	g.drawCursorGlow(window, float32(x), float32(y), currentTime)

	// Draw main line with glow (multiple passes)
	for pass := range 3 {
		thickness := float32(7 + pass*4)           // Increasing thickness for glow layers (half as wide)
		alpha := float32(0.7 - float32(pass)*0.15) // Decreasing alpha for outer glow (brighter)

		g.drawLine(window, thickness, alpha, currentTime)
	}

	// Draw particles
	g.drawParticles(window)
}

func (g *Launcher) drawLine(window *WaylandWindow, baseThickness, baseAlpha, currentTime float32) {
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

func (g *Launcher) drawParticles(window *WaylandWindow) {
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

func (g *Launcher) drawBackground(currentTime float32, window *WaylandWindow) {
	// Background fade-in duration
	fadeDuration := float32(1.0)
	targetAlpha := float32(0.75) // Less translucent (was 0.6)

	var alpha float32
	if currentTime < fadeDuration {
		// Ease-out quint: 1 - (1-t)^5 (more gradual)
		progress := currentTime / fadeDuration
		easedProgress := 1.0 - (1.0-progress)*(1.0-progress)*(1.0-progress)*(1.0-progress)*(1.0-progress)
		alpha = easedProgress * targetAlpha
	} else {
		// Fade complete, stay at target alpha
		alpha = targetAlpha
	}

	// Get cursor position for flashlight effect
	x, y := window.GetCursorPos()
	width, height := window.GetSize()

	// Use alpha blending for the background
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.UseProgram(g.flashProgram)

	alphaLoc := gl.GetUniformLocation(g.flashProgram, gl.Str("alpha\x00"))
	gl.Uniform1f(alphaLoc, alpha)

	cursorPosLoc := gl.GetUniformLocation(g.flashProgram, gl.Str("cursorPos\x00"))
	gl.Uniform2f(cursorPosLoc, float32(x), float32(float64(height)-y)) // Flip Y for OpenGL coordinates

	resolutionLoc := gl.GetUniformLocation(g.flashProgram, gl.Str("resolution\x00"))
	gl.Uniform2f(resolutionLoc, float32(width), float32(height))

	gl.BindVertexArray(g.flashVAO)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	gl.BindVertexArray(0)

	// Restore additive blending for subsequent draws
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
}

func (g *Launcher) drawCursorGlow(window *WaylandWindow, cursorX, cursorY, currentTime float32) {
	width, height := window.GetSize()

	// Elastic ease-out animation for startup
	growDuration := float32(1.2)
	var scale float32
	if currentTime < growDuration {
		t := currentTime / growDuration
		// Elastic ease-out: slight overshoot then settle
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

	gl.UseProgram(g.cursorGlowProgram)

	// Set uniforms
	cursorPosLoc := gl.GetUniformLocation(g.cursorGlowProgram, gl.Str("cursorPos\x00"))
	gl.Uniform2f(cursorPosLoc, cursorX, cursorY)

	resolutionLoc := gl.GetUniformLocation(g.cursorGlowProgram, gl.Str("resolution\x00"))
	gl.Uniform2f(resolutionLoc, float32(width), float32(height))

	glowSizeLoc := gl.GetUniformLocation(g.cursorGlowProgram, gl.Str("glowSize\x00"))
	gl.Uniform1f(glowSizeLoc, 80.0*scale) // Glow radius animated with elastic ease-out

	timeLoc := gl.GetUniformLocation(g.cursorGlowProgram, gl.Str("time\x00"))
	gl.Uniform1f(timeLoc, currentTime)

	velocityLoc := gl.GetUniformLocation(g.cursorGlowProgram, gl.Str("velocity\x00"))
	gl.Uniform1f(velocityLoc, g.smoothVelocity)

	rotationLoc := gl.GetUniformLocation(g.cursorGlowProgram, gl.Str("rotation\x00"))
	gl.Uniform1f(rotationLoc, g.smoothRotation)

	isDrawingLoc := gl.GetUniformLocation(g.cursorGlowProgram, gl.Str("isDrawing\x00"))
	gl.Uniform1f(isDrawingLoc, g.smoothDrawing)

	// Draw quad
	gl.BindVertexArray(g.cursorGlowVAO)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	gl.BindVertexArray(0)
}

func main() {
	// Create Wayland window with layer shell
	window, err := NewWaylandWindow()
	if err != nil {
		log.Fatal("Failed to create Wayland window:", err)
	}
	defer window.Destroy()

	// Initialize launcher and OpenGL
	launcher := &Launcher{
		points:    []Point{},
		particles: []Particle{},
		isDrawing: false,
		startTime: time.Now(),
	}

	if err := launcher.initGL(); err != nil {
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

	// Initialize cursor position
	x, y := window.GetCursorPos()
	launcher.lastCursorX = float32(x)
	launcher.lastCursorY = float32(y)
	launcher.smoothRotation = 0.0
	launcher.smoothDrawing = 0.0
	launcher.smoothVelocity = 0.0

	// Main loop
	for !window.ShouldClose() {
		// Calculate delta time
		now := time.Now()
		dt := float32(now.Sub(lastTime).Seconds())
		lastTime = now

		// Poll events
		window.PollEvents()

		// Update cursor smoothing and velocity
		launcher.updateCursor(dt, window)

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
			launcher.isDrawing = true
		} else if !isPressed && wasPressed {
			launcher.isDrawing = false
		}
		wasPressed = isPressed

		// Update
		if launcher.isDrawing {
			x, y := window.GetCursorPos()
			launcher.addPoint(float32(x), float32(y))
			// Spawn magical sparkles at cursor
			launcher.spawnCursorSparkles(float32(x), float32(y))
		}

		launcher.updateParticles(dt)

		// Draw
		launcher.draw(window)

		// Swap buffers
		window.SwapBuffers()
	}
}
