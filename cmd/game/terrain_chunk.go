package main

import (
	"image"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"

	"go-fossil/internal/terrain"
)

type TerrainChunk struct {
	Coords             ChunkCoords
	OriginX            float32
	OriginZ            float32
	Data               terrain.ChunkData
	SurfaceMesh        *terrain.SurfaceMesh
	SurfaceTexture     *terrain.SurfaceTexture
	Model              *rl.Model
	Mesh               rl.Mesh
	BaseTexture        rl.Texture2D
	ArtifactImage      *image.RGBA
	ArtifactData       *ArtifactData
	Artifacts          []Artifact
	ArtifactTexture    rl.Texture2D
	BurnOverlayImage   *image.RGBA
	BurnOverlayTexture rl.Texture2D
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

	if chunk.Model != nil {
		// The model wraps mesh buffers that originated from Go slices, so we
		// strip those pointers before UnloadModel. raylib will then free the
		// C-side model/material allocations without trying to free Go memory.
		sanitizeModelMeshesForUnload(chunk.Model)
	}

	if chunk.Mesh.VaoID != 0 {
		rl.UnloadMesh(&chunk.Mesh)
	}

	if chunk.Model != nil {
		rl.UnloadModel(*chunk.Model)
		chunk.Model = nil
	}

	rl.UnloadTexture(chunk.BaseTexture)
	rl.UnloadTexture(chunk.ArtifactTexture)
	rl.UnloadTexture(chunk.BurnOverlayTexture)
	chunk.BaseTexture = rl.Texture2D{}
	chunk.ArtifactTexture = rl.Texture2D{}
	chunk.BurnOverlayTexture = rl.Texture2D{}
	chunk.Mesh = rl.Mesh{}
}

// sanitizeModelMeshesForUnload clears mesh pointers that point at Go-owned
// buffers so raylib's model teardown only releases its own allocations.
func sanitizeModelMeshesForUnload(model *rl.Model) {
	meshes := model.GetMeshes()
	for i := range meshes {
		meshes[i].Vertices = nil
		meshes[i].Texcoords = nil
		meshes[i].Texcoords2 = nil
		meshes[i].Normals = nil
		meshes[i].Tangents = nil
		meshes[i].Colors = nil
		meshes[i].Indices = nil
		meshes[i].BoneIndices = nil
		meshes[i].BoneWeights = nil
		meshes[i].AnimVertices = nil
		meshes[i].AnimNormals = nil
		meshes[i].VaoID = 0
		meshes[i].VboID = nil
	}
}

func (chunk *TerrainChunk) BurnOverlayBounds() image.Rectangle {
	if chunk == nil || chunk.BurnOverlayImage == nil {
		return image.Rectangle{}
	}

	return chunk.BurnOverlayImage.Bounds()
}

func (chunk *TerrainChunk) AddBurnMark(worldX, worldZ float32) bool {
	if chunk == nil || chunk.BurnOverlayImage == nil {
		return false
	}

	bounds := chunk.BurnOverlayImage.Bounds()
	if bounds.Empty() {
		return false
	}

	localX := worldX - chunk.OriginX
	localZ := worldZ - chunk.OriginZ
	if localX < 0 || localZ < 0 || localX > float32(chunk.Data.Width) || localZ > float32(chunk.Data.Height) {
		return false
	}

	pixelX := int(localX * float32(bounds.Dx()) / float32(chunk.Data.Width))
	pixelY := int(localZ * float32(bounds.Dy()) / float32(chunk.Data.Height))
	pixelX = clampInt(pixelX, bounds.Min.X, bounds.Max.X-1)
	pixelY = clampInt(pixelY, bounds.Min.Y, bounds.Max.Y-1)
	return chunk.paintBurnMark(pixelX, pixelY)
}

func (chunk *TerrainChunk) paintBurnMark(x, y int) bool {
	if chunk == nil || chunk.BurnOverlayImage == nil {
		return false
	}

	bounds := chunk.BurnOverlayImage.Bounds()
	maskBounds := image.Rect(x, y, x+2, y+2)
	clippedBounds := maskBounds.Intersect(bounds)
	if clippedBounds.Empty() {
		return false
	}

	for py := clippedBounds.Min.Y; py < clippedBounds.Max.Y; py++ {
		for px := clippedBounds.Min.X; px < clippedBounds.Max.X; px++ {
			chunk.BurnOverlayImage.SetRGBA(px, py, color.RGBA{A: 255})
		}
	}

	chunk.uploadBurnOverlayRect(clippedBounds)
	return true
}

func (chunk *TerrainChunk) uploadBurnOverlayRect(rect image.Rectangle) {
	if chunk.BurnOverlayTexture.ID == 0 || rect.Empty() {
		return
	}

	pixels := make([]color.RGBA, 0, rect.Dx()*rect.Dy())
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			pixels = append(pixels, chunk.BurnOverlayImage.RGBAAt(x, y))
		}
	}

	rl.UpdateTextureRec(
		chunk.BurnOverlayTexture,
		rl.NewRectangle(float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy())),
		pixels,
	)
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
