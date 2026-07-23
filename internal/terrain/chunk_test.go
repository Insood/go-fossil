package terrain

import (
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadChunkDataSuccess(t *testing.T) {
	t.Parallel()

	chunkFS := fstest.MapFS{
		"terrain_chunks/default.json": &fstest.MapFile{
			Data: []byte(`{
  "name": "Default Chunk",
  "tiles": [
    [0, 1, 0, 1, 0, 1, 0, 1],
    [1, 0, 1, 0, 1, 0, 1, 0],
    [0, 1, 0, 1, 0, 1, 0, 1],
    [1, 0, 1, 0, 1, 0, 1, 0],
    [0, 1, 0, 1, 0, 1, 0, 1],
    [1, 0, 1, 0, 1, 0, 1, 0],
    [0, 1, 0, 1, 0, 1, 0, 1],
    [1, 0, 1, 0, 1, 0, 1, 0]
  ],
  "tile_definitions": [
    "ground_grid.png",
    "cellphone.png"
  ],
  "height_samples": [
    [1.5, 1.75, 2.0, 2.25, 2.5, 2.75, 3.0, 3.25, 3.5],
    [1.6, 1.85, 2.1, 2.35, 2.6, 2.85, 3.1, 3.35, 3.5],
    [1.7, 1.95, 2.2, 2.45, 2.7, 2.95, 3.2, 3.35, 3.5],
    [1.8, 2.05, 2.3, 2.55, 2.8, 3.05, 3.2, 3.35, 3.5],
    [1.9, 2.15, 2.4, 2.65, 2.9, 3.05, 3.2, 3.35, 3.5],
    [2.0, 2.25, 2.5, 2.75, 2.9, 3.05, 3.2, 3.35, 3.5],
    [2.1, 2.35, 2.6, 2.75, 2.9, 3.05, 3.2, 3.35, 3.5],
    [2.2, 2.45, 2.6, 2.75, 2.9, 3.05, 3.2, 3.35, 3.5],
    [2.3, 2.55, 2.7, 2.85, 3.0, 3.15, 3.3, 3.4, 3.5]
  ],
  "artifacts": [
    { "name": "phone", "x": 128, "z": 128, "orientation": 45 }
  ],
  "models": [
    { "name": "charging_pad", "x": 96, "y": 64, "z": 416 }
  ]
}`),
		},
	}

	chunk, err := LoadChunkData(chunkFS, "terrain_chunks/default.json")
	if err != nil {
		t.Fatalf("LoadChunkData() error = %v", err)
	}

	if chunk.Width != ChunkWidthTiles || chunk.Height != ChunkHeightTiles {
		t.Fatalf("unexpected chunk size: got %dx%d", chunk.Width, chunk.Height)
	}

	if got := chunk.HeightSamples[0][0]; got != 1.5 {
		t.Fatalf("height at [0][0] = %v, want 1.5", got)
	}

	if got := chunk.HeightSamples[0][8]; got != 3.5 {
		t.Fatalf("height at [0][8] = %v, want 3.5", got)
	}

	if got := chunk.HeightSamples[0][4]; got != 2.5 {
		t.Fatalf("height at [0][4] = %v, want 2.5", got)
	}
	if chunk.Name != "Default Chunk" {
		t.Fatalf("unexpected chunk name: got %q", chunk.Name)
	}
	if got, want := len(chunk.Artifacts), 1; got != want {
		t.Fatalf("artifact count = %d, want %d", got, want)
	}
	if got := chunk.Artifacts[0].Name; got != "phone" {
		t.Fatalf("artifact name = %q, want phone", got)
	}
	if got := chunk.Artifacts[0].X; got != 128 {
		t.Fatalf("artifact x = %v, want 128", got)
	}
	if got := chunk.Artifacts[0].Z; got != 128 {
		t.Fatalf("artifact z = %v, want 128", got)
	}
	if got := chunk.Artifacts[0].Orientation; got != 45 {
		t.Fatalf("artifact orientation = %v, want 45", got)
	}
	if got, want := len(chunk.Models), 1; got != want {
		t.Fatalf("model count = %d, want %d", got, want)
	}
	if got := chunk.Models[0].Name; got != "charging_pad" {
		t.Fatalf("model name = %q, want charging_pad", got)
	}
	if got := chunk.Models[0].X; got != 96 {
		t.Fatalf("model x = %v, want 96", got)
	}
	if got := chunk.Models[0].Y; got != 64 {
		t.Fatalf("model y = %v, want 64", got)
	}
	if got := chunk.Models[0].Z; got != 416 {
		t.Fatalf("model z = %v, want 416", got)
	}
}

func TestLoadChunkDataValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		chunkFS fstest.MapFS
		wantErr string
	}{
		{
			name: "missing height samples key",
			chunkFS: fstest.MapFS{
				"terrain_chunks/chunk.json": &fstest.MapFile{
					Data: []byte(`{
  "name": "Chunk",
  "tiles": [[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0]],
  "tile_definitions": ["ground_grid.png"]
}`),
				},
			},
			wantErr: `missing required key "height_samples"`,
		},
		{
			name: "artifact missing name",
			chunkFS: fstest.MapFS{
				"terrain_chunks/chunk.json": &fstest.MapFile{
					Data: []byte(`{
  "name": "Chunk",
  "tiles": [[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0]],
  "tile_definitions": ["ground_grid.png"],
  "height_samples": [[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0]],
  "artifacts": [{"x": 1, "z": 2, "orientation": 45}]
}`),
				},
			},
			wantErr: `artifacts[0] name must not be empty`,
		},
		{
			name: "model missing name",
			chunkFS: fstest.MapFS{
				"terrain_chunks/chunk.json": &fstest.MapFile{
					Data: []byte(`{
  "name": "Chunk",
  "tiles": [[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0]],
  "tile_definitions": ["ground_grid.png"],
  "height_samples": [[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0]],
  "models": [{"x": 1, "y": 2, "z": 3}]
}`),
				},
			},
			wantErr: `models[0] name must not be empty`,
		},
		{
			name: "tiles wrong row count",
			chunkFS: fstest.MapFS{
				"terrain_chunks/chunk.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(tiledRowsJSON(7, 8), heightRowsJSON(9, 9))),
				},
			},
			wantErr: "tiles has 7 rows, want 8",
		},
		{
			name: "height samples wrong row count",
			chunkFS: fstest.MapFS{
				"terrain_chunks/chunk.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(tiledRowsJSON(8, 8), heightRowsJSON(8, 9))),
				},
			},
			wantErr: "height_samples has 8 rows, want 9",
		},
		{
			name: "tiles wrong column count",
			chunkFS: fstest.MapFS{
				"terrain_chunks/chunk.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(`[[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0]]`, heightRowsJSON(ChunkHeightTiles+1, ChunkWidthTiles+1))),
				},
			},
			wantErr: "tiles row 1 has 7 columns, want 8",
		},
		{
			name: "tile index out of range",
			chunkFS: fstest.MapFS{
				"terrain_chunks/chunk.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(`[[2,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0]]`, heightRowsJSON(ChunkHeightTiles+1, ChunkWidthTiles+1))),
				},
			},
			wantErr: "tiles[0][0] references tile definition 2",
		},
		{
			name: "height samples wrong column count",
			chunkFS: fstest.MapFS{
				"terrain_chunks/chunk.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(tiledRowsJSON(ChunkHeightTiles, ChunkWidthTiles), `[[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0]]`)),
				},
			},
			wantErr: "height_samples row 1 has 8 columns, want 9",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := LoadChunkData(test.chunkFS, "terrain_chunks/chunk.json")
			if err == nil {
				t.Fatal("LoadChunkData() error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("LoadChunkData() error = %q, want substring %q", err.Error(), test.wantErr)
			}
		})
	}
}

