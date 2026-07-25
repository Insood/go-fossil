package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type PlayerDroneFireControlSystem struct {
	filter *ecs.Filter2[PlayerFireInput, PlayerControlled]
}

func (system *PlayerDroneFireControlSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter2[PlayerFireInput, PlayerControlled](game.world)
	rl.HideCursor()
}

func (system *PlayerDroneFireControlSystem) Update(game *Game) {
	if debugOverlayVisible {
		rl.ShowCursor()
		system.clearFireControlState()
		return
	}

	rl.HideCursor()

	viewport := droneViewportRectangle()
	mouse := rl.GetMousePosition()
	cursor := droneViewportCursorFromMouse(mouse, viewport)

	if rl.IsGamepadAvailable(droneGamepadIndex) {
		axisX := rl.GetGamepadAxisMovement(droneGamepadIndex, droneGamepadTargetAxisX)
		axisZ := rl.GetGamepadAxisMovement(droneGamepadIndex, droneGamepadTargetAxisZ)
		cursor, _ = playerDroneFireControlCursor(mouse, axisX, axisZ, viewport)
	}

	cursorPixel := droneViewportCursorPixel(cursor, viewport)
	rl.SetMousePosition(int(cursorPixel.X), int(cursorPixel.Y))

	firing := rl.IsMouseButtonDown(rl.MouseLeftButton)
	if rl.IsGamepadButtonDown(droneGamepadIndex, rl.GamepadButtonRightTrigger2) {
		firing = true
	}

	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		control, _ := query.Get()
		control.lastCursor = control.cursor
		control.lastFiring = control.firing
		control.cursor = cursor
		control.firing = firing
	}
}

func (system *PlayerDroneFireControlSystem) Unload() {}

func (system *PlayerDroneFireControlSystem) clearFireControlState() {
	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		control, _ := query.Get()
		control.lastCursor = control.cursor
		control.lastFiring = control.firing
		control.firing = false
	}
}

func playerDroneFireControlCursor(mouse rl.Vector2, axisX, axisZ float32, viewport rl.Rectangle) (rl.Vector2, bool) {
	if axisX != 0 || axisZ != 0 {
		return droneViewportCursorFromGamepad(viewport, axisX, axisZ), true
	}

	return droneViewportCursorFromMouse(mouse, viewport), false
}
