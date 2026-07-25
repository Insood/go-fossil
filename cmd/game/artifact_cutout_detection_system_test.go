package main

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"strings"
	"testing"

	ecs "github.com/mlange-42/ark/ecs"
	"go-fossil/internal/terrain"
)

func TestArtifactCutoutDetectionSystemScansOnSixtiethTickAndClearsDamage(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	manager := &ChunkManager{
		world:           world,
		terrainChunkMap: ecs.NewMap1[TerrainChunkComponent](world),
		chunks:          make(map[ChunkCoords]*TerrainChunk),
	}
	chunk := &TerrainChunk{
		Coords:  ChunkCoords{X: 1, Z: -1},
		OriginX: float32(terrain.ChunkWidthTiles),
		OriginZ: -float32(terrain.ChunkHeightTiles),
		Data: terrain.ChunkData{
			Width:  4,
			Height: 4,
		},
		SurfaceTexture:   &terrain.SurfaceTexture{BaseImage: image.NewRGBA(image.Rect(0, 0, 4, 4))},
		ArtifactImage:    image.NewRGBA(image.Rect(0, 0, 4, 4)),
		BurnOverlayImage: image.NewRGBA(image.Rect(0, 0, 4, 4)),
		ArtifactData:     NewArtifactData(image.Rect(0, 0, 3, 3)),
	}

	for x := 0; x < 3; x++ {
		chunk.SurfaceTexture.BaseImage.SetRGBA(x, 0, color.RGBA{G: 64, A: 255})
		chunk.SurfaceTexture.BaseImage.SetRGBA(x, 2, color.RGBA{G: 96, A: 255})
		chunk.ArtifactData.SetID(x, 0, 1)
		chunk.ArtifactData.SetID(x, 2, 2)
		chunk.ArtifactData.SetID(x, 1, -1)
		chunk.ArtifactImage.SetRGBA(x, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		chunk.ArtifactImage.SetRGBA(x, 2, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	}
	chunk.ArtifactImage.SetRGBA(1, 0, color.RGBA{})

	manager.registerTerrainChunkEntity(chunk)
	manager.chunks[chunk.Coords] = chunk

	system := &ArtifactCutoutDetectionSystem{}
	artifactManager := NewArtifactManager()
	artifactManager.RegisterChunkArtifact(chunk, "artifact-one", 10, 3, 1, 0, image.Rect(0, 0, 1, 1))
	artifactManager.RegisterChunkArtifact(chunk, "artifact-two", 10, 3, 1, 2, image.Rect(0, 0, 1, 1))
	game := &Game{world: world, artifactManager: artifactManager, Tick: 0}
	system.Initialize(game)
	system.damageMap.Add(chunk.Entity, &TerrainChunkDamaged{})

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = originalStdout
	}()

	system.Update(game)
	if err := writer.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}

	var output bytes.Buffer
	if _, err := output.ReadFrom(reader); err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	if got := output.String(); got != "" {
		t.Fatalf("output at tick 0 = %q, want empty", got)
	}
	if !system.damageMap.Has(chunk.Entity) {
		t.Fatal("damage tag was cleared before the scan tick")
	}

	reader, writer, err = os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = writer
	game.Tick = 59
	system.Update(game)
	if err := writer.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}

	output.Reset()
	if _, err := output.ReadFrom(reader); err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	got := output.String()
	if !strings.Contains(got, "terrain chunk damaged: (1,-1)") {
		t.Fatalf("output = %q, want damaged chunk line", got)
	}
	if !strings.Contains(got, "artifact region -2 size: 3") {
		t.Fatalf("output = %q, want first region line", got)
	}
	if !strings.Contains(got, "artifact region -3 size: 3") {
		t.Fatalf("output = %q, want second region line", got)
	}
	for x := 0; x < 3; x++ {
		if got := chunk.ArtifactData.IDAt(x, 0); got != -1 {
			t.Fatalf("top row artifact data at %d = %d, want -1", x, got)
		}
		if got := chunk.ArtifactData.IDAt(x, 2); got != -1 {
			t.Fatalf("bottom row artifact data at %d = %d, want -1", x, got)
		}
		if got := chunk.ArtifactImage.RGBAAt(x, 0); got.A != 0 {
			t.Fatalf("top row artifact alpha at %d = %d, want 0", x, got.A)
		}
		if got := chunk.ArtifactImage.RGBAAt(x, 2); got.A != 0 {
			t.Fatalf("bottom row artifact alpha at %d = %d, want 0", x, got.A)
		}
		if got := chunk.BurnOverlayImage.RGBAAt(x, 0); got.A != dugOutOverlayAlpha {
			t.Fatalf("top row burn alpha at %d = %d, want %d", x, got.A, dugOutOverlayAlpha)
		}
		if got := chunk.BurnOverlayImage.RGBAAt(x, 2); got.A != dugOutOverlayAlpha {
			t.Fatalf("bottom row burn alpha at %d = %d, want %d", x, got.A, dugOutOverlayAlpha)
		}
	}
	if got := chunk.BurnOverlayImage.RGBAAt(0, 1); got.A != 0 {
		t.Fatalf("middle row burn alpha = %d, want 0", got.A)
	}
	if system.damageMap.Has(chunk.Entity) {
		t.Fatal("damage tag was not cleared after the scan")
	}
	if got, want := game.TotalScore, 0; got != want {
		t.Fatalf("game total score = %d, want %d", got, want)
	}
	if _, ok := artifactManager.LookupFragment(1); ok {
		t.Fatal("undersized fragment 1 was recorded")
	}
	if _, ok := artifactManager.LookupFragment(2); ok {
		t.Fatal("undersized fragment 2 was recorded")
	}
}

