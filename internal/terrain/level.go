package terrain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
)

const (
	ChunkWidthTiles  = 8
	ChunkHeightTiles = 8
)

type LevelData struct {
	Name            string
	Width           int
	Height          int
	Tiles           [][]int
	TileDefinitions []string
	SpawnX          float32
	SpawnZ          float32
	HeightSamples   [][]float32
}

type levelMetadata struct {
	Name            *string     `json:"name"`
	Tiles           [][]int     `json:"tiles"`
	TileDefinitions []string    `json:"tile_definitions"`
	SpawnX          *float32    `json:"spawn_x"`
	SpawnZ          *float32    `json:"spawn_z"`
	HeightSamples   [][]float32 `json:"height_samples"`
}

func LoadLevel(levelFS fs.FS, levelPath string) (LevelData, error) {
	return loadLevelMetadata(levelFS, levelPath)
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
		Width:           ChunkWidthTiles,
		Height:          ChunkHeightTiles,
		Tiles:           metadata.Tiles,
		TileDefinitions: metadata.TileDefinitions,
		SpawnX:          *metadata.SpawnX,
		SpawnZ:          *metadata.SpawnZ,
		HeightSamples:   metadata.HeightSamples,
	}, nil
}

func (metadata levelMetadata) validate(levelPath string) error {
	if metadata.Name == nil {
		return fmt.Errorf("%s: missing required key %q", levelPath, "name")
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
	if metadata.HeightSamples == nil {
		return fmt.Errorf("%s: missing required key %q", levelPath, "height_samples")
	}

	if *metadata.SpawnX < 0 || *metadata.SpawnX > float32(ChunkWidthTiles) {
		return fmt.Errorf("%s: spawn_x %.3f must be within 0..%d", levelPath, *metadata.SpawnX, ChunkWidthTiles)
	}
	if *metadata.SpawnZ < 0 || *metadata.SpawnZ > float32(ChunkHeightTiles) {
		return fmt.Errorf("%s: spawn_z %.3f must be within 0..%d", levelPath, *metadata.SpawnZ, ChunkHeightTiles)
	}
	if len(metadata.Tiles) != ChunkHeightTiles {
		return fmt.Errorf("%s: tiles has %d rows, want %d", levelPath, len(metadata.Tiles), ChunkHeightTiles)
	}
	if len(metadata.TileDefinitions) == 0 {
		return fmt.Errorf("%s: tile_definitions must not be empty", levelPath)
	}
	if len(metadata.HeightSamples) != ChunkHeightTiles+1 {
		return fmt.Errorf("%s: height_samples has %d rows, want %d", levelPath, len(metadata.HeightSamples), ChunkHeightTiles+1)
	}

	for rowIndex, row := range metadata.Tiles {
		if len(row) != ChunkWidthTiles {
			return fmt.Errorf("%s: tiles row %d has %d columns, want %d", levelPath, rowIndex, len(row), ChunkWidthTiles)
		}

		for colIndex, tileIndex := range row {
			if tileIndex < 0 || tileIndex >= len(metadata.TileDefinitions) {
				return fmt.Errorf("%s: tiles[%d][%d] references tile definition %d, but valid indices are 0..%d", levelPath, rowIndex, colIndex, tileIndex, len(metadata.TileDefinitions)-1)
			}
		}
	}

	for rowIndex, row := range metadata.HeightSamples {
		if len(row) != ChunkWidthTiles+1 {
			return fmt.Errorf("%s: height_samples row %d has %d columns, want %d", levelPath, rowIndex, len(row), ChunkWidthTiles+1)
		}
	}

	return nil
}
