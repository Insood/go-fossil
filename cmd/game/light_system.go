package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type LightSystem struct {
	lightFilter *ecs.Filter1[Light]
}

func (system *LightSystem) Initialize(game *Game) {
	system.lightFilter = ecs.NewFilter1[Light](game.world)
}

func (system *LightSystem) Update(game *Game) {
	query := system.lightFilter.Query()
	defer query.Close()

	for query.Next() {
		light := query.Get()
		light.camera.Position = light.origin
		light.camera.Target = light.target
		light.camera.Up = light.up
		light.camera.Fovy = light.orthographicSize
		light.camera.Projection = rl.CameraOrthographic
	}
}
