package main

import (
	"image"
)

type ArtifactManager struct {
	nextID    int32
	artifacts map[int32]*Artifact
}

func NewArtifactManager() *ArtifactManager {
	return &ArtifactManager{
		artifacts: make(map[int32]*Artifact),
	}
}

func (manager *ArtifactManager) Lookup(id int32) (*Artifact, bool) {
	artifact, ok := manager.artifacts[id]
	return artifact, ok
}

func (manager *ArtifactManager) RegisterChunkArtifact(
	chunk *TerrainChunk,
	name string,
	value int,
	centerX float32,
	centerZ float32,
	pixelBounds image.Rectangle,
) *Artifact {
	manager.nextID++
	artifactID := manager.nextID

	artifact := &Artifact{
		ID:          artifactID,
		Name:        name,
		Value:       value,
		CenterX:     centerX,
		CenterZ:     centerZ,
		PixelBounds: pixelBounds,
		Chunk:       chunk,
	}

	manager.artifacts[artifactID] = artifact
	return artifact
}
