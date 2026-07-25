package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"

	"go-fossil/internal/terrain"
)

func spawnChargingPadChunkEntity(
	manager *ChunkManager,
	chunk *TerrainChunk,
	placement terrain.EntityPlacement,
) (ecs.Entity, error) {
	model, ok := manager.assets.LookupModel(chargingPadModelName)
	if !ok || model == nil {
		return ecs.Entity{}, fmt.Errorf("missing model %q", chargingPadModelName)
	}
	if manager.chargingPadMap == nil {
		manager.chargingPadMap = ecs.NewMap3[Position3, Renderable, ChargingPad](manager.world)
	}

	position := chunkEntityPlacementPosition(chunk, placement)
	return manager.chargingPadMap.NewEntity(
		&Position3{X: position.X, Y: position.Y, Z: position.Z},
		&Renderable{
			model:          model,
			scale:          1,
			tint:           rl.White,
			castsShadow:    true,
			receivesShadow: true,
		},
		&ChargingPad{},
	), nil
}

func nearestChargingPadPosition(
	filter *ecs.Filter2[Position3, ChargingPad],
	reference rl.Vector3,
) (rl.Vector3, bool) {
	if filter == nil {
		return rl.Vector3{}, false
	}

	query := filter.Query()
	defer query.Close()

	nearest := rl.Vector3{}
	nearestDistance := float32(0)
	found := false
	for query.Next() {
		position, _ := query.Get()
		candidate := rl.Vector3(*position)
		distance := rl.Vector2Distance(xzVector(reference), xzVector(candidate))
		if found && distance >= nearestDistance {
			continue
		}

		nearest = candidate
		nearestDistance = distance
		found = true
	}

	return nearest, found
}
