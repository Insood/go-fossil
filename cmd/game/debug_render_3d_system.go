package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type DebugRender3DSystem struct {
	droneFilter *ecs.Filter2[Position3, Drone]
	lightFilter *ecs.Filter1[Light]
}

func (system *DebugRender3DSystem) Initialize(game *Game) {
	system.droneFilter = ecs.NewFilter2[Position3, Drone](game.world)
	system.lightFilter = ecs.NewFilter1[Light](game.world)
}

func (system *DebugRender3DSystem) Update(game *Game) {
	if !debugOverlayVisible {
		return
	}

	rl.BeginMode3D(game.camera)
	system.drawCoordinateSystemAt(rl.NewVector3(0, 0, 0))
	system.drawDroneGroundRay()
	system.drawLightGuide()
	rl.EndMode3D()
}

func (system *DebugRender3DSystem) drawCoordinateSystemAt(origin rl.Vector3) {
	rl.DrawLine3D(origin, rl.Vector3Add(origin, rl.NewVector3(debugAxisLength, 0, 0)), rl.Red)
	rl.DrawLine3D(origin, rl.Vector3Add(origin, rl.NewVector3(0, debugAxisLength, 0)), rl.Green)
	rl.DrawLine3D(origin, rl.Vector3Add(origin, rl.NewVector3(0, 0, debugAxisLength)), rl.Blue)
}

func (system *DebugRender3DSystem) drawDroneGroundRay() {
	query := system.droneFilter.Query()
	defer query.Close()

	if !query.Next() {
		return
	}

	position, _ := query.Get()
	start := rl.Vector3(*position)
	end := rl.NewVector3(start.X, 0, start.Z)

	rl.DrawLine3D(start, end, rl.Yellow)
	rl.DrawSphere(end, 0.08, rl.Orange)
}

func (system *DebugRender3DSystem) drawLightGuide() {
	query := system.lightFilter.Query()
	defer query.Close()

	if !query.Next() {
		return
	}

	light := query.Get()
	rl.DrawLine3D(light.origin, light.target, rl.SkyBlue)
	rl.DrawSphere(light.origin, 0.1, rl.Blue)
}
