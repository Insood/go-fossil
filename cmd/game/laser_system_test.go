package main

import (
	"image"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"

	"go-fossil/internal/terrain"
)

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
