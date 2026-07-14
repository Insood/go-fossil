package main

import (
	"fmt"
	"image"
	"path"

	rl "github.com/gen2brain/raylib-go/raylib"

	"go-fossil/internal/terrain"
)

type ChunkManager struct {
	assets    *AssetManager
	chunkData map[string]terrain.ChunkData
	chunks    map[string]*TerrainChunk
}

func NewChunkManager(assets *AssetManager) *ChunkManager {
	return &ChunkManager{
		assets:    assets,
		chunkData: make(map[string]terrain.ChunkData),
		chunks:    make(map[string]*TerrainChunk),
	}
}

func (manager *ChunkManager) LoadChunk(chunkName string) *TerrainChunk {
	if chunk, ok := manager.chunks[chunkName]; ok {
		return chunk
	}

	chunk := manager.buildChunk(chunkName)
	manager.chunks[chunkName] = chunk
	return chunk
}

func (manager *ChunkManager) Chunk(chunkName string) *TerrainChunk {
	chunk, ok := manager.chunks[chunkName]
	if !ok {
		panic("terrain chunk not loaded: " + chunkName)
	}
	return chunk
}

func (manager *ChunkManager) Unload() {
	for _, chunk := range manager.chunks {
		chunk.Unload()
	}
}

func (manager *ChunkManager) buildChunk(chunkName string) *TerrainChunk {
	chunk := manager.loadChunkData(chunkName)
	tileImages := make(map[string]image.Image, len(chunk.TileDefinitions))
	for _, tileDefinition := range chunk.TileDefinitions {
		tileImages[tileDefinition] = manager.assets.Image(path.Join("textures", tileDefinition))
	}

	surfaceMesh, err := terrain.BuildSurfaceMesh(chunk)
	if err != nil {
		panic(fmt.Errorf("build terrain mesh for chunk %q: %w", chunkName, err))
	}

	surfaceTexture, err := terrain.BuildSurfaceTexture(chunk, tileImages, terrainTexturePixelsPerTile)
	if err != nil {
		panic(fmt.Errorf("build terrain texture for chunk %q: %w", chunkName, err))
	}

	mesh := newTerrainMesh(surfaceMesh)
	rl.UploadMesh(&mesh, false)
	model := rl.LoadModelFromMesh(mesh)
	model.Materials.Shader = manager.assets.Shader("shadow_receiver")
	baseTexture := loadTextureFromImage(surfaceTexture.BaseImage)
	rl.SetMaterialTexture(model.Materials, rl.MapAlbedo, baseTexture)
	model.Materials.Shader.UpdateLocation(
		rl.ShaderLocMapHeight,
		rl.GetShaderLocation(model.Materials.Shader, "shadowMap"),
	)

	return &TerrainChunk{
		Data:           chunk,
		SurfaceMesh:    surfaceMesh,
		SurfaceTexture: surfaceTexture,
		Model:          &model,
		Mesh:           mesh,
		BaseTexture:    baseTexture,
		OverlayImage:   image.NewRGBA(surfaceTexture.BaseImage.Bounds()),
	}
}

func (manager *ChunkManager) loadChunkData(chunkName string) terrain.ChunkData {
	if chunk, ok := manager.chunkData[chunkName]; ok {
		return chunk
	}

	chunkPath := path.Join("terrain_chunks", chunkName+".json")
	chunk, err := terrain.LoadChunkData(manager.assets.assetFS, chunkPath)
	if err != nil {
		panic(fmt.Errorf("load terrain chunk %q: %w", chunkPath, err))
	}

	manager.chunkData[chunkName] = chunk
	return chunk
}

func newTerrainMesh(surface *terrain.SurfaceMesh) rl.Mesh {
	return rl.Mesh{
		VertexCount:   int32(len(surface.Vertices) / 3),
		TriangleCount: int32(len(surface.Indices) / 3),
		Vertices:      &surface.Vertices[0],
		Texcoords:     &surface.Texcoords[0],
		Normals:       &surface.Normals[0],
		Indices:       &surface.Indices[0],
	}
}
