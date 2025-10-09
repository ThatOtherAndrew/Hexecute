package main

import rl "github.com/gen2brain/raylib-go/raylib"

func main() {
	rl.SetConfigFlags(rl.FlagWindowTransparent)
	rl.InitWindow(500, 500, "hexecute")
	rl.SetWindowState(rl.FlagWindowUndecorated)
	defer rl.CloseWindow()

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Blank)
		rl.DrawText("Hello, World!", 10, 10, 20, rl.Red)
		rl.EndDrawing()
	}
}
