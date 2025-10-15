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
