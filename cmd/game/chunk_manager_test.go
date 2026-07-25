package main

import (
	"fmt"
	"image"
	"math"
	"math/rand"
	"strings"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
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

func TestChunkManagerRegistersTypedChunkEntities(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	model := &rl.Model{}
	manager := &ChunkManager{
		world:  world,
		assets: &AssetManager{models: map[string]*rl.Model{chargingPadModelName: model}},
	}
	chunk := &TerrainChunk{
		Coords:  ChunkCoords{X: 1, Z: -1},
		OriginX: 8,
		OriginZ: -8,
		Data: terrain.ChunkData{
			Entities: []terrain.EntityPlacement{
				{Type: chargingPadEntityType, X: 96, Y: 64, Z: 416},
			},
		},
	}

	manager.registerChunkEntities(chunk)

	if got, want := len(chunk.ChunkEntities), 1; got != want {
		t.Fatalf("chunk entity count = %d, want %d", got, want)
	}

	entity := chunk.ChunkEntities[0]
	if !world.Alive(entity) {
		t.Fatal("chunk entity is not alive")
	}

	position, renderable, chargingPad := manager.chargingPadMap.Get(entity)
	assertVector3(t, rl.Vector3(*position), 9.5, 1, -1.5)
	if got, want := renderable.model, model; got != want {
		t.Fatalf("renderable model = %p, want %p", got, want)
	}
	if chargingPad == nil {
		t.Fatal("charging pad component is nil")
	}
	if !renderable.castsShadow {
		t.Fatal("renderable castsShadow = false, want true")
	}
	if !renderable.receivesShadow {
		t.Fatal("renderable receivesShadow = false, want true")
	}

	manager.removeChunkEntities(chunk)
	if world.Alive(entity) {
		t.Fatal("chunk entity is still alive after removal")
	}
	if len(chunk.ChunkEntities) != 0 {
		t.Fatalf("chunk entity count after removal = %d, want 0", len(chunk.ChunkEntities))
	}
}

func TestChunkManagerRejectsUnsupportedChunkEntityType(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	manager := &ChunkManager{
		world:  world,
		assets: &AssetManager{},
	}
	chunk := &TerrainChunk{
		Coords: ChunkCoords{X: 2, Z: -3},
		Data: terrain.ChunkData{
			Entities: []terrain.EntityPlacement{
				{Type: "unknown"},
			},
		},
	}

	assertPanicsContaining(t, `chunk (2,-3) entity placement 0 has unsupported type "unknown"`, func() {
		manager.registerChunkEntities(chunk)
	})
}

func TestChunkManagerReportsMissingChunkEntityModel(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	manager := &ChunkManager{
		world:  world,
		assets: &AssetManager{models: map[string]*rl.Model{}},
	}
	chunk := &TerrainChunk{
		Coords: ChunkCoords{X: -1, Z: 4},
		Data: terrain.ChunkData{
			Entities: []terrain.EntityPlacement{
				{Type: chargingPadEntityType},
			},
		},
	}

	assertPanicsContaining(t, `chunk (-1,4) entity placement 0 type "charging_pad": missing model "charging_pad"`, func() {
		manager.registerChunkEntities(chunk)
	})
}

func assertPanicsContaining(t *testing.T, want string, action func()) {
	t.Helper()

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("action did not panic, want message containing %q", want)
		}
		if got := fmt.Sprint(recovered); !strings.Contains(got, want) {
			t.Fatalf("panic = %q, want substring %q", got, want)
		}
	}()

	action()
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
			{Name: "alpha", RelativeScarcity: 1},
			{Name: "beta", RelativeScarcity: 1},
			{Name: "gamma", RelativeScarcity: 1},
			{Name: "delta", RelativeScarcity: 1},
		},
		4,
		rand.New(rand.NewSource(1)),
	)

	if len(placements) == 0 {
		t.Fatal("placement count = 0, want at least 1")
	}
	if len(placements) > generatedArtifactMaxCount {
		t.Fatalf("placement count = %d, want at most %d", len(placements), generatedArtifactMaxCount)
	}

	validNames := map[string]bool{"alpha": true, "beta": true, "gamma": true, "delta": true}
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
		if !validNames[placement.Name] {
			t.Fatalf("placement %d name = %q, want a loaded artifact name", i, placement.Name)
		}

		for j := 0; j < i; j++ {
			dx := placement.X - placements[j].X
			dz := placement.Z - placements[j].Z
			if dx*dx+dz*dz < minGapSquared {
				t.Fatalf("placements %d and %d are too close: %#v %#v", j, i, placements[j], placement)
			}
		}
	}
}

func TestRandomArtifactDefinitionUsesRelativeScarcity(t *testing.T) {
	t.Parallel()

	definitions := []*ArtifactDefinition{
		{Name: "a", RelativeScarcity: 1},
		{Name: "b", RelativeScarcity: 5},
	}
	rng := rand.New(rand.NewSource(1))

	const selections = 6000
	counts := make(map[string]int)
	for range selections {
		definition := randomArtifactDefinition(definitions, 6, rng)
		if definition == nil {
			t.Fatal("randomArtifactDefinition() = nil, want a definition")
		}
		counts[definition.Name]++
	}

	if got := counts["a"]; got < 850 || got > 1150 {
		t.Fatalf("artifact a selected %d/%d times, want approximately 1/6", got, selections)
	}
	if got := counts["b"]; got != selections-counts["a"] {
		t.Fatalf("artifact b selected %d/%d times, want remaining selections", got, selections)
	}
}

