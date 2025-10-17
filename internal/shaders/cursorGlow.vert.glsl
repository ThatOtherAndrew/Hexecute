#version 410 core
layout (location = 0) in vec2 position;

uniform vec2 cursorPos;
uniform vec2 resolution;
uniform float glowSize;
uniform float rotation;

out vec2 vTexCoord;

void main() {
	float c = cos(rotation);
	float s = sin(rotation);
	vec2 rotatedPos = vec2(position.x * c - position.y * s, position.x * s + position.y * c);
	vec2 worldPos = cursorPos + rotatedPos * glowSize;
	vec2 normalized = (worldPos / resolution) * 2.0 - 1.0;
	normalized.y = -normalized.y;
	gl_Position = vec4(normalized, 0.0, 1.0);
	vTexCoord = rotatedPos * 0.5 + 0.5;
}
