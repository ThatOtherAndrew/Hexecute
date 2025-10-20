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
	float hue = mod(vPosition.x * 0.001 + vPosition.y * 0.001 + time * 0.5, 1.0);
	vec3 color = hsv2rgb(vec3(hue, 0.8, 1.0));

	float sparkle = sin(vPosition.x * 0.1 + time * 3.0) * sin(vPosition.y * 0.1 + time * 2.0);
	sparkle = smoothstep(0.7, 1.0, sparkle) * 0.5;

	FragColor = vec4(color * (1.0 + sparkle * 2.0), vAlpha);
}
