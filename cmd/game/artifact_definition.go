package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
)

type ArtifactDefinition struct {
	Name      string `json:"name"`
	ImagePath string `json:"image_path"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Value     int    `json:"value"`
}

func loadArtifactDefinitionAsset(assetFS fs.FS, definitionPath string) (ArtifactDefinition, error) {
	definitionBytes, err := fs.ReadFile(assetFS, definitionPath)
	if err != nil {
		return ArtifactDefinition{}, fmt.Errorf("%s: read artifact definition: %w", definitionPath, err)
	}

	var definition ArtifactDefinition
	decoder := json.NewDecoder(bytes.NewReader(definitionBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&definition); err != nil {
		return ArtifactDefinition{}, fmt.Errorf("%s: decode artifact definition: %w", definitionPath, err)
	}

	if definition.Name == "" {
		return ArtifactDefinition{}, fmt.Errorf("%s: artifact definition name must not be empty", definitionPath)
	}
	if definition.ImagePath == "" {
		return ArtifactDefinition{}, fmt.Errorf("%s: artifact image_path must not be empty", definitionPath)
	}
	if definition.Width <= 0 {
		return ArtifactDefinition{}, fmt.Errorf("%s: artifact width must be positive", definitionPath)
	}
	if definition.Height <= 0 {
		return ArtifactDefinition{}, fmt.Errorf("%s: artifact height must be positive", definitionPath)
	}
	if definition.Value <= 0 {
		return ArtifactDefinition{}, fmt.Errorf("%s: artifact value must be positive", definitionPath)
	}

	return definition, nil
}
