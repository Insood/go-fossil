package main

import (
	"math/rand"
	"testing"
)

func TestChunkSpawnerSystemSpawnsOnScoreThreshold(t *testing.T) {
	t.Parallel()

	manager := &fakeChunkSpawnerManager{
		chunks: []*TerrainChunk{
			{Coords: ChunkCoords{X: 1, Z: 0}},
			{Coords: ChunkCoords{X: 0, Z: 0}},
		},
	}
	system := &ChunkSpawnerSystem{
		chunkManager: manager,
		generator:    NewChunkGenerator(),
		rng:          rand.New(rand.NewSource(1)),
		lastScore:    0,
	}
	game := &Game{TotalScore: 451}

	system.Update(game)

	if got, want := len(manager.loaded), 1; got != want {
		t.Fatalf("loaded chunk count = %d, want %d", got, want)
	}
	if got := manager.loaded[0]; !isOneOfChunkCoords(got, []ChunkCoords{
		{X: 2, Z: 0},
		{X: -1, Z: 0},
		{X: 1, Z: 1},
		{X: 1, Z: -1},
		{X: 0, Z: 1},
		{X: 0, Z: -1},
	}) {
		t.Fatalf("loaded chunk coords = %#v, want an exposed edge", got)
	}

	game.TotalScore = 451
	system.Update(game)
	if got, want := len(manager.loaded), 1; got != want {
		t.Fatalf("loaded chunk count after stable score = %d, want %d", got, want)
	}
}

func TestChunkSpawnerSystemDoesNotSpawnBelowThreshold(t *testing.T) {
	t.Parallel()

	manager := &fakeChunkSpawnerManager{
		chunks: []*TerrainChunk{
			{Coords: ChunkCoords{X: 0, Z: 0}},
		},
	}
	system := &ChunkSpawnerSystem{
		chunkManager: manager,
		generator:    NewChunkGenerator(),
		rng:          rand.New(rand.NewSource(1)),
		lastScore:    0,
	}
	game := &Game{TotalScore: 450}

	system.Update(game)

	if got, want := len(manager.loaded), 0; got != want {
		t.Fatalf("loaded chunk count = %d, want %d", got, want)
	}
}

type fakeChunkSpawnerManager struct {
	chunks []*TerrainChunk
	loaded []ChunkCoords
}

func (manager *fakeChunkSpawnerManager) Chunks() []*TerrainChunk {
	return manager.chunks
}

func (manager *fakeChunkSpawnerManager) LoadGeneratedChunk(coords ChunkCoords, generator *ChunkGenerator) *TerrainChunk {
	manager.loaded = append(manager.loaded, coords)
	chunk := &TerrainChunk{Coords: coords}
	manager.chunks = append(manager.chunks, chunk)
	return chunk
}

func isOneOfChunkCoords(value ChunkCoords, options []ChunkCoords) bool {
	for _, option := range options {
		if value == option {
			return true
		}
	}

	return false
}
