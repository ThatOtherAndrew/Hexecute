package shaders

const CursorGlowFragmentShader = `
#version 410 core
in vec2 vTexCoord;
out vec4 FragColor;

uniform float time;
uniform float velocity;
uniform float isDrawing;
uniform float exitProgress;

vec3 hsv2rgb(vec3 c) {
	vec4 K = vec4(1.0, 2.0 / 3.0, 1.0 / 3.0, 3.0);
	vec3 p = abs(fract(c.xxx + K.xyz) * 6.0 - K.www);
	return c.z * mix(K.xxx, clamp(p - K.xxx, 0.0, 1.0), c.y);
}

float smin(float a, float b, float k) {
	float h = clamp(0.5 + 0.5 * (b - a) / k, 0.0, 1.0);
	return mix(b, a, h) - k * h * (1.0 - h);
}

float hash(vec2 p) {
	p = fract(p * vec2(123.34, 456.21));
	p += dot(p, p + 45.32);
	return fract(p.x * p.y);
}

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
	vec2 coord = vTexCoord * 2.0 - 1.0;
	float velocityNorm = clamp(velocity * 0.01, 0.0, 1.0);
	float energy = mix(0.3, 1.0, velocityNorm) + isDrawing * 0.7;

	float sdf = 1000.0;
	float centralSize = mix(0.15, 0.35, velocityNorm) + isDrawing * 0.05;
	float pulseSpeed = (3.0 + velocityNorm * 2.0) * (1.0 + isDrawing * 0.75);
	float pulseAmount = (0.1 * energy + isDrawing * 0.075);
	float pulse = sin(time * pulseSpeed) * pulseAmount + 0.9;
	float centralDist = length(coord) - centralSize * pulse;
	sdf = centralDist;

	float numBlobsFloat = mix(5.0, 9.0, velocityNorm) + isDrawing * 1.0;
	int numBlobs = int(numBlobsFloat);
	float blobFade = fract(numBlobsFloat);

	for(int i = 0; i < 10; i++) {
		if(i > numBlobs) break;
		if(i == numBlobs && blobFade < 0.01) break;
		float baseRotation = time * 0.8;
		float velocityRotation = time * velocityNorm * 0.4;
		float angle = float(i) * 6.28 / float(numBlobs) + baseRotation + velocityRotation;

		float baseRadius = mix(0.2, 0.5, velocityNorm) + isDrawing * 0.075;
		float radiusVariation = sin(time * (1.5 + isDrawing * 0.5) + float(i) * 0.8) * mix(0.05, 0.15, velocityNorm);
		float chaoticRadius = sin(time * 4.0 + float(i) * 2.1) * cos(time * 3.5 + float(i) * 1.7) * 0.003 * isDrawing;
		float radius = baseRadius + radiusVariation + chaoticRadius;
		vec2 orbPos = vec2(cos(angle), sin(angle)) * radius;

		float baseBlobSize = mix(0.08, 0.18, velocityNorm) + isDrawing * 0.04;
		float sizeVariation = sin(time * (2.5 + isDrawing * 1.0) + float(i) * 0.6) * mix(0.02, 0.05, velocityNorm);
		float drawingGrowth = sin(time * 5.0 + float(i) * 1.3) * 0.03 * isDrawing;
		float blobSize = baseBlobSize + sizeVariation + drawingGrowth;
		float blobDist = length(coord - orbPos) - blobSize;

		if(i == numBlobs) {
			blobDist += (1.0 - blobFade) * 0.5;
		}
		float blendAmount = mix(0.15, 0.3, velocityNorm) + isDrawing * 0.075;
		sdf = smin(sdf, blobDist, blendAmount);
	}

	float noiseZoom = 3.0 + isDrawing * 0.5;
	vec2 noiseCoord = coord * noiseZoom;
	noiseCoord += vec2(time * 0.3, time * 0.2);
	float swirl = fbm(noiseCoord) * 2.0 - 1.0;

	sdf += swirl * (0.1 * energy + exitProgress * 0.8);
	float intensity = exp(-max(sdf, 0.0) * 4.0);
	float outerGlow = exp(-max(sdf, 0.0) * 1.5) * 0.4 * energy;
	float innerGlow = exp(-max(sdf, 0.0) * 8.0) * 0.8;

	float totalIntensity = intensity + outerGlow + innerGlow;

	totalIntensity *= smoothstep(1.0, 0.7, max(abs(coord.x), abs(coord.y)));

	float hueSpeed = mix(0.2, 0.6, velocityNorm);
	float hue = mod(time * hueSpeed + atan(coord.y, coord.x) / 6.28 + swirl * 0.3, 1.0);
	vec3 mainColor = hsv2rgb(vec3(hue, mix(0.7, 0.75, velocityNorm), 1.0));
	vec3 accentColor = hsv2rgb(vec3(mod(hue + 0.5, 1.0), 0.75, 1.2));
	vec3 finalColor = mainColor * intensity;
	finalColor += accentColor * innerGlow;
	finalColor += mainColor * 0.5 * outerGlow;

	float sparkle = smoothstep(0.85, 1.0, noise(coord * 20.0 + time * 5.0 * energy)) * totalIntensity * velocityNorm;
	finalColor += sparkle;

	float edge = smoothstep(0.05, -0.05, sdf) - smoothstep(0.15, 0.05, sdf);
	finalColor += accentColor * edge * energy;

	finalColor *= sin(time * (2.5 + isDrawing * 0.75)) * (0.1 + velocityNorm * 0.1 + isDrawing * 0.075) + 0.9;

	float alpha = clamp(totalIntensity * mix(0.8, 1.3, velocityNorm), 0.0, 1.0) * (1.0 - exitProgress);

	FragColor = vec4(finalColor, alpha * 0.95);
}
` + "\x00"
