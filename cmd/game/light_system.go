package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type LightSystem struct {
	lightFilter *ecs.Filter1[Light]
	droneFilter *ecs.Filter2[Position3, Drone]
}

func (system *LightSystem) Initialize(game *Game) {
	system.lightFilter = ecs.NewFilter1[Light](game.world)
	system.droneFilter = ecs.NewFilter2[Position3, Drone](game.world)
}

func (system *LightSystem) Update(game *Game) {
	droneQuery := system.droneFilter.Query()
	defer droneQuery.Close()

	if !droneQuery.Next() {
		return
	}

	position, _ := droneQuery.Get()
	dronePosition := rl.Vector3(*position)

	lightQuery := system.lightFilter.Query()
	defer lightQuery.Close()

	for lightQuery.Next() {
		light := lightQuery.Get()
		light.origin = rl.NewVector3(dronePosition.X, lightHeight, dronePosition.Z)
		light.target = rl.NewVector3(dronePosition.X, 0, dronePosition.Z)
		light.camera.Position = light.origin
		light.camera.Target = light.target
		light.camera.Up = light.up
		light.camera.Fovy = light.orthographicSize
		light.camera.Projection = rl.CameraOrthographic
	}
}

func (system *LightSystem) Unload() {}
