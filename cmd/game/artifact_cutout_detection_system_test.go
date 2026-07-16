package main

import (
	"bytes"
	"image"
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
		BurnOverlayImage: image.NewRGBA(image.Rect(0, 0, 4, 4)),
	}

	manager.registerTerrainChunkEntity(chunk)
	manager.chunks[chunk.Coords] = chunk

	system := &ArtifactCutoutDetectionSystem{}
	game := &Game{world: world, Tick: 0}
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

	if got := output.String(); !strings.Contains(got, "terrain chunk damaged: (1,-1)") {
		t.Fatalf("output = %q, want damaged chunk line", got)
	}
	if system.damageMap.Has(chunk.Entity) {
		t.Fatal("damage tag was not cleared after the scan")
	}
}
