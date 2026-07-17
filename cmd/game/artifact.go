package main

import "image"

type Artifact struct {
	ID          int32
	Name        string
	Value       int
	Size        int
	CenterX     float32
	CenterZ     float32
	PixelBounds image.Rectangle
	Chunk       *TerrainChunk
}
