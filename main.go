package main

import rl "github.com/gen2brain/raylib-go/raylib"

func main() {
	rl.SetConfigFlags(rl.FlagWindowTransparent)
	rl.InitWindow(500, 500, "hexecute")
	defer rl.CloseWindow()

	points := []rl.Vector2{}
	isDrawing := false

	println("WindowScaleDPI:", rl.GetWindowScaleDPI().X, rl.GetWindowScaleDPI().Y)
	println("Monitor:", rl.GetMonitorWidth(0), rl.GetMonitorHeight(0))
	println("Screen:", rl.GetScreenWidth(), rl.GetScreenHeight())
	println("Render:", rl.GetRenderWidth(), rl.GetRenderHeight())

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.EnableColorBlend()
		rl.ClearBackground(rl.Blank)
		rl.DrawRectangle(0, 0, 500, 500, rl.Fade(rl.Red, 0.5))

		isMouseDown := rl.IsMouseButtonDown(rl.MouseLeftButton)
		if isMouseDown {
			if !isDrawing { // on mouse down
				println("Start drawing")
				isDrawing = true
			}
			points = append(points, rl.GetMousePosition())
		} else if isDrawing { // on mouse up
			println("Stop drawing")
			isDrawing = false
			points = []rl.Vector2{}
		}

		for i := 1; i < len(points)-1; i++ {
			rl.DrawLineEx(points[i], points[i+1], 5, rl.Magenta)
		}

		rl.EndDrawing()
	}
}
