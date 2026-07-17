package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type DroneFireControlSystem struct {
	filter *ecs.Filter1[DroneFireControl]
}

func (system *DroneFireControlSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter1[DroneFireControl](game.world)
	rl.HideCursor()
}

func (system *DroneFireControlSystem) Update(game *Game) {
	viewport := droneViewportRectangle()
	mouse := rl.GetMousePosition()
	cursor := droneViewportCursorFromMouse(mouse, viewport)

	if rl.IsGamepadAvailable(droneGamepadIndex) {
		axisX := rl.GetGamepadAxisMovement(droneGamepadIndex, droneGamepadTargetAxisX)
		axisZ := rl.GetGamepadAxisMovement(droneGamepadIndex, droneGamepadTargetAxisZ)
		cursor, _ = droneFireControlCursor(mouse, axisX, axisZ, viewport)
	}

	rl.SetMousePosition(int(cursor.X), int(cursor.Y))

	firing := rl.IsMouseButtonDown(rl.MouseLeftButton)
	if rl.IsGamepadButtonDown(droneGamepadIndex, rl.GamepadButtonRightTrigger2) {
		firing = true
	}

	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		control := query.Get()
		control.cursor = cursor
		control.firing = firing
	}
}

func droneFireControlCursor(mouse rl.Vector2, axisX, axisZ float32, viewport rl.Rectangle) (rl.Vector2, bool) {
	if axisX != 0 || axisZ != 0 {
		return droneViewportCursorFromGamepad(viewport, axisX, axisZ), true
	}

	return droneViewportCursorFromMouse(mouse, viewport), false
}
