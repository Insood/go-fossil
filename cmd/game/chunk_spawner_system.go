package main

import (
	"math/rand"
	"time"
)

type chunkSpawnerManager interface {
	Chunks() []*TerrainChunk
	LoadGeneratedChunk(coords ChunkCoords, generator *ChunkGenerator) *TerrainChunk
}

type ChunkSpawnerSystem struct {
	chunkManager chunkSpawnerManager
	generator    *ChunkGenerator
	rng          *rand.Rand
	lastScore    int
}

func (system *ChunkSpawnerSystem) Initialize(game *Game) {
	system.chunkManager = game.chunkManager
	system.generator = NewChunkGenerator()
	system.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	system.lastScore = game.TotalScore
}

func (system *ChunkSpawnerSystem) Update(game *Game) {
	if system.chunkManager == nil || system.generator == nil {
		return
	}

	if game.TotalScore-system.lastScore <= 450 {
		return
	}

	coords, ok := nextChunkSpawnCoords(system.chunkManager.Chunks(), system.rng)
	if !ok {
		system.lastScore = game.TotalScore
		return
	}

	system.chunkManager.LoadGeneratedChunk(coords, system.generator)
	system.lastScore = game.TotalScore
}

func nextChunkSpawnCoords(chunks []*TerrainChunk, rng *rand.Rand) (ChunkCoords, bool) {
	if len(chunks) == 0 {
		return ChunkCoords{}, false
	}

	occupied := make(map[ChunkCoords]struct{}, len(chunks))
	for _, chunk := range chunks {
		if chunk == nil {
			continue
		}
		occupied[chunk.Coords] = struct{}{}
	}

	directions := []ChunkCoords{
		{X: 1, Z: 0},
		{X: -1, Z: 0},
		{X: 0, Z: 1},
		{X: 0, Z: -1},
	}

	candidates := make([]ChunkCoords, 0)
	for _, chunk := range chunks {
		if chunk == nil {
			continue
		}

		for _, direction := range directions {
			candidate := ChunkCoords{
				X: chunk.Coords.X + direction.X,
				Z: chunk.Coords.Z + direction.Z,
			}
			if _, exists := occupied[candidate]; exists {
				continue
			}

			candidates = append(candidates, candidate)
		}
	}

	if len(candidates) == 0 {
		return ChunkCoords{}, false
	}

	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	return candidates[rng.Intn(len(candidates))], true
}
