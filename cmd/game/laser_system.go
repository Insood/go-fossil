package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type LaserSystem struct {
	filter *ecs.Filter3[Position3, Drone, Laser]
}

func (system *LaserSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter3[Position3, Drone, Laser](game.world)
}

func (system *LaserSystem) Update(game *Game) {
	surface := game.primaryChunk().SurfaceMesh
	mouse := rl.GetMousePosition()
	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		position, _, laser := query.Get()
		target, ok := droneViewportWorldTarget(mouse, rl.Vector3(*position), surface)
		if !rl.IsMouseButtonDown(rl.MouseLeftButton) || !ok {
			laser.active = false
			continue
		}

		laser.active = true
		laser.target = target
	}
}
