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

	tutorialCompleted := false
	for !rl.WindowShouldClose() {
		splash := InitializeSplashScreen(assets)
		for !rl.WindowShouldClose() && !splash.StartRequested {
			rl.BeginDrawing()
			rl.ClearBackground(rl.SkyBlue)
			splash.FrameTime = rl.GetFrameTime()
			splash.UpdateSystems()
			rl.EndDrawing()
		}
		splash.Unload()
		if rl.WindowShouldClose() {
			return
		}

		game := InitializeGame(assets, tutorialCompleted)
		for !rl.WindowShouldClose() && game.Running {
			rl.BeginDrawing()
			rl.ClearBackground(rl.SkyBlue)
			game.FrameTime = rl.GetFrameTime()
			game.UpdateSystems()
			rl.EndDrawing()
		}

		returnToMenu := game.ReturnToMenuRequested
		tutorialCompleted = game.TutorialCompleted
		game.Unload()
		if rl.WindowShouldClose() || !returnToMenu {
			return
		}
	}
}
