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
