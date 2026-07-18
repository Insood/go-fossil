package main

import (
	"image"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"

	"go-fossil/internal/terrain"
)

func TestLaserBurnCursorsUsesCurrentCursorWhenPreviousFrameWasNotFiring(t *testing.T) {
	t.Parallel()

	cursors := laserBurnCursors(DroneFireControl{
		cursor:     rl.NewVector2(20, 30),
		firing:     true,
		lastCursor: rl.NewVector2(10, 30),
		lastFiring: false,
	})

	if len(cursors) != 1 {
		t.Fatalf("len(cursors) = %d, want 1", len(cursors))
	}
	if cursors[0].X != 20 || cursors[0].Y != 30 {
		t.Fatalf("cursor = (%.2f, %.2f), want (20.00, 30.00)", cursors[0].X, cursors[0].Y)
	}
}

func TestLaserBurnCursorsInterpolatesWithConfiguredMaximumStep(t *testing.T) {
	t.Parallel()

	cursors := laserBurnCursors(DroneFireControl{
		cursor:     rl.NewVector2(12, 0),
		firing:     true,
		lastCursor: rl.NewVector2(0, 0),
		lastFiring: true,
	})

	if len(cursors) != 3 {
		t.Fatalf("len(cursors) = %d, want 3", len(cursors))
	}

	wantX := []float32{4, 8, 12}
	for i, cursor := range cursors {
		if cursor.X != wantX[i] || cursor.Y != 0 {
			t.Fatalf("cursor %d = (%.2f, %.2f), want (%.2f, 0.00)", i, cursor.X, cursor.Y, wantX[i])
		}
	}
}

func TestLaserBurnCursorsKeepsSmallMotionToSingleCurrentCursor(t *testing.T) {
	t.Parallel()

	cursors := laserBurnCursors(DroneFireControl{
		cursor:     rl.NewVector2(4, 0),
		firing:     true,
		lastCursor: rl.NewVector2(0, 0),
		lastFiring: true,
	})

	if len(cursors) != 1 {
		t.Fatalf("len(cursors) = %d, want 1", len(cursors))
	}
	if cursors[0].X != 4 || cursors[0].Y != 0 {
		t.Fatalf("cursor = (%.2f, %.2f), want (4.00, 0.00)", cursors[0].X, cursors[0].Y)
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
