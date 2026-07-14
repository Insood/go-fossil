package main

import (
	"image"

	rl "github.com/gen2brain/raylib-go/raylib"

	"go-fossil/internal/terrain"
)

type TerrainChunk struct {
	Level          terrain.LevelData
	SurfaceMesh    *terrain.SurfaceMesh
	SurfaceTexture *terrain.SurfaceTexture
	Model          *rl.Model
	Mesh           rl.Mesh
	BaseTexture    rl.Texture2D
	OverlayImage   *image.RGBA
	OverlayTexture rl.Texture2D
}

func (chunk *TerrainChunk) Center() rl.Vector3 {
	return rl.NewVector3(float32(chunk.Level.Width)/2, 0, float32(chunk.Level.Height)/2)
}
