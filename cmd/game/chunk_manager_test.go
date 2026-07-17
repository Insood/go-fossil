package main

import (
	"image"
	"math/rand"
	"testing"

	ecs "github.com/mlange-42/ark/ecs"

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
		wantOK bool
	}{
		{name: "default chunk", worldX: 4, worldZ: 4, want: ChunkCoords{X: 0, Z: 0}, wantY: 1, wantOK: true},
		{name: "north chunk", worldX: 4, worldZ: -4, want: ChunkCoords{X: 0, Z: -1}, wantY: 2, wantOK: true},
		{name: "seam on north edge", worldX: 4, worldZ: 0, want: ChunkCoords{X: 0, Z: 0}, wantY: 1, wantOK: true},
		{name: "default east edge", worldX: 8, worldZ: 0, want: ChunkCoords{X: 0, Z: 0}, wantY: 1, wantOK: true},
		{name: "missing chunk falls back to zero", worldX: 100, worldZ: 100, wantY: 0, wantOK: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			chunk, ok := manager.ChunkForWorldPosition(test.worldX, test.worldZ)
			if ok != test.wantOK {
				t.Fatalf("chunk lookup ok = %v, want %v", ok, test.wantOK)
			}
			if ok && chunk.Coords != test.want {
				t.Fatalf("chunk coords = %#v, want %#v", chunk.Coords, test.want)
			}

			if got := manager.SampleHeight(test.worldX, test.worldZ); got != test.wantY {
				t.Fatalf("sample height = %v, want %v", got, test.wantY)
			}
		})
	}
}

func TestChunkManagerRegistersTerrainChunkEntity(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	manager := &ChunkManager{
		world:           world,
		terrainChunkMap: ecs.NewMap1[TerrainChunkComponent](world),
		chunks:          make(map[ChunkCoords]*TerrainChunk),
	}
	chunk := &TerrainChunk{
		Coords: ChunkCoords{X: 0, Z: 0},
		Data: terrain.ChunkData{
			Width:  terrain.ChunkWidthTiles,
			Height: terrain.ChunkHeightTiles,
		},
	}

	manager.registerTerrainChunkEntity(chunk)

	if chunk.Entity.IsZero() {
		t.Fatal("chunk entity is zero")
	}
	if !world.Alive(chunk.Entity) {
		t.Fatal("chunk entity is not alive")
	}
	if !manager.terrainChunkMap.HasAll(chunk.Entity) {
		t.Fatal("chunk entity is missing TerrainChunkComponent")
	}

	component := manager.terrainChunkMap.Get(chunk.Entity)
	if component == nil {
		t.Fatal("chunk component lookup returned nil")
	}
	if got, want := component.Chunk, chunk; got != want {
		t.Fatalf("chunk component points to %#v, want %#v", got, want)
	}
}

func TestChunkManagerBurnAtWorldPositionBurnsOverlay(t *testing.T) {
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

	if !manager.BurnAtWorldPosition(0, 0) {
		t.Fatal("BurnAtWorldPosition returned false")
	}
	if got := chunk.BurnOverlayImage.RGBAAt(0, 0); got.A != 255 {
		t.Fatalf("burn overlay at (0,0) = %#v, want alpha 255", got)
	}
}

func TestRandomGeneratedArtifactPlacementsRespectSpacingRules(t *testing.T) {
	t.Parallel()

	placements := randomGeneratedArtifactPlacements(
		terrain.ChunkWidthTiles,
		terrain.ChunkHeightTiles,
		[]*ArtifactDefinition{
			{Name: "alpha"},
			{Name: "beta"},
			{Name: "gamma"},
			{Name: "delta"},
		},
		rand.New(rand.NewSource(1)),
	)

	if len(placements) == 0 {
		t.Fatal("placement count = 0, want at least 1")
	}
	if len(placements) > generatedArtifactMaxCount {
		t.Fatalf("placement count = %d, want at most %d", len(placements), generatedArtifactMaxCount)
	}

	seenNames := make(map[string]struct{}, len(placements))
	minX := float32(generatedArtifactEdgeMarginPixels)
	minZ := float32(generatedArtifactEdgeMarginPixels)
	maxX := float32(terrain.ChunkWidthTiles*terrainTexturePixelsPerTile - generatedArtifactEdgeMarginPixels)
	maxZ := float32(terrain.ChunkHeightTiles*terrainTexturePixelsPerTile - generatedArtifactEdgeMarginPixels)
	minGapSquared := float32(generatedArtifactMinCenterGapPixels * generatedArtifactMinCenterGapPixels)

	for i, placement := range placements {
		if placement.X < minX || placement.X > maxX {
			t.Fatalf("placement %d x = %v, want within [%v, %v]", i, placement.X, minX, maxX)
		}
		if placement.Z < minZ || placement.Z > maxZ {
			t.Fatalf("placement %d z = %v, want within [%v, %v]", i, placement.Z, minZ, maxZ)
		}
		if _, exists := seenNames[placement.Name]; exists {
			t.Fatalf("placement %d name = %q, want unique names", i, placement.Name)
		}
		seenNames[placement.Name] = struct{}{}

		for j := 0; j < i; j++ {
			dx := placement.X - placements[j].X
			dz := placement.Z - placements[j].Z
			if dx*dx+dz*dz < minGapSquared {
				t.Fatalf("placements %d and %d are too close: %#v %#v", j, i, placements[j], placement)
			}
		}
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
