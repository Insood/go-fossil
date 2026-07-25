package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type InputSystem struct {
	gameOverFilter *ecs.Filter1[GameOver]
}

func (system *InputSystem) Initialize(game *Game) {
	system.gameOverFilter = ecs.NewFilter1[GameOver](game.world)
}

func (system *InputSystem) Update(game *Game) {
	spacePressed := rl.IsKeyPressed(rl.KeySpace)
	gamepadAPressed := rl.IsGamepadAvailable(droneGamepadIndex) &&
		rl.IsGamepadButtonPressed(droneGamepadIndex, rl.GamepadButtonRightFaceDown)
	handleGameInput(
		game,
		rl.IsGamepadButtonPressed(droneGamepadIndex, gamePadQuitButton1),
		gameOverActive(system.gameOverFilter),
		spacePressed,
		gamepadAPressed,
	)
}

func (system *InputSystem) Unload() {}

func handleGameInput(game *Game, quitPressed, gameOver, spacePressed, gamepadAPressed bool) {
	if quitPressed {
		game.Running = false
		return
	}

	if returnToMenuRequested(gameOver, spacePressed, gamepadAPressed) {
		game.ReturnToMenuRequested = true
		game.Running = false
	}
}

func returnToMenuRequested(gameOver, spacePressed, gamepadAPressed bool) bool {
	return gameOver && (spacePressed || gamepadAPressed)
}
