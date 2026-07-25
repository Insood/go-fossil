package main

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

func TestPlayerFireCursorsUsesCurrentCursorWhenPreviousFrameWasNotFiring(t *testing.T) {
	t.Parallel()

	cursors := playerFireCursors(PlayerFireInput{
		cursor:     rl.NewVector2(0.2, 0.3),
		firing:     true,
		lastCursor: rl.NewVector2(0.1, 0.3),
		lastFiring: false,
	})

	if len(cursors) != 1 {
		t.Fatalf("len(cursors) = %d, want 1", len(cursors))
	}
	if cursors[0].X != 0.2 || cursors[0].Y != 0.3 {
		t.Fatalf("cursor = (%.2f, %.2f), want (0.20, 0.30)", cursors[0].X, cursors[0].Y)
	}
}

func TestPlayerFireCursorsInterpolatesWithConfiguredMaximumStep(t *testing.T) {
	t.Parallel()

	cursors := playerFireCursors(PlayerFireInput{
		cursor:     rl.NewVector2(12.0/float32(droneViewPixels)*2, 0),
		firing:     true,
		lastCursor: rl.NewVector2(0, 0),
		lastFiring: true,
	})

	if len(cursors) != 3 {
		t.Fatalf("len(cursors) = %d, want 3", len(cursors))
	}

	wantX := []float32{
		4.0 / float32(droneViewPixels) * 2,
		8.0 / float32(droneViewPixels) * 2,
		12.0 / float32(droneViewPixels) * 2,
	}
	for i, cursor := range cursors {
		if cursor.X != wantX[i] || cursor.Y != 0 {
			t.Fatalf("cursor %d = (%.4f, %.4f), want (%.4f, 0.0000)", i, cursor.X, cursor.Y, wantX[i])
		}
	}
}

func TestPlayerFireCursorsKeepsSmallMotionToSingleCurrentCursor(t *testing.T) {
	t.Parallel()

	cursors := playerFireCursors(PlayerFireInput{
		cursor:     rl.NewVector2(4.0/float32(droneViewPixels)*2, 0),
		firing:     true,
		lastCursor: rl.NewVector2(0, 0),
		lastFiring: true,
	})

	if len(cursors) != 1 {
		t.Fatalf("len(cursors) = %d, want 1", len(cursors))
	}
	wantX := 4.0 / float32(droneViewPixels) * 2
	if cursors[0].X != wantX || cursors[0].Y != 0 {
		t.Fatalf("cursor = (%.4f, %.4f), want (%.4f, 0.0000)", cursors[0].X, cursors[0].Y, wantX)
	}
}

func TestPlayerDroneFireTargetSystemClearsTargetsWhileNotFiring(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	mapper := ecs.NewMap5[Position3, Drone, PlayerFireInput, DroneFireTargets, PlayerControlled](world)
	entity := mapper.NewEntity(
		&Position3{},
		&Drone{},
		&PlayerFireInput{},
		&DroneFireTargets{targets: []rl.Vector3{rl.NewVector3(1, 2, 3)}},
		&PlayerControlled{},
	)
	game := &Game{world: world}
	system := &PlayerDroneFireTargetSystem{}
	system.Initialize(game)

	system.Update(game)

	targets := ecs.NewMap[DroneFireTargets](world).Get(entity)
	if len(targets.targets) != 0 {
		t.Fatalf("target count = %d, want 0", len(targets.targets))
	}
}

func TestPlayerDroneFireTargetSystemSkipsInvalidCursorTargets(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	mapper := ecs.NewMap5[Position3, Drone, PlayerFireInput, DroneFireTargets, PlayerControlled](world)
	entity := mapper.NewEntity(
		&Position3{},
		&Drone{},
		&PlayerFireInput{cursor: rl.NewVector2(2, 0), firing: true},
		&DroneFireTargets{targets: []rl.Vector3{rl.NewVector3(1, 2, 3)}},
		&PlayerControlled{},
	)
	game := &Game{world: world, chunkManager: &ChunkManager{}}
	system := &PlayerDroneFireTargetSystem{}
	system.Initialize(game)

	system.Update(game)

	targets := ecs.NewMap[DroneFireTargets](world).Get(entity)
	if len(targets.targets) != 0 {
		t.Fatalf("target count = %d, want 0", len(targets.targets))
	}
}

func TestPlayerDroneFireTargetSystemConvertsCursorToWorldTarget(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	chunk := newTestTerrainChunk(ChunkCoords{X: 0, Z: 0}, 0, 0, 0)
	manager := &ChunkManager{
		chunks: map[ChunkCoords]*TerrainChunk{
			chunk.Coords: chunk,
		},
	}
	mapper := ecs.NewMap5[Position3, Drone, PlayerFireInput, DroneFireTargets, PlayerControlled](world)
	entity := mapper.NewEntity(
		&Position3{X: 4, Y: droneCenterY, Z: 4},
		&Drone{},
		&PlayerFireInput{cursor: rl.NewVector2(0, 0), firing: true},
		&DroneFireTargets{},
		&PlayerControlled{},
	)
	game := &Game{world: world, chunkManager: manager}
	system := &PlayerDroneFireTargetSystem{}
	system.Initialize(game)

	system.Update(game)

	targets := ecs.NewMap[DroneFireTargets](world).Get(entity)
	if len(targets.targets) != 1 {
		t.Fatalf("target count = %d, want 1", len(targets.targets))
	}
	target := targets.targets[0]
	if target.X != 4 || target.Z != 4 || target.Y < -0.001 || target.Y > 0.001 {
		t.Fatalf("world target = %+v, want terrain near (4, 0, 4)", target)
	}
}
