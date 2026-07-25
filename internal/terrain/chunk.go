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

type ChunkData struct {
	Name            string
	Width           int
	Height          int
	Tiles           [][]int
	TileDefinitions []string
	HeightSamples   [][]float32
	Artifacts       []ArtifactPlacement
	Entities        []EntityPlacement
}

type ArtifactPlacement struct {
	Name        string  `json:"name"`
	X           float32 `json:"x"`
	Z           float32 `json:"z"`
	Orientation float32 `json:"orientation"`
}

type EntityPlacement struct {
	Type string  `json:"type"`
	X    float32 `json:"x"`
	Y    float32 `json:"y"`
	Z    float32 `json:"z"`
}

type chunkMetadata struct {
	Name            *string             `json:"name"`
	Tiles           [][]int             `json:"tiles"`
	TileDefinitions []string            `json:"tile_definitions"`
	HeightSamples   [][]float32         `json:"height_samples"`
	Artifacts       []ArtifactPlacement `json:"artifacts"`
	Entities        []EntityPlacement   `json:"entities"`
}

func LoadChunkData(chunkFS fs.FS, chunkPath string) (ChunkData, error) {
	return loadChunkMetadata(chunkFS, chunkPath)
}

func loadChunkMetadata(chunkFS fs.FS, chunkPath string) (ChunkData, error) {
	chunkBytes, err := fs.ReadFile(chunkFS, chunkPath)
	if err != nil {
		return ChunkData{}, fmt.Errorf("%s: read chunk metadata: %w", chunkPath, err)
	}

	var metadata chunkMetadata
	decoder := json.NewDecoder(bytes.NewReader(chunkBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&metadata); err != nil {
		return ChunkData{}, fmt.Errorf("%s: decode chunk metadata: %w", chunkPath, err)
	}

	if err := metadata.validate(chunkPath); err != nil {
		return ChunkData{}, err
	}

	return ChunkData{
		Name:            *metadata.Name,
		Width:           ChunkWidthTiles,
		Height:          ChunkHeightTiles,
		Tiles:           metadata.Tiles,
		TileDefinitions: metadata.TileDefinitions,
		HeightSamples:   metadata.HeightSamples,
		Artifacts:       metadata.Artifacts,
		Entities:        metadata.Entities,
	}, nil
}

func (metadata chunkMetadata) validate(chunkPath string) error {
	if metadata.Name == nil {
		return fmt.Errorf("%s: missing required key %q", chunkPath, "name")
	}
	if metadata.Tiles == nil {
		return fmt.Errorf("%s: missing required key %q", chunkPath, "tiles")
	}
	if metadata.TileDefinitions == nil {
		return fmt.Errorf("%s: missing required key %q", chunkPath, "tile_definitions")
	}
	if metadata.HeightSamples == nil {
		return fmt.Errorf("%s: missing required key %q", chunkPath, "height_samples")
	}

	if len(metadata.Tiles) != ChunkHeightTiles {
		return fmt.Errorf("%s: tiles has %d rows, want %d", chunkPath, len(metadata.Tiles), ChunkHeightTiles)
	}
	if len(metadata.TileDefinitions) == 0 {
		return fmt.Errorf("%s: tile_definitions must not be empty", chunkPath)
	}
	if len(metadata.HeightSamples) != ChunkHeightTiles+1 {
		return fmt.Errorf("%s: height_samples has %d rows, want %d", chunkPath, len(metadata.HeightSamples), ChunkHeightTiles+1)
	}

	for rowIndex, row := range metadata.Tiles {
		if len(row) != ChunkWidthTiles {
			return fmt.Errorf("%s: tiles row %d has %d columns, want %d", chunkPath, rowIndex, len(row), ChunkWidthTiles)
		}

		for colIndex, tileIndex := range row {
			if tileIndex < 0 || tileIndex >= len(metadata.TileDefinitions) {
				return fmt.Errorf("%s: tiles[%d][%d] references tile definition %d, but valid indices are 0..%d", chunkPath, rowIndex, colIndex, tileIndex, len(metadata.TileDefinitions)-1)
			}
		}
	}

	for rowIndex, row := range metadata.HeightSamples {
		if len(row) != ChunkWidthTiles+1 {
			return fmt.Errorf("%s: height_samples row %d has %d columns, want %d", chunkPath, rowIndex, len(row), ChunkWidthTiles+1)
		}
	}

	for artifactIndex, artifact := range metadata.Artifacts {
		if artifact.Name == "" {
			return fmt.Errorf("%s: artifacts[%d] name must not be empty", chunkPath, artifactIndex)
		}
	}
	for entityIndex, entity := range metadata.Entities {
		if entity.Type == "" {
			return fmt.Errorf("%s: entities[%d] type must not be empty", chunkPath, entityIndex)
		}
	}

	return nil
}
