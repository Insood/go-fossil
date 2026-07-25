package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.InitWindow(screenWidth, screenHeight, windowTitle)
	defer rl.CloseWindow()

	rl.InitAudioDevice()
	defer rl.CloseAudioDevice()

	rl.SetTargetFPS(targetFPS)

	splash := InitializeSplashScreen()
	for !rl.WindowShouldClose() && !splash.StartRequested {
		rl.BeginDrawing()
		rl.ClearBackground(rl.SkyBlue)
		splash.UpdateSystems()
		rl.EndDrawing()
	}
	splash.Unload()

	if rl.WindowShouldClose() {
		return
	}

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
