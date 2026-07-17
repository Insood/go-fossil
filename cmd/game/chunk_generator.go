package main

import (
	"fmt"
	"math"

	"go-fossil/internal/terrain"
)

type ChunkGenerator struct{}

func NewChunkGenerator() *ChunkGenerator {
	return &ChunkGenerator{}
}

func (generator *ChunkGenerator) GenerateFlat(coords ChunkCoords) terrain.ChunkData {
	tiles := make([][]int, terrain.ChunkHeightTiles)
	for row := range tiles {
		tiles[row] = make([]int, terrain.ChunkWidthTiles)
	}

	heightSamples := make([][]float32, terrain.ChunkHeightTiles+1)
	for row := range heightSamples {
		heightSamples[row] = make([]float32, terrain.ChunkWidthTiles+1)
	}

	return terrain.ChunkData{
		Name:            fmt.Sprintf("generated_%d_%d", coords.X, coords.Z),
		Width:           terrain.ChunkWidthTiles,
		Height:          terrain.ChunkHeightTiles,
		Tiles:           tiles,
		TileDefinitions: []string{"ground_grid.png"},
		HeightSamples:   heightSamples,
	}
}

func (generator *ChunkGenerator) GenerateRandom(coords ChunkCoords) terrain.ChunkData {
	tiles := make([][]int, terrain.ChunkHeightTiles)
	for row := range tiles {
		tiles[row] = make([]int, terrain.ChunkWidthTiles)
	}

	heightSamples := make([][]float32, terrain.ChunkHeightTiles+1)
	for row := range heightSamples {
		heightSamples[row] = make([]float32, terrain.ChunkWidthTiles+1)
		for column := range heightSamples[row] {
			heightSamples[row][column] = float32(math.NaN())
		}
	}

	return terrain.ChunkData{
		Name:            fmt.Sprintf("generated_%d_%d", coords.X, coords.Z),
		Width:           terrain.ChunkWidthTiles,
		Height:          terrain.ChunkHeightTiles,
		Tiles:           tiles,
		TileDefinitions: []string{"ground_grid.png"},
		HeightSamples:   heightSamples,
	}
}
