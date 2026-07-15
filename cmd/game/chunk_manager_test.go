package main

import (
	"testing"

	"go-fossil/internal/terrain"
)

func TestChunkGeneratorGeneratesFlatChunk(t *testing.T) {
	t.Parallel()

	generator := NewChunkGenerator()
	chunk := generator.GenerateFlat(ChunkCoords{X: 0, Z: -1})

	if got, want := chunk.Width, terrain.ChunkWidthTiles; got != want {
		t.Fatalf("chunk width = %d, want %d", got, want)
	}
	if got, want := chunk.Height, terrain.ChunkHeightTiles; got != want {
		t.Fatalf("chunk height = %d, want %d", got, want)
	}
	if got, want := len(chunk.Tiles), terrain.ChunkHeightTiles; got != want {
		t.Fatalf("tile row count = %d, want %d", got, want)
	}
	if got, want := len(chunk.HeightSamples), terrain.ChunkHeightTiles+1; got != want {
		t.Fatalf("height sample row count = %d, want %d", got, want)
	}
	if got, want := chunk.TileDefinitions, []string{"ground_grid.png"}; len(got) != 1 || got[0] != want[0] {
		t.Fatalf("tile definitions = %#v, want %#v", got, want)
	}

	for row := range chunk.Tiles {
		for column := range chunk.Tiles[row] {
			if got := chunk.Tiles[row][column]; got != 0 {
				t.Fatalf("tiles[%d][%d] = %d, want 0", row, column, got)
			}
		}
	}

	for row := range chunk.HeightSamples {
		for column := range chunk.HeightSamples[row] {
			if got := chunk.HeightSamples[row][column]; got != 0 {
				t.Fatalf("height_samples[%d][%d] = %v, want 0", row, column, got)
			}
		}
	}
}

func TestChunkManagerChunkLookupUsesWorldPosition(t *testing.T) {
	t.Parallel()

	manager := &ChunkManager{
		chunks: map[ChunkCoords]*TerrainChunk{
			{X: 0, Z: 0}:  newTestTerrainChunk(ChunkCoords{X: 0, Z: 0}, 0, 0, 1),
			{X: 0, Z: -1}: newTestTerrainChunk(ChunkCoords{X: 0, Z: -1}, 0, -8, 2),
		},
	}

	tests := []struct {
		name   string
		worldX float32
		worldZ float32
		want   ChunkCoords
		wantY  float32
	}{
		{name: "default chunk", worldX: 4, worldZ: 4, want: ChunkCoords{X: 0, Z: 0}, wantY: 1},
		{name: "north chunk", worldX: 4, worldZ: -4, want: ChunkCoords{X: 0, Z: -1}, wantY: 2},
		{name: "seam on north edge", worldX: 4, worldZ: 0, want: ChunkCoords{X: 0, Z: 0}, wantY: 1},
		{name: "default east edge", worldX: 8, worldZ: 0, want: ChunkCoords{X: 0, Z: 0}, wantY: 1},
		{name: "missing chunk falls back", worldX: 100, worldZ: 100, want: ChunkCoords{X: 0, Z: 0}, wantY: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			chunk := manager.ChunkAtWorldPosition(test.worldX, test.worldZ)
			if chunk.Coords != test.want {
				t.Fatalf("chunk coords = %#v, want %#v", chunk.Coords, test.want)
			}

			if got := manager.SampleHeight(test.worldX, test.worldZ); got != test.wantY {
				t.Fatalf("sample height = %v, want %v", got, test.wantY)
			}
		})
	}
}

func newTestTerrainChunk(coords ChunkCoords, originX, originZ, height float32) *TerrainChunk {
	heightSamples := make([][]float32, terrain.ChunkHeightTiles+1)
	for row := range heightSamples {
		heightSamples[row] = make([]float32, terrain.ChunkWidthTiles+1)
		for column := range heightSamples[row] {
			heightSamples[row][column] = height
		}
	}

	return &TerrainChunk{
		Coords:  coords,
		OriginX: originX,
		OriginZ: originZ,
		Data: terrain.ChunkData{
			Width:         terrain.ChunkWidthTiles,
			Height:        terrain.ChunkHeightTiles,
			HeightSamples: heightSamples,
		},
		SurfaceMesh: &terrain.SurfaceMesh{
			Width:         terrain.ChunkWidthTiles,
			Height:        terrain.ChunkHeightTiles,
			HeightSamples: heightSamples,
		},
	}
}
