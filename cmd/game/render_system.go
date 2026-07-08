package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type RenderSystem3D struct {
	filter *ecs.Filter2[Position3, Renderable]
}

func (system *RenderSystem3D) Initialize(game *Game) {
	system.filter = ecs.NewFilter2[Position3, Renderable](game.world)
}

func (system *RenderSystem3D) Update(game *Game) {
	rl.BeginMode3D(game.camera)
	system.renderModels()
	rl.EndMode3D()
}

func (system *RenderSystem3D) renderModels() {
	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		position, renderable := query.Get()
		rl.DrawModel(*renderable.model, rl.Vector3(*position), renderable.scale, renderable.tint)
	}
}
