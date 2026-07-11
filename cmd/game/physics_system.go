package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type PhysicsSystem struct {
	filter *ecs.Filter2[Position3, Velocity3]
}

func (system *PhysicsSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter2[Position3, Velocity3](game.world)
}

func (system *PhysicsSystem) Update(game *Game) {
	dt := rl.GetFrameTime()
	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		position, velocity := query.Get()
		position.X += velocity.X * dt
		position.Y += velocity.Y * dt
		position.Z += velocity.Z * dt
	}
}
