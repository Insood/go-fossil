package main

import rl "github.com/gen2brain/raylib-go/raylib"

type DebugRender3DSystem struct{}

func (system *DebugRender3DSystem) Initialize(game *Game) {}

func (system *DebugRender3DSystem) Update(game *Game) {
	rl.BeginMode3D(game.camera)
	system.drawCoordinateSystemAt(rl.NewVector3(0, 0, 0))
	rl.EndMode3D()
}

func (system *DebugRender3DSystem) drawCoordinateSystemAt(origin rl.Vector3) {
	rl.DrawLine3D(origin, rl.Vector3Add(origin, rl.NewVector3(axisLength, 0, 0)), rl.Red)
	rl.DrawLine3D(origin, rl.Vector3Add(origin, rl.NewVector3(0, axisLength, 0)), rl.Green)
	rl.DrawLine3D(origin, rl.Vector3Add(origin, rl.NewVector3(0, 0, axisLength)), rl.Blue)
}