func validMetadataJSON(tilesJSON, heightSamplesJSON string) string {
	return fmt.Sprintf(`{
  "name": "Chunk",
  "tiles": %s,
  "tile_definitions": ["ground_grid.png", "cellphone.png"],
  "height_samples": %s,
  "artifacts": [{"name":"phone","x":128,"z":128,"orientation":45}]
}`, tilesJSON, heightSamplesJSON)
}

func tiledRowsJSON(rowCount, columnCount int) string {
	rows := make([]string, rowCount)
	for row := range rowCount {
		columns := make([]string, columnCount)
		for column := range columnCount {
			columns[column] = "0"
		}
		rows[row] = "[" + strings.Join(columns, ",") + "]"
	}
	return "[" + strings.Join(rows, ",") + "]"
}

func heightRowsJSON(rowCount, columnCount int) string {
	rows := make([]string, rowCount)
	for row := range rowCount {
		columns := make([]string, columnCount)
		for column := range columnCount {
			columns[column] = "0"
		}
		rows[row] = "[" + strings.Join(columns, ",") + "]"
	}
	return "[" + strings.Join(rows, ",") + "]"
}

func roughlyEqual(a, b float32) bool {
	const epsilon = 0.0001
	delta := a - b
	if delta < 0 {
		delta = -delta
	}
	return delta < epsilon
}
