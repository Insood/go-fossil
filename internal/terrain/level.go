package terrain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"path"

	_ "image/png"
)

type LevelData struct {
	Name            string
	Width           int
	Height          int
	Tiles           [][]int
	TileDefinitions []string
	HeightmapImage  string
	MinHeight       float32
	MaxHeight       float32
	SpawnX          float32
	SpawnZ          float32
	HeightSamples   [][]float32
}

type levelMetadata struct {
	Name            *string  `json:"name"`
	Width           *int     `json:"width"`
	Height          *int     `json:"height"`
	Tiles           [][]int  `json:"tiles"`
	TileDefinitions []string `json:"tile_definitions"`
	HeightmapImage  *string  `json:"heightmap_image"`
	MinHeight       *float32 `json:"min_height"`
	MaxHeight       *float32 `json:"max_height"`
	SpawnX          *float32 `json:"spawn_x"`
	SpawnZ          *float32 `json:"spawn_z"`
}

func LoadLevel(levelFS fs.FS, levelPath string) (LevelData, error) {
	metadata, err := loadLevelMetadata(levelFS, levelPath)
	if err != nil {
		return LevelData{}, err
	}

	heightmapPath := path.Clean(path.Join(path.Dir(levelPath), metadata.HeightmapImage))
	heightSamples, err := loadHeightSamples(levelFS, levelPath, heightmapPath, metadata.Width, metadata.Height, metadata.MinHeight, metadata.MaxHeight)
	if err != nil {
		return LevelData{}, err
	}

	metadata.HeightSamples = heightSamples
	return metadata, nil
}

func loadLevelMetadata(levelFS fs.FS, levelPath string) (LevelData, error) {
	levelBytes, err := fs.ReadFile(levelFS, levelPath)
	if err != nil {
		return LevelData{}, fmt.Errorf("%s: read level metadata: %w", levelPath, err)
	}

	var metadata levelMetadata
	decoder := json.NewDecoder(bytes.NewReader(levelBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&metadata); err != nil {
		return LevelData{}, fmt.Errorf("%s: decode level metadata: %w", levelPath, err)
	}

	if err := metadata.validate(levelPath); err != nil {
		return LevelData{}, err
	}

	return LevelData{
		Name:            *metadata.Name,
		Width:           *metadata.Width,
		Height:          *metadata.Height,
		Tiles:           metadata.Tiles,
		TileDefinitions: metadata.TileDefinitions,
		HeightmapImage:  *metadata.HeightmapImage,
		MinHeight:       *metadata.MinHeight,
		MaxHeight:       *metadata.MaxHeight,
		SpawnX:          *metadata.SpawnX,
		SpawnZ:          *metadata.SpawnZ,
	}, nil
}

func (metadata levelMetadata) validate(levelPath string) error {
	if metadata.Name == nil {
		return fmt.Errorf("%s: missing required key %q", levelPath, "name")
	}
	if metadata.Width == nil {
		return fmt.Errorf("%s: missing required key %q", levelPath, "width")
	}
	if metadata.Height == nil {
		return fmt.Errorf("%s: missing required key %q", levelPath, "height")
	}
	if metadata.HeightmapImage == nil {
		return fmt.Errorf("%s: missing required key %q", levelPath, "heightmap_image")
	}
	if metadata.MinHeight == nil {
		return fmt.Errorf("%s: missing required key %q", levelPath, "min_height")
	}
	if metadata.MaxHeight == nil {
		return fmt.Errorf("%s: missing required key %q", levelPath, "max_height")
	}
	if metadata.Tiles == nil {
		return fmt.Errorf("%s: missing required key %q", levelPath, "tiles")
	}
	if metadata.TileDefinitions == nil {
		return fmt.Errorf("%s: missing required key %q", levelPath, "tile_definitions")
	}
	if metadata.SpawnX == nil {
		return fmt.Errorf("%s: missing required key %q", levelPath, "spawn_x")
	}
	if metadata.SpawnZ == nil {
		return fmt.Errorf("%s: missing required key %q", levelPath, "spawn_z")
	}

	if *metadata.Width <= 0 {
		return fmt.Errorf("%s: width must be positive, got %d", levelPath, *metadata.Width)
	}
	if *metadata.Height <= 0 {
		return fmt.Errorf("%s: height must be positive, got %d", levelPath, *metadata.Height)
	}
	if *metadata.MinHeight > *metadata.MaxHeight {
		return fmt.Errorf("%s: min_height %.3f must be less than or equal to max_height %.3f", levelPath, *metadata.MinHeight, *metadata.MaxHeight)
	}
	if *metadata.SpawnX < 0 || *metadata.SpawnX > float32(*metadata.Width) {
		return fmt.Errorf("%s: spawn_x %.3f must be within 0..%d", levelPath, *metadata.SpawnX, *metadata.Width)
	}
	if *metadata.SpawnZ < 0 || *metadata.SpawnZ > float32(*metadata.Height) {
		return fmt.Errorf("%s: spawn_z %.3f must be within 0..%d", levelPath, *metadata.SpawnZ, *metadata.Height)
	}
	if len(metadata.Tiles) != *metadata.Height {
		return fmt.Errorf("%s: tiles has %d rows, want %d", levelPath, len(metadata.Tiles), *metadata.Height)
	}
	if len(metadata.TileDefinitions) == 0 {
		return fmt.Errorf("%s: tile_definitions must not be empty", levelPath)
	}

	for rowIndex, row := range metadata.Tiles {
		if len(row) != *metadata.Width {
			return fmt.Errorf("%s: tiles row %d has %d columns, want %d", levelPath, rowIndex, len(row), *metadata.Width)
		}

		for colIndex, tileIndex := range row {
			if tileIndex < 0 || tileIndex >= len(metadata.TileDefinitions) {
				return fmt.Errorf("%s: tiles[%d][%d] references tile definition %d, but valid indices are 0..%d", levelPath, rowIndex, colIndex, tileIndex, len(metadata.TileDefinitions)-1)
			}
		}
	}

	return nil
}

func loadHeightSamples(levelFS fs.FS, levelPath, heightmapPath string, width, height int, minHeight, maxHeight float32) ([][]float32, error) {
	heightmapBytes, err := fs.ReadFile(levelFS, heightmapPath)
	if err != nil {
		return nil, fmt.Errorf("%s: read heightmap image %q: %w", levelPath, heightmapPath, err)
	}

	img, _, err := image.Decode(bytes.NewReader(heightmapBytes))
	if err != nil {
		return nil, fmt.Errorf("%s: decode heightmap image %q: %w", levelPath, heightmapPath, err)
	}

	bounds := img.Bounds()
	expectedWidth := width + 1
	expectedHeight := height + 1
	if bounds.Dx() != expectedWidth || bounds.Dy() != expectedHeight {
		return nil, fmt.Errorf("%s: heightmap image %q is %dx%d, want %dx%d", levelPath, heightmapPath, bounds.Dx(), bounds.Dy(), expectedWidth, expectedHeight)
	}

	heightRange := maxHeight - minHeight
	samples := make([][]float32, expectedHeight)
	for z := 0; z < expectedHeight; z++ {
		samples[z] = make([]float32, expectedWidth)
		for x := 0; x < expectedWidth; x++ {
			gray := color.GrayModel.Convert(img.At(bounds.Min.X+x, bounds.Min.Y+z)).(color.Gray)
			normalized := float32(gray.Y) / 255.0
			samples[z][x] = minHeight + normalized*heightRange
		}
	}

	return samples, nil
}
