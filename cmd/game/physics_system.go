package main

import (
	ecs "github.com/mlange-42/ark/ecs"
)

type PhysicsSystem struct {
	filter *ecs.Filter2[Position3, Velocity3]
}

func (system *PhysicsSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter2[Position3, Velocity3](game.world)
}

func (system *PhysicsSystem) Update(game *Game) {
	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		position, velocity := query.Get()
		position.X += velocity.X * game.FrameTime
		position.Y += velocity.Y * game.FrameTime
		position.Z += velocity.Z * game.FrameTime
	}
}

func (system *PhysicsSystem) Unload() {}
