package main

import (
	ecs "github.com/mlange-42/ark/ecs"
)

type TerrainCollisionDetectionSystem struct {
	filter *ecs.Filter2[Position3, Velocity3]
}

func (system *TerrainCollisionDetectionSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter2[Position3, Velocity3](game.world)
}

func (system *TerrainCollisionDetectionSystem) Update(game *Game) {
	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		position, velocity := query.Get()
		if velocity.Y == 0 {
			continue
		}

		resolveTerrainCollision(
			position,
			velocity,
			game.chunkManager.SampleHeight(position.X, position.Z),
		)
	}
}

func (system *TerrainCollisionDetectionSystem) Unload() {}

func resolveTerrainCollision(position *Position3, velocity *Velocity3, terrainHeight float32) bool {
	if position.Y >= terrainHeight {
		return false
	}

	position.Y = terrainHeight
	velocity.Y = 0
	return true
}
