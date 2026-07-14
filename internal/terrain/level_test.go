package terrain

import (
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadLevelSuccess(t *testing.T) {
	t.Parallel()

	levelFS := fstest.MapFS{
		"levels/level01.json": &fstest.MapFile{
			Data: []byte(`{
  "name": "Level 1",
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
  "spawn_x": 1,
  "spawn_z": 2
}`),
		},
	}

	level, err := LoadLevel(levelFS, "levels/level01.json")
	if err != nil {
		t.Fatalf("LoadLevel() error = %v", err)
	}

	if level.Width != ChunkWidthTiles || level.Height != ChunkHeightTiles {
		t.Fatalf("unexpected level size: got %dx%d", level.Width, level.Height)
	}

	if got := level.HeightSamples[0][0]; got != 1.5 {
		t.Fatalf("height at [0][0] = %v, want 1.5", got)
	}

	if got := level.HeightSamples[0][8]; got != 3.5 {
		t.Fatalf("height at [0][8] = %v, want 3.5", got)
	}

	if got := level.HeightSamples[0][4]; got != 2.5 {
		t.Fatalf("height at [0][4] = %v, want 2.5", got)
	}
	if level.SpawnX != 1 || level.SpawnZ != 2 {
		t.Fatalf("unexpected spawn position: got (%v, %v)", level.SpawnX, level.SpawnZ)
	}
}

func TestLoadLevelValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		levelFS fstest.MapFS
		wantErr string
	}{
		{
			name: "missing height samples key",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(`{
  "name": "Level",
  "tiles": [[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0]],
  "tile_definitions": ["ground_grid.png"],
  "spawn_x": 0,
  "spawn_z": 0
}`),
				},
			},
			wantErr: `missing required key "height_samples"`,
		},
		{
			name: "tiles wrong row count",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(tiledRowsJSON(7, 8), heightRowsJSON(9, 9), 0, 0)),
				},
			},
			wantErr: "tiles has 7 rows, want 8",
		},
		{
			name: "height samples wrong row count",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(tiledRowsJSON(8, 8), heightRowsJSON(8, 9), 0, 0)),
				},
			},
			wantErr: "height_samples has 8 rows, want 9",
		},
		{
			name: "missing spawn x key",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(`{
  "name": "Level",
  "tiles": [[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0]],
  "tile_definitions": ["ground_grid.png"],
  "height_samples": [[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0]],
  "spawn_z": 0
}`),
				},
			},
			wantErr: `missing required key "spawn_x"`,
		},
		{
			name: "spawn z out of bounds",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(tiledRowsJSON(ChunkHeightTiles, ChunkWidthTiles), heightRowsJSON(ChunkHeightTiles+1, ChunkWidthTiles+1), 0, ChunkHeightTiles+1)),
				},
			},
			wantErr: `spawn_z 9.000 must be within 0..8`,
		},
		{
			name: "tiles wrong column count",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(`[[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0]]`, heightRowsJSON(ChunkHeightTiles+1, ChunkWidthTiles+1), 0, 0)),
				},
			},
			wantErr: "tiles row 1 has 7 columns, want 8",
		},
		{
			name: "tile index out of range",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(`[[2,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0]]`, heightRowsJSON(ChunkHeightTiles+1, ChunkWidthTiles+1), 0, 0)),
				},
			},
			wantErr: "tiles[0][0] references tile definition 2",
		},
		{
			name: "height samples wrong column count",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(tiledRowsJSON(ChunkHeightTiles, ChunkWidthTiles), `[[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0,0]]`, 0, 0)),
				},
			},
			wantErr: "height_samples row 1 has 8 columns, want 9",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := LoadLevel(test.levelFS, "levels/level.json")
			if err == nil {
				t.Fatal("LoadLevel() error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("LoadLevel() error = %q, want substring %q", err.Error(), test.wantErr)
			}
		})
	}
}

func validMetadataJSON(tilesJSON, heightSamplesJSON string, spawnX, spawnZ float32) string {
	return fmt.Sprintf(`{
  "name": "Level",
  "tiles": %s,
  "tile_definitions": ["ground_grid.png", "cellphone.png"],
  "height_samples": %s,
  "spawn_x": %g,
  "spawn_z": %g
}`, tilesJSON, heightSamplesJSON, spawnX, spawnZ)
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
