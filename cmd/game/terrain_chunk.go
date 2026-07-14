package main

import (
	"image"

	rl "github.com/gen2brain/raylib-go/raylib"

	"go-fossil/internal/terrain"
)

type TerrainChunk struct {
	Data           terrain.ChunkData
	SurfaceMesh    *terrain.SurfaceMesh
	SurfaceTexture *terrain.SurfaceTexture
	Model          *rl.Model
	Mesh           rl.Mesh
	BaseTexture    rl.Texture2D
	OverlayImage   *image.RGBA
	OverlayTexture rl.Texture2D
}

func (chunk *TerrainChunk) Center() rl.Vector3 {
	return rl.NewVector3(float32(chunk.Data.Width)/2, 0, float32(chunk.Data.Height)/2)
}

func (chunk *TerrainChunk) Unload() {
	if chunk == nil {
		return
	}

	rl.UnloadTexture(chunk.BaseTexture)
	rl.UnloadMesh(&chunk.Mesh)
}
