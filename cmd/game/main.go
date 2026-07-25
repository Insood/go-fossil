package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.InitWindow(screenWidth, screenHeight, windowTitle)
	defer rl.CloseWindow()

	rl.InitAudioDevice()
	defer rl.CloseAudioDevice()

	game := InitializeGame()
	defer game.UnloadAssets()

	for !rl.WindowShouldClose() && game.Running {
		rl.BeginDrawing()
		rl.ClearBackground(rl.SkyBlue)
		game.FrameTime = rl.GetFrameTime()
		game.UpdateSystems()
		rl.EndDrawing()
	}
}