func TestRandomGeneratedArtifactPlacementsAllowRepeatedDefinitions(t *testing.T) {
	t.Parallel()

	definitions := []*ArtifactDefinition{{Name: "only", RelativeScarcity: 1}}
	foundThreePlacements := false
	for seed := int64(0); seed < 100; seed++ {
		placements := randomGeneratedArtifactPlacements(
			terrain.ChunkWidthTiles,
			terrain.ChunkHeightTiles,
			definitions,
			1,
			rand.New(rand.NewSource(seed)),
		)
		if len(placements) == generatedArtifactMaxCount {
			foundThreePlacements = true
			for i, placement := range placements {
				if placement.Name != "only" {
					t.Fatalf("placement %d name = %q, want only", i, placement.Name)
				}
			}
			break
		}
	}

	if !foundThreePlacements {
		t.Fatalf("never generated %d placements from one definition", generatedArtifactMaxCount)
	}
}

func TestPopulateGeneratedChunkHeightsMatchesNeighborEdges(t *testing.T) {
	t.Parallel()

	manager := &ChunkManager{
		chunks: make(map[ChunkCoords]*TerrainChunk),
		rng:    rand.New(rand.NewSource(1)),
	}

	neighborHeights := map[ChunkCoords]float32{
		{X: -1, Z: -1}: 0.1,
		{X: 0, Z: -1}:  0.2,
		{X: 1, Z: -1}:  0.3,
		{X: -1, Z: 0}:  0.4,
		{X: 1, Z: 0}:   0.5,
		{X: -1, Z: 1}:  0.6,
		{X: 0, Z: 1}:   0.7,
		{X: 1, Z: 1}:   0.8,
	}
	for coords, height := range neighborHeights {
		manager.chunks[coords] = newTestTerrainChunk(coords, float32(coords.X*terrain.ChunkWidthTiles), float32(coords.Z*terrain.ChunkHeightTiles), height)
	}

	chunkData := terrain.ChunkData{
		Width:         terrain.ChunkWidthTiles,
		Height:        terrain.ChunkHeightTiles,
		HeightSamples: make([][]float32, terrain.ChunkHeightTiles+1),
	}
	for row := range chunkData.HeightSamples {
		chunkData.HeightSamples[row] = make([]float32, terrain.ChunkWidthTiles+1)
		for column := range chunkData.HeightSamples[row] {
			chunkData.HeightSamples[row][column] = float32(math.NaN())
		}
	}

	manager.populateGeneratedChunkHeights(ChunkCoords{X: 0, Z: 0}, &chunkData)

	for z := 0; z <= terrain.ChunkHeightTiles; z++ {
		for x := 0; x <= terrain.ChunkWidthTiles; x++ {
			got := chunkData.HeightSamples[z][x]
			if math.IsNaN(float64(got)) {
				t.Fatalf("height sample (%d,%d) is NaN", x, z)
			}
			if got < 0 || got > 2 {
				t.Fatalf("height sample (%d,%d) = %v, want within [0,2]", x, z, got)
			}
		}
	}

	if got, want := chunkData.HeightSamples[0][0], neighborHeights[ChunkCoords{X: -1, Z: -1}]; got != want {
		t.Fatalf("northwest corner = %v, want %v", got, want)
	}
	if got, want := chunkData.HeightSamples[0][terrain.ChunkWidthTiles], neighborHeights[ChunkCoords{X: 1, Z: -1}]; got != want {
		t.Fatalf("northeast corner = %v, want %v", got, want)
	}
	if got, want := chunkData.HeightSamples[terrain.ChunkHeightTiles][0], neighborHeights[ChunkCoords{X: -1, Z: 1}]; got != want {
		t.Fatalf("southwest corner = %v, want %v", got, want)
	}
	if got, want := chunkData.HeightSamples[terrain.ChunkHeightTiles][terrain.ChunkWidthTiles], neighborHeights[ChunkCoords{X: 1, Z: 1}]; got != want {
		t.Fatalf("southeast corner = %v, want %v", got, want)
	}

	for x := 1; x < terrain.ChunkWidthTiles; x++ {
		if got, want := chunkData.HeightSamples[0][x], neighborHeights[ChunkCoords{X: 0, Z: -1}]; got != want {
			t.Fatalf("north edge x=%d = %v, want %v", x, got, want)
		}
		if got, want := chunkData.HeightSamples[terrain.ChunkHeightTiles][x], neighborHeights[ChunkCoords{X: 0, Z: 1}]; got != want {
			t.Fatalf("south edge x=%d = %v, want %v", x, got, want)
		}
	}
	for z := 1; z < terrain.ChunkHeightTiles; z++ {
		if got, want := chunkData.HeightSamples[z][0], neighborHeights[ChunkCoords{X: -1, Z: 0}]; got != want {
			t.Fatalf("west edge z=%d = %v, want %v", z, got, want)
		}
		if got, want := chunkData.HeightSamples[z][terrain.ChunkWidthTiles], neighborHeights[ChunkCoords{X: 1, Z: 0}]; got != want {
			t.Fatalf("east edge z=%d = %v, want %v", z, got, want)
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
