package main

import rl "github.com/gen2brain/raylib-go/raylib"

func main() {
	rl.SetConfigFlags(rl.FlagWindowTransparent | rl.FlagWindowHighdpi)
	rl.InitWindow(500, 500, "hexecute")
	rl.SetWindowState(rl.FlagWindowUndecorated)
	defer rl.CloseWindow()

	points := []rl.Vector2{}
	isDrawing := false

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Blank)

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
