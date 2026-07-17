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
	artifactManager.fragmentOutputDir = t.TempDir()
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
	if _, ok := artifactManager.LookupFragment(1); !ok {
		t.Fatal("fragment 1 was not recorded")
	}
	if _, ok := artifactManager.LookupFragment(2); !ok {
		t.Fatal("fragment 2 was not recorded")
	}
	fragment, ok := artifactManager.LookupFragment(1)
	if !ok {
		t.Fatal("fragment 1 missing for pixel inspection")
	}
	if got, want := fragment.Weight, 3; got != want {
		t.Fatalf("fragment 1 weight = %d, want %d", got, want)
	}
	if got, want := fragment.Score, 0; got != want {
		t.Fatalf("fragment 1 score = %d, want %d", got, want)
	}
	if got := color.NRGBAModel.Convert(fragment.Image.At(1, 0)).(color.NRGBA); got.G < 60 || got.A != 255 {
		t.Fatalf("fragment pixel (1,0) = %#v, want ground preserved", got)
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
