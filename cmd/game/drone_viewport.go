package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func droneViewportRectangle() rl.Rectangle {
	panelSize := float32(droneViewPixels)
	margin := float32(droneViewMargin)
	return rl.NewRectangle(
		float32(screenWidth)-panelSize-margin,
		float32(screenHeight)-panelSize-margin,
		panelSize,
		panelSize,
	)
}

func droneCameraForPosition(dronePosition rl.Vector3) rl.Camera3D {
	cameraPosition := rl.NewVector3(dronePosition.X, dronePosition.Y-droneHeight/2, dronePosition.Z)
	cameraTarget := rl.NewVector3(cameraPosition.X, cameraPosition.Y-1, cameraPosition.Z)

	return rl.NewCamera3D(
		cameraPosition,
		cameraTarget,
		rl.NewVector3(0, 0, -1),
		droneViewSizeWorld,
		rl.CameraOrthographic,
	)
}

func droneViewportWorldTarget(mouse rl.Vector2, dronePosition rl.Vector3, sampleHeight func(worldX, worldZ float32) float32) (rl.Vector3, bool) {
	viewport := droneViewportRectangle()
	if !rl.CheckCollisionPointRec(mouse, viewport) {
		return rl.Vector3{}, false
	}

	u := (mouse.X - viewport.X) / viewport.Width
	v := (mouse.Y - viewport.Y) / viewport.Height
	worldX := dronePosition.X + (u-0.5)*droneViewSizeWorld
	worldZ := dronePosition.Z + (v-0.5)*droneViewSizeWorld
	worldY := sampleHeight(worldX, worldZ)

	return rl.NewVector3(worldX, worldY, worldZ), true
}

func droneLaserEmitterPosition(dronePosition rl.Vector3) rl.Vector3 {
	return rl.NewVector3(
		dronePosition.X-droneWidth/2,
		dronePosition.Y+droneHeight/2,
		dronePosition.Z-droneDepth/2,
	)
}
