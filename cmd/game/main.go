package main

import rl "github.com/gen2brain/raylib-go/raylib"

func main() {
	rl.InitWindow(screenWidth, screenHeight, windowTitle)
	defer rl.CloseWindow()

	game := InitializeGame()
	defer game.UnloadAssets()

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.SkyBlue)
		game.UpdateSystems()
		rl.EndDrawing()
	}
}
