package main

import rl "github.com/gen2brain/raylib-go/raylib"

type DebugRender3DSystem struct {
	overlayActive bool
}

func (system *DebugRender3DSystem) Initialize(game *Game) {
	system.overlayActive = false
}

func (system *DebugRender3DSystem) Update(game *Game) {
	if rl.IsKeyPressed(rl.KeyF10) {
		system.overlayActive = !system.overlayActive
	}

	if !system.overlayActive {
		return
	}

	rl.BeginMode3D(game.camera)
	system.drawCoordinateSystemAt(rl.NewVector3(0, 0, 0))
	rl.EndMode3D()
}

func (system *DebugRender3DSystem) drawCoordinateSystemAt(origin rl.Vector3) {
	rl.DrawLine3D(origin, rl.Vector3Add(origin, rl.NewVector3(axisLength, 0, 0)), rl.Red)
	rl.DrawLine3D(origin, rl.Vector3Add(origin, rl.NewVector3(0, axisLength, 0)), rl.Green)
	rl.DrawLine3D(origin, rl.Vector3Add(origin, rl.NewVector3(0, 0, axisLength)), rl.Blue)
}
