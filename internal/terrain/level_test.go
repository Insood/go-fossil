package terrain

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
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
  "width": 2,
  "height": 2,
  "tiles": [
    [0, 1],
    [1, 0]
  ],
  "tile_definitions": [
    "ground_grid.png",
    "cellphone.png"
  ],
  "heightmap_image": "level01_height.png",
  "min_height": 1.5,
  "max_height": 3.5,
  "spawn_x": 1,
  "spawn_z": 2
}`),
		},
		"levels/level01_height.png": &fstest.MapFile{
			Data: grayscalePNG(t, [][]uint8{
				{0, 128, 255},
				{64, 128, 192},
				{255, 128, 0},
			}),
		},
	}

	level, err := LoadLevel(levelFS, "levels/level01.json")
	if err != nil {
		t.Fatalf("LoadLevel() error = %v", err)
	}

	if level.Width != 2 || level.Height != 2 {
		t.Fatalf("unexpected level size: got %dx%d", level.Width, level.Height)
	}

	if got := level.HeightSamples[0][0]; got != 1.5 {
		t.Fatalf("height at [0][0] = %v, want 1.5", got)
	}

	if got := level.HeightSamples[0][2]; got != 3.5 {
		t.Fatalf("height at [0][2] = %v, want 3.5", got)
	}

	if got := level.HeightSamples[0][1]; !roughlyEqual(got, 1.5+(float32(128)/255.0)*2.0) {
		t.Fatalf("height at [0][1] = %v, want midpoint mapping", got)
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
			name: "missing heightmap image key",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(`{
  "name": "Level",
  "width": 1,
  "height": 1,
  "tiles": [[0]],
  "tile_definitions": ["ground_grid.png"],
  "min_height": 0,
  "max_height": 1,
  "spawn_z": 0
}`),
				},
			},
			wantErr: `missing required key "heightmap_image"`,
		},
		{
			name: "missing spawn x key",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(`{
  "name": "Level",
  "width": 1,
  "height": 1,
  "tiles": [[0]],
  "tile_definitions": ["ground_grid.png"],
  "heightmap_image": "level.png",
  "min_height": 0,
  "max_height": 1,
  "spawn_z": 0
}`),
				},
				"levels/level.png": &fstest.MapFile{Data: grayscalePNG(t, [][]uint8{{0, 0}, {0, 0}})},
			},
			wantErr: `missing required key "spawn_x"`,
		},
		{
			name: "spawn z out of bounds",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(1, 1, `[[0]]`, `"level_height.png"`, 0, 1, 0, 2)),
				},
				"levels/level_height.png": &fstest.MapFile{Data: grayscalePNG(t, [][]uint8{{0, 0}, {0, 0}})},
			},
			wantErr: `spawn_z 2.000 must be within 0..1`,
		},
		{
			name: "tiles wrong row count",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(2, 2, `[[0,0]]`, `"level_height.png"`, 0, 1, 0, 0)),
				},
				"levels/level_height.png": &fstest.MapFile{Data: grayscalePNG(t, [][]uint8{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}})},
			},
			wantErr: "tiles has 1 rows, want 2",
		},
		{
			name: "tiles wrong column count",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(2, 2, `[[0,0],[0]]`, `"level_height.png"`, 0, 1, 0, 0)),
				},
				"levels/level_height.png": &fstest.MapFile{Data: grayscalePNG(t, [][]uint8{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}})},
			},
			wantErr: "tiles row 1 has 1 columns, want 2",
		},
		{
			name: "tile index out of range",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(1, 1, `[[2]]`, `"level_height.png"`, 0, 1, 0, 0)),
				},
				"levels/level_height.png": &fstest.MapFile{Data: grayscalePNG(t, [][]uint8{{0, 0}, {0, 0}})},
			},
			wantErr: "tiles[0][0] references tile definition 2",
		},
		{
			name: "min height exceeds max height",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(1, 1, `[[0]]`, `"level_height.png"`, 2, 1, 0, 0)),
				},
				"levels/level_height.png": &fstest.MapFile{Data: grayscalePNG(t, [][]uint8{{0, 0}, {0, 0}})},
			},
			wantErr: "min_height 2.000 must be less than or equal to max_height 1.000",
		},
		{
			name: "missing heightmap image asset",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(1, 1, `[[0]]`, `"missing.png"`, 0, 1, 0, 0)),
				},
			},
			wantErr: `read heightmap image "levels/missing.png"`,
		},
		{
			name: "malformed heightmap image asset",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(1, 1, `[[0]]`, `"broken.png"`, 0, 1, 0, 0)),
				},
				"levels/broken.png": &fstest.MapFile{Data: []byte("not a png")},
			},
			wantErr: `decode heightmap image "levels/broken.png"`,
		},
		{
			name: "heightmap dimensions mismatch",
			levelFS: fstest.MapFS{
				"levels/level.json": &fstest.MapFile{
					Data: []byte(validMetadataJSON(2, 2, `[[0,0],[0,0]]`, `"level_height.png"`, 0, 1, 0, 0)),
				},
				"levels/level_height.png": &fstest.MapFile{Data: grayscalePNG(t, [][]uint8{{0, 0}, {0, 0}})},
			},
			wantErr: `heightmap image "levels/level_height.png" is 2x2, want 3x3`,
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

func grayscalePNG(t *testing.T, rows [][]uint8) []byte {
	t.Helper()

	height := len(rows)
	width := len(rows[0])
	img := image.NewGray(image.Rect(0, 0, width, height))
	for y, row := range rows {
		for x, value := range row {
			img.SetGray(x, y, color.Gray{Y: value})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}

	return buf.Bytes()
}

func validMetadataJSON(width, height int, tilesJSON, heightmapImage string, minHeight, maxHeight, spawnX, spawnZ float32) string {
	return fmt.Sprintf(`{
  "name": "Level",
  "width": %d,
  "height": %d,
  "tiles": %s,
  "tile_definitions": ["ground_grid.png", "cellphone.png"],
  "heightmap_image": %s,
  "min_height": %g,
  "max_height": %g,
  "spawn_x": %g,
  "spawn_z": %g
}`, width, height, tilesJSON, heightmapImage, minHeight, maxHeight, spawnX, spawnZ)
}

func roughlyEqual(a, b float32) bool {
	const epsilon = 0.0001
	delta := a - b
	if delta < 0 {
		delta = -delta
	}
	return delta < epsilon
}
