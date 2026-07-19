package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type LaserSystem struct {
	filter    *ecs.Filter5[Position3, Drone, Laser, DroneFireControl, Battery]
	damageMap *ecs.Map[TerrainChunkDamaged]
}

func (system *LaserSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter5[Position3, Drone, Laser, DroneFireControl, Battery](game.world)
	system.damageMap = ecs.NewMap[TerrainChunkDamaged](game.world)
}

func (system *LaserSystem) Update(game *Game) {
	query := system.filter.Query()
	burnTargets := make([]rl.Vector3, 0, 1)

	for query.Next() {
		position, _, laser, control, battery := query.Get()
		laser.active = false
		if !control.firing || battery.charge <= 0 {
			continue
		}

		droneTargets := make([]rl.Vector3, 0, 1)
		for _, cursor := range laserBurnCursors(*control) {
			target, ok := droneViewportWorldTarget(
				cursor,
				rl.Vector3(*position),
				game.chunkManager.SampleHeight,
				func(worldX, worldZ float32) bool {
					_, ok := game.chunkManager.ChunkForWorldPosition(worldX, worldZ)
					return ok
				},
			)
			if !ok {
				continue
			}

			droneTargets = append(droneTargets, target)
		}

		if len(droneTargets) == 0 {
			continue
		}

		laser.active = true
		laser.target = droneTargets[len(droneTargets)-1]
		drainLaserBattery(battery)
		burnTargets = append(burnTargets, droneTargets...)
	}

	query.Close()
	system.applyBurnTargets(game, burnTargets)
}

func laserBurnCursors(control DroneFireControl) []rl.Vector2 {
	if !control.lastFiring {
		return []rl.Vector2{control.cursor}
	}

	distance := rl.Vector2Distance(control.lastCursor, control.cursor)
	if distance == 0 {
		return []rl.Vector2{control.cursor}
	}

	steps := int(math.Ceil(float64(distance / laserCursorBurnStepPixels)))
	if steps < 1 {
		steps = 1
	}

	cursors := make([]rl.Vector2, 0, steps)
	for step := 1; step <= steps; step++ {
		t := float32(step) / float32(steps)
		cursors = append(cursors, rl.Vector2Lerp(control.lastCursor, control.cursor, t))
	}

	return cursors
}

func drainLaserBattery(battery *Battery) {
	battery.charge -= laserBatteryDrainPerBurn
	if battery.charge < 0 {
		battery.charge = 0
	}
}

func (system *LaserSystem) applyBurnTargets(game *Game, targets []rl.Vector3) {
	if len(targets) == 0 {
		return
	}

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