func TestDetectArtifactRegionsUsesUniqueNegativeTags(t *testing.T) {
	t.Parallel()

	data := NewArtifactData(image.Rect(0, 0, 3, 3))
	for x := 0; x < 3; x++ {
		data.SetID(x, 0, 10)
		data.SetID(x, 2, 20)
		data.SetID(x, 1, -1)
	}

	regions := detectArtifactRegions(data)
	if len(regions) != 2 {
		t.Fatalf("region count = %d, want 2", len(regions))
	}
	if regions[0].tag != -2 || regions[0].size != 3 {
		t.Fatalf("first region = %#v, want tag -2 size 3", regions[0])
	}
	if regions[1].tag != -3 || regions[1].size != 3 {
		t.Fatalf("second region = %#v, want tag -3 size 3", regions[1])
	}
	if got := data.IDAt(0, 0); got != 10 {
		t.Fatalf("original data was mutated, top cell = %d", got)
	}
}

func TestScoreArtifactRegionReturnsScoreAndGrade(t *testing.T) {
	t.Parallel()

	manager := NewArtifactManager()
	chunk := &TerrainChunk{
		ArtifactData:  NewArtifactData(image.Rect(0, 0, 4, 1)),
		ArtifactImage: image.NewRGBA(image.Rect(0, 0, 4, 1)),
	}
	artifact := manager.RegisterChunkArtifact(chunk, "test artifact", 100, 4, 0, 0, image.Rect(0, 0, 4, 1))
	for x := 0; x < 3; x++ {
		chunk.ArtifactData.SetID(x, 0, artifact.ID)
		chunk.ArtifactImage.SetRGBA(x, 0, color.RGBA{A: 255})
	}

	score, grade := scoreArtifactRegion(manager, chunk, artifactRegion{
		points: []image.Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}},
	})

	if got, want := score, 75.0; got != want {
		t.Fatalf("score = %f, want %f", got, want)
	}
	if got, want := grade, 0.75; got != want {
		t.Fatalf("grade = %f, want %f", got, want)
	}
}

func TestScoreArtifactRegionCapsGradeAtOne(t *testing.T) {
	t.Parallel()

	manager := NewArtifactManager()
	chunk := &TerrainChunk{
		ArtifactData:  NewArtifactData(image.Rect(0, 0, 3, 1)),
		ArtifactImage: image.NewRGBA(image.Rect(0, 0, 3, 1)),
	}
	artifact := manager.RegisterChunkArtifact(chunk, "test artifact", 10, 2, 0, 0, image.Rect(0, 0, 3, 1))
	points := []image.Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}}
	for _, point := range points {
		chunk.ArtifactData.SetID(point.X, point.Y, artifact.ID)
		chunk.ArtifactImage.SetRGBA(point.X, point.Y, color.RGBA{A: 255})
	}

	score, grade := scoreArtifactRegion(manager, chunk, artifactRegion{points: points})

	if got, want := score, 15.0; got != want {
		t.Fatalf("score = %f, want %f", got, want)
	}
	if got, want := grade, 1.0; got != want {
		t.Fatalf("grade = %f, want %f", got, want)
	}
}
