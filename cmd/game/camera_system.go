package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type CameraSystem struct {
	droneFilter *ecs.Filter2[Position3, Drone]
	focus       rl.Vector3
	offset      rl.Vector3
	deadZoneXZ  float32
	deadZoneY   float32
}

func (system *CameraSystem) Initialize(game *Game) {
	system.droneFilter = ecs.NewFilter2[Position3, Drone](game.world)
	system.focus = game.camera.Target
	system.offset = rl.Vector3Subtract(game.camera.Position, game.camera.Target)
	system.deadZoneXZ = cameraFollowDeadZoneXZ
	system.deadZoneY = cameraFollowDeadZoneY
}

func (system *CameraSystem) Update(game *Game) {
	query := system.droneFilter.Query()
	defer query.Close()

	if !query.Next() {
		return
	}

	position, _ := query.Get()
	changed := false

	if position.X > system.focus.X+system.deadZoneXZ {
		system.focus.X = position.X - system.deadZoneXZ
		changed = true
	} else if position.X < system.focus.X-system.deadZoneXZ {
		system.focus.X = position.X + system.deadZoneXZ
		changed = true
	}

	if position.Z > system.focus.Z+system.deadZoneXZ {
		system.focus.Z = position.Z - system.deadZoneXZ
		changed = true
	} else if position.Z < system.focus.Z-system.deadZoneXZ {
		system.focus.Z = position.Z + system.deadZoneXZ
		changed = true
	}

	if position.Y > system.focus.Y+system.deadZoneY {
		system.focus.Y = position.Y - system.deadZoneY
		changed = true
	} else if position.Y < system.focus.Y-system.deadZoneY {
		system.focus.Y = position.Y + system.deadZoneY
		changed = true
	}

	if !changed {
		return
	}

	game.camera.Target = rl.NewVector3(system.focus.X, system.focus.Y, system.focus.Z)
	game.camera.Position = rl.Vector3Add(game.camera.Target, system.offset)
}

func (system *CameraSystem) Unload() {}
