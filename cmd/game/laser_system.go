package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type LaserSystem struct {
	filter    *ecs.Filter3[Position3, Drone, Laser]
	damageMap *ecs.Map[TerrainChunkDamaged]
}

func (system *LaserSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter3[Position3, Drone, Laser](game.world)
	system.damageMap = ecs.NewMap[TerrainChunkDamaged](game.world)
}

func (system *LaserSystem) Update(game *Game) {
	mouse := rl.GetMousePosition()
	query := system.filter.Query()
	defer query.Close()
	burnTargets := make([]rl.Vector3, 0, 1)

	for query.Next() {
		position, _, laser := query.Get()
		target, ok := droneViewportWorldTarget(
			mouse,
			rl.Vector3(*position),
			game.chunkManager.SampleHeight,
			func(worldX, worldZ float32) bool {
				_, ok := game.chunkManager.ChunkForWorldPosition(worldX, worldZ)
				return ok
			},
		)
		if !rl.IsMouseButtonDown(rl.MouseLeftButton) || !ok {
			laser.active = false
			continue
		}

		laser.active = true
		laser.target = target
		burnTargets = append(burnTargets, target)
	}

	system.applyBurnTargets(game, burnTargets)
}

func (system *LaserSystem) applyBurnTargets(game *Game, targets []rl.Vector3) {
	for _, target := range targets {
		chunk, ok := game.chunkManager.ChunkForWorldPosition(target.X, target.Z)
		if !ok {
			panic(fmt.Sprintf("laser target %v is not inside a loaded chunk", target))
		}
		if !game.chunkManager.BurnAtWorldPosition(target.X, target.Z) {
			panic(fmt.Sprintf("failed to burn terrain at %v", target))
		}
		if system.damageMap.Has(chunk.Entity) {
			continue
		}

		system.damageMap.Add(chunk.Entity, &TerrainChunkDamaged{})
	}
}
