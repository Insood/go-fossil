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

	assets := NewAssetManager()
	assets.Load()
	defer assets.Unload()

	splash := InitializeSplashScreen(assets)
	for !rl.WindowShouldClose() && !splash.StartRequested {
		rl.BeginDrawing()
		rl.ClearBackground(rl.SkyBlue)
		splash.UpdateSystems()
		rl.EndDrawing()
	}

	if rl.WindowShouldClose() {
		splash.Unload()
		return
	}

	splash.Unload()

	game := InitializeGame(assets)
	defer game.Unload()

	for !rl.WindowShouldClose() && game.Running {
		rl.BeginDrawing()
		rl.ClearBackground(rl.SkyBlue)
		game.FrameTime = rl.GetFrameTime()
		game.UpdateSystems()
		rl.EndDrawing()
	}
}
