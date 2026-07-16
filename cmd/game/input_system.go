package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type InputSystem struct {
}

func (system *InputSystem) Initialize(game *Game) {
}

func (system *InputSystem) Update(game *Game) {
	if rl.IsGamepadButtonPressed(droneGamepadIndex, gamePadQuitButton1) &&
		rl.IsGamepadButtonPressed(droneGamepadIndex, gamePadQuitButton2) {
		game.Running = false
	}
}
