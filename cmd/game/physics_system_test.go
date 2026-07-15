package main

import (
	"testing"

	"go-fossil/internal/terrain"
)

func TestClampDroneVelocityToTerrainBoundsStopsAtMissingChunkEdge(t *testing.T) {
	t.Parallel()

	manager := &ChunkManager{
		chunks: map[ChunkCoords]*TerrainChunk{
			{X: 0, Z: 0}: newTestTerrainChunk(ChunkCoords{X: 0, Z: 0}, 0, 0, 1),
		},
	}

	position := Position3{X: 4, Y: 2, Z: 4}
	velocity := Velocity3{X: droneTopSpeed, Y: 0, Z: 0}

	clamped := clampDroneVelocityToTerrainBounds(position, velocity, 1, manager)

	if got, want := clamped.X, float32(0); got != want {
		t.Fatalf("velocity.X = %v, want %v", got, want)
	}
	if got, want := clamped.Z, float32(0); got != want {
		t.Fatalf("velocity.Z = %v, want %v", got, want)
	}
}

func TestClampDroneVelocityToTerrainBoundsAllowsLoadedNeighborChunk(t *testing.T) {
	t.Parallel()

	manager := &ChunkManager{
		chunks: map[ChunkCoords]*TerrainChunk{
			{X: 0, Z: 0}: newTestTerrainChunk(ChunkCoords{X: 0, Z: 0}, 0, 0, 1),
			{X: 1, Z: 0}: newTestTerrainChunk(ChunkCoords{X: 1, Z: 0}, terrain.ChunkWidthTiles, 0, 2),
		},
	}

	position := Position3{X: 4, Y: 2, Z: 4}
	velocity := Velocity3{X: droneTopSpeed, Y: 0, Z: 0}

	clamped := clampDroneVelocityToTerrainBounds(position, velocity, 1, manager)

	if got, want := clamped.X, float32(4); got != want {
		t.Fatalf("velocity.X = %v, want %v", got, want)
	}
	if got, want := clamped.Z, float32(0); got != want {
		t.Fatalf("velocity.Z = %v, want %v", got, want)
	}
}
