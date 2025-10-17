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
	vec2 coord = gl_PointCoord - vec2(0.5);
	float dist = length(coord);
	if (dist > 0.5) discard;

	float alpha = smoothstep(0.5, 0.2, dist) * vLife;
	vec3 color = hsv2rgb(vec3(vHue, 0.9, 1.0)) * (1.0 + (1.0 - dist * 2.0) * 2.0);

	FragColor = vec4(color, alpha * 0.8);
}
