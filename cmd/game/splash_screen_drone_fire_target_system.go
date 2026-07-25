package main

import (
	"math/rand"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type SplashScreenDroneFireTargetSystem struct {
	filter         *ecs.Filter3[Position3, Drone, DroneFireTargets]
	rng            *rand.Rand
	offset         rl.Vector2
	firing         bool
	stateRemaining float32
}

func (system *SplashScreenDroneFireTargetSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter3[Position3, Drone, DroneFireTargets](game.world)
	if system.rng == nil {
		system.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	system.offset = rl.Vector2{}
	system.firing = true
	system.stateRemaining = randomSplashDroneFireStateDuration(system.rng)
}

func (system *SplashScreenDroneFireTargetSystem) Update(game *Game) {
	system.offset = randomWalkSplashDroneFireOffset(system.offset, system.rng)
	system.firing, system.stateRemaining = advanceSplashDroneFireState(
		system.firing,
		system.stateRemaining,
		game.FrameTime,
		system.rng,
	)

	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		position, _, fireTargets := query.Get()
		fireTargets.targets = fireTargets.targets[:0]
		if !system.firing {
			continue
		}

		targetX := position.X + system.offset.X
		targetZ := position.Z + system.offset.Y
		if _, ok := game.chunkManager.ChunkForWorldPosition(targetX, targetZ); !ok {
			continue
		}

		fireTargets.targets = append(fireTargets.targets, rl.NewVector3(
			targetX,
			game.chunkManager.SampleHeight(targetX, targetZ),
			targetZ,
		))
	}
}

func (system *SplashScreenDroneFireTargetSystem) Unload() {}

func randomWalkSplashDroneFireOffset(offset rl.Vector2, rng *rand.Rand) rl.Vector2 {
	offset.X += randomSplashDroneFireWalkStep(rng)
	offset.Y += randomSplashDroneFireWalkStep(rng)
	offset.X = clampFloat32(offset.X, -splashDroneFireOffsetLimit, splashDroneFireOffsetLimit)
	offset.Y = clampFloat32(offset.Y, -splashDroneFireOffsetLimit, splashDroneFireOffsetLimit)
	return offset
}

func randomSplashDroneFireWalkStep(rng *rand.Rand) float32 {
	return (rng.Float32()*2 - 1) * splashDroneFireWalkStep
}

func advanceSplashDroneFireState(
	firing bool,
	remaining float32,
	dt float32,
	rng *rand.Rand,
) (bool, float32) {
	remaining -= max(dt, 0)
	for remaining <= 0 {
		firing = !firing
		remaining += randomSplashDroneFireStateDuration(rng)
	}
	return firing, remaining
}

func randomSplashDroneFireStateDuration(rng *rand.Rand) float32 {
	return splashDroneFireStateDurationMin +
		rng.Float32()*(splashDroneFireStateDurationMax-splashDroneFireStateDurationMin)
}
