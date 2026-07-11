package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type PhysicsSystem struct {
	filter      *ecs.Filter2[Position3, Velocity3]
	hoverFilter *ecs.Filter3[Position3, Velocity3, HoverMotion]
}

func (system *PhysicsSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter2[Position3, Velocity3](game.world)
	system.hoverFilter = ecs.NewFilter3[Position3, Velocity3, HoverMotion](game.world)
}

func (system *PhysicsSystem) Update(game *Game) {
	dt := rl.GetFrameTime()
	system.applyHoverMotion(dt)

	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		position, velocity := query.Get()
		position.X += velocity.X * dt
		position.Y += velocity.Y * dt
		position.Z += velocity.Z * dt
	}
}

func (system *PhysicsSystem) applyHoverMotion(dt float32) {
	if dt <= 0 {
		return
	}

	elapsed := float32(rl.GetTime())
	query := system.hoverFilter.Query()
	defer query.Close()

	for query.Next() {
		position, velocity, hover := query.Get()
		targetY := hover.baseY + hover.amplitude*float32(math.Sin(float64(elapsed*hover.angularSpeed)))
		velocity.Y = (targetY - position.Y) / dt
	}
}
