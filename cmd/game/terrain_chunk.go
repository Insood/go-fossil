package main

import (
	"image"

	rl "github.com/gen2brain/raylib-go/raylib"

	"go-fossil/internal/terrain"
)

type TerrainChunk struct {
	Coords         ChunkCoords
	OriginX        float32
	OriginZ        float32
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
	return rl.NewVector3(
		chunk.OriginX+float32(chunk.Data.Width)/2,
		0,
		chunk.OriginZ+float32(chunk.Data.Height)/2,
	)
}

func (chunk *TerrainChunk) ContainsWorldPosition(worldX, worldZ float32) bool {
	return worldX >= chunk.OriginX &&
		worldX <= chunk.OriginX+float32(chunk.Data.Width) &&
		worldZ >= chunk.OriginZ &&
		worldZ <= chunk.OriginZ+float32(chunk.Data.Height)
}

func (chunk *TerrainChunk) HeightAtWorldPosition(worldX, worldZ float32) float32 {
	return chunk.SurfaceMesh.SampleHeight(worldX-chunk.OriginX, worldZ-chunk.OriginZ)
}

func (chunk *TerrainChunk) Unload() {
	if chunk == nil {
		return
	}

	rl.UnloadTexture(chunk.BaseTexture)
	rl.UnloadMesh(&chunk.Mesh)
}
