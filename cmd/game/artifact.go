package main

import "image"

type Artifact struct {
	ID          int32
	Name        string
	Value       int
	CenterX     float32
	CenterZ     float32
	PixelBounds image.Rectangle
	Chunk       *TerrainChunk
}
