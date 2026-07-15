package main

import (
	"fmt"
	"math"

	"go-fossil/internal/terrain"
)

type ChunkCoords struct {
	X int
	Z int
}

func (coords ChunkCoords) String() string {
	return fmt.Sprintf("(%d,%d)", coords.X, coords.Z)
}

func chunkCoordsForWorldPosition(worldX, worldZ float32) ChunkCoords {
	return ChunkCoords{
		X: int(math.Floor(float64(worldX) / float64(terrain.ChunkWidthTiles))),
		Z: int(math.Floor(float64(worldZ) / float64(terrain.ChunkHeightTiles))),
	}
}

func chunkOriginForCoords(coords ChunkCoords) (float32, float32) {
	return float32(coords.X * terrain.ChunkWidthTiles), float32(coords.Z * terrain.ChunkHeightTiles)
}
