package main

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	screenWidth  = 1280
	screenHeight = 720
	windowTitle  = "go-fossil"
)

func main() {
	rl.InitWindow(screenWidth, screenHeight, windowTitle)
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)
		rl.DrawText("hello world", 40, 40, 32, rl.DarkGray)

		rl.EndDrawing()
	}
}
