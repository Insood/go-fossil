package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type PlayerDroneFireTargetSystem struct {
	filter *ecs.Filter5[Position3, Drone, PlayerFireInput, DroneFireTargets, PlayerControlled]
}

func (system *PlayerDroneFireTargetSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter5[Position3, Drone, PlayerFireInput, DroneFireTargets, PlayerControlled](game.world)
}

func (system *PlayerDroneFireTargetSystem) Update(game *Game) {
	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		position, _, input, fireTargets, _ := query.Get()
		fireTargets.targets = fireTargets.targets[:0]
		if !input.firing {
			continue
		}

		for _, cursor := range playerFireCursors(*input) {
			target, ok := droneViewportWorldTarget(
				cursor,
				rl.Vector3(*position),
				game.chunkManager.SampleHeight,
				func(worldX, worldZ float32) bool {
					_, ok := game.chunkManager.ChunkForWorldPosition(worldX, worldZ)
					return ok
				},
			)
			if ok {
				fireTargets.targets = append(fireTargets.targets, target)
			}
		}
	}
}

func (system *PlayerDroneFireTargetSystem) Unload() {}

func playerFireCursors(input PlayerFireInput) []rl.Vector2 {
	if !input.lastFiring {
		return []rl.Vector2{input.cursor}
	}

	pixelDistance := normalizedDroneViewportDistancePixels(input.lastCursor, input.cursor)
	if pixelDistance == 0 {
		return []rl.Vector2{input.cursor}
	}

	steps := int(math.Ceil(float64(pixelDistance / laserCursorBurnStepPixels)))
	if steps < 1 {
		steps = 1
	}

	cursors := make([]rl.Vector2, 0, steps)
	for step := 1; step <= steps; step++ {
		t := float32(step) / float32(steps)
		cursors = append(cursors, rl.Vector2Lerp(input.lastCursor, input.cursor, t))
	}

	return cursors
}

func normalizedDroneViewportDistancePixels(left, right rl.Vector2) float32 {
	return rl.Vector2Distance(left, right) * float32(droneViewPixels) * 0.5
}
