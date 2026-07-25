package main

import (
	"math/rand"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

func TestRandomWalkSplashDroneFireOffsetStaysWithinBounds(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(1))
	offset := rl.Vector2{}
	for i := 0; i < 10_000; i++ {
		offset = randomWalkSplashDroneFireOffset(offset, rng)
		if offset.X < -splashDroneFireOffsetLimit || offset.X > splashDroneFireOffsetLimit {
			t.Fatalf("offset X = %v, want within [%v, %v]", offset.X, -splashDroneFireOffsetLimit, splashDroneFireOffsetLimit)
		}
		if offset.Y < -splashDroneFireOffsetLimit || offset.Y > splashDroneFireOffsetLimit {
			t.Fatalf("offset Z = %v, want within [%v, %v]", offset.Y, -splashDroneFireOffsetLimit, splashDroneFireOffsetLimit)
		}
	}
}

func TestSplashDroneFireStateDurationUsesConfiguredRange(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(2))
	for i := 0; i < 100; i++ {
		duration := randomSplashDroneFireStateDuration(rng)
		if duration < splashDroneFireStateDurationMin || duration > splashDroneFireStateDurationMax {
			t.Fatalf(
				"fire state duration = %v, want within [%v, %v]",
				duration,
				splashDroneFireStateDurationMin,
				splashDroneFireStateDurationMax,
			)
		}
	}
}

func TestAdvanceSplashDroneFireStateAlternatesFiringAndCooldown(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(3))
	firing, remaining := advanceSplashDroneFireState(true, 0.1, 0.1, rng)
	if firing {
		t.Fatal("fire state did not enter cooldown")
	}
	if remaining < splashDroneFireStateDurationMin || remaining > splashDroneFireStateDurationMax {
		t.Fatalf("cooldown = %v, want configured duration range", remaining)
	}

	firing, remaining = advanceSplashDroneFireState(firing, remaining, remaining, rng)
	if !firing {
		t.Fatal("fire state did not resume firing")
	}
	if remaining < splashDroneFireStateDurationMin || remaining > splashDroneFireStateDurationMax {
		t.Fatalf("firing duration = %v, want configured duration range", remaining)
	}
}

func TestSplashScreenDroneFireTargetSystemProducesTerrainTargetWhileFiring(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	chunk := newTestTerrainChunk(ChunkCoords{X: 0, Z: 0}, 0, 0, 1)
	manager := &ChunkManager{
		chunks: map[ChunkCoords]*TerrainChunk{
			chunk.Coords: chunk,
		},
	}
	mapper := ecs.NewMap3[Position3, Drone, DroneFireTargets](world)
	entity := mapper.NewEntity(
		&Position3{X: 4, Y: 3, Z: 4},
		&Drone{},
		&DroneFireTargets{},
	)
	game := &Game{world: world, chunkManager: manager}
	system := &SplashScreenDroneFireTargetSystem{
		rng: rand.New(rand.NewSource(4)),
	}
	system.Initialize(game)

	system.Update(game)

	targets := ecs.NewMap[DroneFireTargets](world).Get(entity)
	if len(targets.targets) != 1 {
		t.Fatalf("target count = %d, want 1", len(targets.targets))
	}
	target := targets.targets[0]
	if target.X < 3.5 || target.X > 4.5 || target.Z < 3.5 || target.Z > 4.5 {
		t.Fatalf("fire target = %+v, want within 0.5 X/Z of drone", target)
	}
	if target.Y != 1 {
		t.Fatalf("fire target Y = %v, want sampled terrain height 1", target.Y)
	}
}

func TestSplashScreenDroneFireTargetSystemClearsTargetsDuringCooldown(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	mapper := ecs.NewMap3[Position3, Drone, DroneFireTargets](world)
	entity := mapper.NewEntity(
		&Position3{},
		&Drone{},
		&DroneFireTargets{targets: []rl.Vector3{rl.NewVector3(1, 2, 3)}},
	)
	game := &Game{world: world, chunkManager: &ChunkManager{}}
	system := &SplashScreenDroneFireTargetSystem{
		filter:         ecs.NewFilter3[Position3, Drone, DroneFireTargets](world),
		rng:            rand.New(rand.NewSource(5)),
		firing:         false,
		stateRemaining: 1,
	}

	system.Update(game)

	targets := ecs.NewMap[DroneFireTargets](world).Get(entity)
	if len(targets.targets) != 0 {
		t.Fatalf("target count = %d, want 0 during cooldown", len(targets.targets))
	}
}
