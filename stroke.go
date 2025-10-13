// https://depts.washington.edu/acelab/proj/dollar/dollar.pdf

package main

import "math"

// Step 1

func resample(points []Point, n int) []Point {
	I := pathLength(points) / float32(n-1)
	D := float32(0)
	newPoints := []Point{points[0]}
	for i := 1; i < len(points); i++ {
		d := distance(points[i-1], points[i])
		if D+d >= I {
			qx := points[i-1].X + ((I-D)/d)*(points[i].X-points[i-1].X)
			qy := points[i-1].Y + ((I-D)/d)*(points[i].Y-points[i-1].Y)
			q := Point{X: qx, Y: qy}
			newPoints = append(newPoints, q)
			points = append(points[:i], append([]Point{q}, points[i:]...)...)
			D = 0
		} else {
			D += d
		}
	}
	for len(newPoints) < n {
		newPoints = append(newPoints, points[len(points)-1])
	}
	return newPoints
}

func pathLength(A []Point) float32 {
	d := float32(0)
	for i := 1; i < len(A); i++ {
		d += distance(A[i-1], A[i])
	}
	return d
}

func distance(a Point, b Point) float32 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return dx*dx + dy*dy
}

// Step 2

func rotateBy(points []Point, angle float64) []Point {
	c := centroid(points)
	for _, p := range points {
		qx := float64(p.X-c.X)*math.Cos(angle) - float64(p.Y-c.Y)*math.Sin(angle) + float64(c.X)
		qy := float64(p.X-c.X)*math.Sin(angle) + float64(p.Y-c.Y)*math.Cos(angle) + float64(c.Y)
		p.X = float32(qx)
		p.Y = float32(qy)
	}
	return points
}

// Step 3

func scaleTo(points []Point, size float32) []Point {
	B := boundingBox(points)
	for i := range points {
		p := &points[i]
		p.X = p.X * (size / B.width)
		p.Y = p.Y * (size / B.height)
	}
	return points
}

func translateTo(points []Point, k Point) []Point {
	c := centroid(points)
	for i := range points {
		p := &points[i]
		p.X += k.X - c.X
		p.Y += k.Y - c.Y
	}
	return points
}

func boundingBox(points []Point) struct{ width, height float32 } {
	minX, minY := points[0].X, points[0].Y
	maxX, maxY := points[0].X, points[0].Y
	for _, p := range points {
		if p.X < minX {
			minX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	return struct{ width, height float32 }{width: maxX - minX, height: maxY - minY}
}

func centroid(points []Point) Point {
	var x, y float32
	for _, p := range points {
		x += p.X
		y += p.Y
	}
	n := float32(len(points))
	return Point{X: x / n, Y: y / n}
}

// Step 4

func recognise(points []Point, templates [][]Point, size float64) (bestMatch int, bestScore float64) {
	b := math.Inf(1)
	for i, T := range templates {
		d := distanceAtBestAngle(points, T, -math.Pi/4, math.Pi/4, math.Pi/90)
		if d < b {
			b = d
			bestMatch = i
		}
	}
	bestScore = 1 - b/(0.5*math.Sqrt(2*size*size))
	return bestMatch, bestScore
}

func distanceAtBestAngle(points, T []Point, a, b, delta float64) float64 {
	x1 := math.Phi*a + (1-math.Phi)*b
	f1 := distanceAtAngle(points, T, x1)
	x2 := (1-math.Phi)*a + math.Phi*b
	f2 := distanceAtAngle(points, T, x2)
	for math.Abs(b-a) > delta {
		if f1 < f2 {
			b = x2
			x2 = x1
			f2 = f1
			x1 = math.Phi*a + (1-math.Phi)*b
			f1 = distanceAtAngle(points, T, x1)
		} else {
			a = x1
			x1 = x2
			f1 = f2
			x2 = (1-math.Phi)*a + math.Phi*b
			f2 = distanceAtAngle(points, T, x2)
		}
	}
	return math.Min(f1, f2)
}

func distanceAtAngle(points, T []Point, angle float64) float64 {
	newPoints := rotateBy(points, angle)
	d := pathDistance(newPoints, T)
	return d
}

func pathDistance(A, B []Point) float64 {
	d := float64(0)
	for i := range A {
		d += math.Sqrt(float64(distance(A[i], B[i])))
	}
	return d / float64(len(A))
}

// Entry points

const n = 64
const size = 250.

func ProcessStroke(points []Point) []Point {
	// Step 1
	points = resample(points, n)
	// Step 3 (skipping rotation)
	points = scaleTo(points, size)
	points = translateTo(points, Point{X: 0, Y: 0})

	return points
}

func UnistrokeRecognise(points []Point, templates [][]Point) (bestMatch int, bestScore float64) {
	return recognise(points, templates, size)
}
