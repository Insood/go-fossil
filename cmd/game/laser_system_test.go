package main

import (
	"image"
	"math/rand"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"

	"go-fossil/internal/terrain"
)

func TestLaserBurnCursorsUsesCurrentCursorWhenPreviousFrameWasNotFiring(t *testing.T) {
	t.Parallel()

	cursors := laserBurnCursors(DroneFireControl{
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

func TestLaserBurnCursorsInterpolatesWithConfiguredMaximumStep(t *testing.T) {
	t.Parallel()

	cursors := laserBurnCursors(DroneFireControl{
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

func TestLaserBurnCursorsKeepsSmallMotionToSingleCurrentCursor(t *testing.T) {
	t.Parallel()

	cursors := laserBurnCursors(DroneFireControl{
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

func TestLaserSystemApplyBurnTargetsMarksChunkDamaged(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	manager := &ChunkManager{
		world:           world,
		terrainChunkMap: ecs.NewMap1[TerrainChunkComponent](world),
		chunks:          make(map[ChunkCoords]*TerrainChunk),
	}
	chunk := &TerrainChunk{
		Coords:  ChunkCoords{X: 0, Z: 0},
		OriginX: 0,
		OriginZ: 0,
		Data: terrain.ChunkData{
			Width:  4,
			Height: 4,
		},
		BurnOverlayImage: image.NewRGBA(image.Rect(0, 0, 4, 4)),
	}

	manager.registerTerrainChunkEntity(chunk)
	manager.chunks[chunk.Coords] = chunk

	system := &LaserSystem{}
	game := &Game{world: world, chunkManager: manager}
	system.Initialize(game)

	system.applyBurnTargets(game, []rl.Vector3{rl.NewVector3(0, 0, 0)})

	if got := chunk.BurnOverlayImage.RGBAAt(0, 0); got.A != 255 {
		t.Fatalf("burn overlay at (0,0) = %#v, want alpha 255", got)
	}
	if !system.damageMap.Has(chunk.Entity) {
		t.Fatal("chunk entity is missing TerrainChunkDamaged")
	}
}

func TestLaserSystemApplyBurnTargetsIgnoresRepeatedDamageTags(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	manager := &ChunkManager{
		world:           world,
		terrainChunkMap: ecs.NewMap1[TerrainChunkComponent](world),
		chunks:          make(map[ChunkCoords]*TerrainChunk),
	}
	chunk := &TerrainChunk{
		Coords:  ChunkCoords{X: 0, Z: 0},
		OriginX: 0,
		OriginZ: 0,
		Data: terrain.ChunkData{
			Width:  4,
			Height: 4,
		},
		BurnOverlayImage: image.NewRGBA(image.Rect(0, 0, 4, 4)),
	}

	manager.registerTerrainChunkEntity(chunk)
	manager.chunks[chunk.Coords] = chunk

	system := &LaserSystem{}
	game := &Game{world: world, chunkManager: manager}
	system.Initialize(game)

	target := rl.NewVector3(0, 0, 0)
	system.applyBurnTargets(game, []rl.Vector3{target, target})

	if got := chunk.BurnOverlayImage.RGBAAt(0, 0); got.A != 255 {
		t.Fatalf("burn overlay at (0,0) = %#v, want alpha 255", got)
	}
	if !system.damageMap.Has(chunk.Entity) {
		t.Fatal("chunk entity is missing TerrainChunkDamaged")
	}
}

func TestLaserSystemApplyBurnTargetsSkipsEmptyTargets(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	manager := &ChunkManager{
		world:           world,
		terrainChunkMap: ecs.NewMap1[TerrainChunkComponent](world),
		chunks:          make(map[ChunkCoords]*TerrainChunk),
	}
	chunk := &TerrainChunk{
		Coords:  ChunkCoords{X: 0, Z: 0},
		OriginX: 0,
		OriginZ: 0,
		Data: terrain.ChunkData{
			Width:  4,
			Height: 4,
		},
		BurnOverlayImage: image.NewRGBA(image.Rect(0, 0, 4, 4)),
	}

	manager.registerTerrainChunkEntity(chunk)
	manager.chunks[chunk.Coords] = chunk

	system := &LaserSystem{}
	game := &Game{world: world, chunkManager: manager}
	system.Initialize(game)
	system.applyBurnTargets(game, nil)

	if got := chunk.BurnOverlayImage.RGBAAt(0, 0); got.A != 0 {
		t.Fatalf("burn overlay at (0,0) alpha = %d, want 0", got.A)
	}
	if system.damageMap.Has(chunk.Entity) {
		t.Fatal("chunk entity should not be marked damaged")
	}
}

func TestLaserSystemApplyBurnTargetsSpawnsParticlesAfterSuccessfulBurn(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	manager := &ChunkManager{
		world:           world,
		terrainChunkMap: ecs.NewMap1[TerrainChunkComponent](world),
		chunks:          make(map[ChunkCoords]*TerrainChunk),
	}
	chunk := &TerrainChunk{
		Coords:  ChunkCoords{X: 0, Z: 0},
		OriginX: 0,
		OriginZ: 0,
		Data: terrain.ChunkData{
			Width:  4,
			Height: 4,
		},
		BurnOverlayImage: image.NewRGBA(image.Rect(0, 0, 4, 4)),
	}

	manager.registerTerrainChunkEntity(chunk)
	manager.chunks[chunk.Coords] = chunk

	model := &rl.Model{}
	system := &LaserSystem{}
	game := &Game{world: world, chunkManager: manager}
	system.Initialize(game)
	system.particleSpawner = NewParticleSpawnerFactory(world, model, rand.New(rand.NewSource(1)))

	system.applyBurnTargets(game, []rl.Vector3{rl.NewVector3(0, 0, 0)})

	particleFilter := ecs.NewFilter1[Particle](world)
	query := particleFilter.Query()
	defer query.Close()

	count := 0
	for query.Next() {
		count++
	}

	if count != laserStrikeParticleCount {
		t.Fatalf("particle count = %d, want %d", count, laserStrikeParticleCount)
	}
}

func TestDrainLaserBatteryUsesConfiguredAmount(t *testing.T) {
	t.Parallel()

	battery := &Battery{charge: droneBatteryCharge}

	drainLaserBattery(battery)

	if battery.charge != droneBatteryCharge-laserBatteryDrainPerBurn {
		t.Fatalf("battery charge = %.2f, want %.2f", battery.charge, droneBatteryCharge-laserBatteryDrainPerBurn)
	}
}

func TestDrainLaserBatteryClampsAtZero(t *testing.T) {
	t.Parallel()

	battery := &Battery{charge: laserBatteryDrainPerBurn / 2}

	drainLaserBattery(battery)

	if battery.charge != 0 {
		t.Fatalf("battery charge = %.2f, want 0", battery.charge)
	}
}
