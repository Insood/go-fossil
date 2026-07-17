package main

import (
	"fmt"

	rg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type DebugRenderSystem2D struct {
	lightFilter *ecs.Filter1[Light]
	droneFilter *ecs.Filter2[Position3, Drone]
}

func (system *DebugRenderSystem2D) Initialize(game *Game) {
	system.lightFilter = ecs.NewFilter1[Light](game.world)
	system.droneFilter = ecs.NewFilter2[Position3, Drone](game.world)
}

func (system *DebugRenderSystem2D) Update(game *Game) {
	if rl.IsKeyPressed(rl.KeyF10) || rl.IsGamepadButtonDown(droneGamepadIndex, rl.GamepadButtonLeftTrigger2) {
		debugOverlayVisible = !debugOverlayVisible
	}

	if !debugOverlayVisible {
		return
	}

	system.drawShadowControls()
}

func (system *DebugRenderSystem2D) drawShadowControls() {
	const (
		panelWidth   = 356
		panelPadding = 12
		rowHeight    = 28
		rowGap       = 10
		labelWidth   = 96
		buttonWidth  = 28
		valueWidth   = 148
	)

	rows := 8
	panelHeight := float32(panelPadding*2 + 28 + rows*rowHeight + (rows-1)*rowGap)
	panelX := float32(screenWidth) - panelWidth - panelPadding
	panelY := float32(panelPadding)

	rg.GroupBox(rl.NewRectangle(panelX, panelY, panelWidth, panelHeight), "Shadow Tuning")

	query := system.lightFilter.Query()
	defer query.Close()

	if !query.Next() {
		return
	}

	light := query.Get()

	rowY := panelY + panelPadding + 24
	system.drawTuningRow(panelX+panelPadding, rowY, labelWidth, buttonWidth, valueWidth, "Light Size", &light.orthographicSize)
	rowY += rowHeight + rowGap
	system.drawValueRow(panelX+panelPadding, rowY, labelWidth, valueWidth, "Shadow X", light.origin.X)
	rowY += rowHeight + rowGap
	system.drawValueRow(panelX+panelPadding, rowY, labelWidth, valueWidth, "Shadow Z", light.origin.Z)
	rowY += rowHeight + rowGap
	system.drawTuningRow(panelX+panelPadding, rowY, labelWidth, buttonWidth, valueWidth, "Near Plane", &shadowNearPlane)
	rowY += rowHeight + rowGap
	system.drawTuningRow(panelX+panelPadding, rowY, labelWidth, buttonWidth, valueWidth, "Far Plane", &shadowFarPlane)
	rowY += rowHeight + rowGap
	system.drawTuningRow(panelX+panelPadding, rowY, labelWidth, buttonWidth, valueWidth, "Shadow Bias", &shadowBias)
	rowY += rowHeight + rowGap
	system.drawTuningRow(panelX+panelPadding, rowY, labelWidth, buttonWidth, valueWidth, "Slope Bias", &shadowSlopeBias)
	rowY += rowHeight + rowGap
	system.drawDroneCoordinates(panelX+panelPadding, rowY, labelWidth, valueWidth)

	if shadowNearPlane >= shadowFarPlane {
		shadowFarPlane = shadowNearPlane * 1.25
	}
}

func (system *DebugRenderSystem2D) drawDroneCoordinates(x, y, labelWidth, valueWidth float32) {
	const spacing = 8

	rg.Label(rl.NewRectangle(x, y, labelWidth, 28), "Drone Pos")

	query := system.droneFilter.Query()
	defer query.Close()

	if !query.Next() {
		rg.Label(rl.NewRectangle(x+labelWidth+spacing, y, valueWidth, 28), "n/a")
		return
	}

	position, _ := query.Get()
	rg.Label(rl.NewRectangle(x+labelWidth+spacing, y, valueWidth, 28), fmt.Sprintf("%.2f, %.2f, %.2f", position.X, position.Y, position.Z))
}

func (system *DebugRenderSystem2D) drawValueRow(x, y, labelWidth, valueWidth float32, label string, value float32) {
	const spacing = 8

	rg.Label(rl.NewRectangle(x, y, labelWidth, 28), label)
	rg.Label(rl.NewRectangle(x+labelWidth+spacing, y, valueWidth, 28), fmt.Sprintf("%.2f", value))
}

func (system *DebugRenderSystem2D) drawTuningRow(x, y, labelWidth, buttonWidth, valueWidth float32, label string, value *float32) {
	const spacing = 8

	rg.Label(rl.NewRectangle(x, y, labelWidth, 28), label)

	downX := x + labelWidth + spacing
	if rg.Button(rl.NewRectangle(downX, y, buttonWidth, 28), "<") {
		*value *= 0.75
	}

	valueX := downX + buttonWidth + spacing
	rg.Label(rl.NewRectangle(valueX, y, valueWidth, 28), fmt.Sprintf("%.6f", *value))

	upX := valueX + valueWidth + spacing
	if rg.Button(rl.NewRectangle(upX, y, buttonWidth, 28), ">") {
		*value *= 1.25
	}
}
