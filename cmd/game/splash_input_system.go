package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type SplashInputSystem struct{}

func (system *SplashInputSystem) Initialize(screen *SplashScreen) {}

func (system *SplashInputSystem) Update(screen *SplashScreen) {
	spacePressed := rl.IsKeyPressed(rl.KeySpace)
	gamepadAPressed := rl.IsGamepadAvailable(droneGamepadIndex) &&
		rl.IsGamepadButtonPressed(droneGamepadIndex, rl.GamepadButtonRightFaceDown)
	screen.StartRequested = splashStartRequested(spacePressed, gamepadAPressed)
}

func (system *SplashInputSystem) Unload() {}

func splashStartRequested(spacePressed, gamepadAPressed bool) bool {
	return spacePressed || gamepadAPressed
}
