package main

import (
	"fmt"
	"image"
	"path"

	rl "github.com/gen2brain/raylib-go/raylib"

	"go-fossil/internal/terrain"
)

type ChunkManager struct {
	assets *AssetManager
	chunks map[string]*TerrainChunk
}

func NewChunkManager(assets *AssetManager) *ChunkManager {
	return &ChunkManager{
		assets: assets,
		chunks: make(map[string]*TerrainChunk),
	}
}

func (manager *ChunkManager) LoadChunk(levelName string) *TerrainChunk {
	if chunk, ok := manager.chunks[levelName]; ok {
		return chunk
	}

	chunk := manager.buildChunk(levelName)
	manager.chunks[levelName] = chunk
	return chunk
}

func (manager *ChunkManager) Chunk(levelName string) *TerrainChunk {
	chunk, ok := manager.chunks[levelName]
	if !ok {
		panic("terrain chunk not loaded: " + levelName)
	}
	return chunk
}

func (manager *ChunkManager) Unload() {
	for _, chunk := range manager.chunks {
		chunk.Unload()
	}
}

func (manager *ChunkManager) buildChunk(levelName string) *TerrainChunk {
	level := manager.assets.Level(levelName)
	tileImages := make(map[string]image.Image, len(level.TileDefinitions))
	for _, tileDefinition := range level.TileDefinitions {
		tileImages[tileDefinition] = manager.assets.Image(path.Join("textures", tileDefinition))
	}

	surfaceMesh, err := terrain.BuildSurfaceMesh(level)
	if err != nil {
		panic(fmt.Errorf("build terrain mesh for level %q: %w", levelName, err))
	}

	surfaceTexture, err := terrain.BuildSurfaceTexture(level, tileImages, terrainTexturePixelsPerTile)
	if err != nil {
		panic(fmt.Errorf("build terrain texture for level %q: %w", levelName, err))
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
		Level:          level,
		SurfaceMesh:    surfaceMesh,
		SurfaceTexture: surfaceTexture,
		Model:          &model,
		Mesh:           mesh,
		BaseTexture:    baseTexture,
		OverlayImage:   image.NewRGBA(surfaceTexture.BaseImage.Bounds()),
	}
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
