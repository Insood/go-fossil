package main

import (
	"fmt"
	"image"
	"path"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"

	"go-fossil/internal/terrain"
)

type ChunkManager struct {
	world           *ecs.World
	assets          *AssetManager
	artifactManager *ArtifactManager
	terrainChunkMap *ecs.Map1[TerrainChunkComponent]
	chunkData       map[ChunkCoords]terrain.ChunkData
	chunks          map[ChunkCoords]*TerrainChunk
}

func NewChunkManager(world *ecs.World, assets *AssetManager, artifactManager *ArtifactManager) *ChunkManager {
	return &ChunkManager{
		world:           world,
		assets:          assets,
		artifactManager: artifactManager,
		terrainChunkMap: ecs.NewMap1[TerrainChunkComponent](world),
		chunkData:       make(map[ChunkCoords]terrain.ChunkData),
		chunks:          make(map[ChunkCoords]*TerrainChunk),
	}
}

func (manager *ChunkManager) LoadDiskChunk(coords ChunkCoords, chunkName string) *TerrainChunk {
	if chunk, ok := manager.chunks[coords]; ok {
		return chunk
	}

	chunkData := manager.loadChunkData(coords, chunkName)
	chunk := manager.buildChunk(coords, chunkData)
	manager.chunks[coords] = chunk
	return chunk
}

func (manager *ChunkManager) LoadGeneratedChunk(coords ChunkCoords, generator *ChunkGenerator) *TerrainChunk {
	if chunk, ok := manager.chunks[coords]; ok {
		return chunk
	}

	chunk := manager.buildChunk(coords, generator.GenerateFlat(coords))
	manager.chunks[coords] = chunk
	return chunk
}

func (manager *ChunkManager) Chunk(coords ChunkCoords) *TerrainChunk {
	chunk, ok := manager.chunks[coords]
	if !ok {
		panic("terrain chunk not loaded: " + coords.String())
	}
	return chunk
}

func (manager *ChunkManager) ChunkForWorldPosition(worldX, worldZ float32) (*TerrainChunk, bool) {
	preferredCoords := chunkCoordsForWorldPosition(worldX, worldZ)
	if chunk, ok := manager.chunks[preferredCoords]; ok && chunk.ContainsWorldPosition(worldX, worldZ) {
		return chunk, true
	}

	var candidate *TerrainChunk
	for _, chunk := range manager.chunks {
		if !chunk.ContainsWorldPosition(worldX, worldZ) {
			continue
		}

		if candidate == nil || chunk.OriginX > candidate.OriginX || (chunk.OriginX == candidate.OriginX && chunk.OriginZ > candidate.OriginZ) {
			candidate = chunk
		}
	}

	if candidate == nil {
		return nil, false
	}

	return candidate, true
}

func (manager *ChunkManager) SampleHeight(worldX, worldZ float32) float32 {
	chunk, ok := manager.ChunkForWorldPosition(worldX, worldZ)
	if !ok {
		return 0
	}

	return chunk.HeightAtWorldPosition(worldX, worldZ)
}

func (manager *ChunkManager) DroneFitsAtWorldPosition(worldX, worldZ float32) bool {
	halfWidth := float32(droneWidth / 2)
	halfDepth := float32(droneDepth / 2)

	corners := [4]rl.Vector2{
		rl.NewVector2(worldX-halfWidth, worldZ-halfDepth),
		rl.NewVector2(worldX+halfWidth, worldZ-halfDepth),
		rl.NewVector2(worldX-halfWidth, worldZ+halfDepth),
		rl.NewVector2(worldX+halfWidth, worldZ+halfDepth),
	}

	for _, corner := range corners {
		if _, ok := manager.ChunkForWorldPosition(corner.X, corner.Y); !ok {
			return false
		}
	}

	return true
}

func (manager *ChunkManager) Chunks() []*TerrainChunk {
	chunks := make([]*TerrainChunk, 0, len(manager.chunks))
	for _, chunk := range manager.chunks {
		chunks = append(chunks, chunk)
	}

	slices.SortFunc(chunks, func(a, b *TerrainChunk) int {
		if a.Coords.Z != b.Coords.Z {
			return a.Coords.Z - b.Coords.Z
		}
		return a.Coords.X - b.Coords.X
	})

	return chunks
}

func (manager *ChunkManager) Unload() {
	for _, chunk := range manager.chunks {
		if chunk.Entity != (ecs.Entity{}) && manager.world != nil && manager.world.Alive(chunk.Entity) {
			manager.world.RemoveEntity(chunk.Entity)
			chunk.Entity = ecs.Entity{}
		}
		chunk.Unload()
	}
}

func (manager *ChunkManager) buildChunk(coords ChunkCoords, chunkData terrain.ChunkData) *TerrainChunk {
	originX, originZ := chunkOriginForCoords(coords)
	chunk := &TerrainChunk{
		Coords:  coords,
		OriginX: originX,
		OriginZ: originZ,
		Data:    chunkData,
	}

	tileImages := make(map[string]image.Image, len(chunkData.TileDefinitions))
	for _, tileDefinition := range chunkData.TileDefinitions {
		tileImages[tileDefinition] = Must(manager.assets.LookupImage(path.Join("textures", tileDefinition)))
	}

	surfaceMesh, err := terrain.BuildSurfaceMesh(chunkData)
	if err != nil {
		panic(fmt.Errorf("build terrain mesh for chunk %s: %w", coords.String(), err))
	}

	surfaceTexture, err := terrain.BuildSurfaceTexture(chunkData, tileImages, terrainTexturePixelsPerTile)
	if err != nil {
		panic(fmt.Errorf("build terrain texture for chunk %s: %w", coords.String(), err))
	}
	artifactImage := buildArtifactImageLayer(chunkData, manager.assets, surfaceTexture.BaseImage.Bounds())
	artifactData := buildArtifactDataLayer(manager.artifactManager, chunk, manager.assets, surfaceTexture.BaseImage.Bounds())

	mesh := newTerrainMesh(surfaceMesh)
	rl.UploadMesh(&mesh, false)
	model := rl.LoadModelFromMesh(mesh)
	model.Materials.Shader = Must(manager.assets.LookupShader("shadow_receiver"))
	baseTexture := loadTextureFromImage(surfaceTexture.BaseImage)
	rl.SetMaterialTexture(model.Materials, rl.MapAlbedo, baseTexture)
	artifactTexture := loadTextureFromImage(artifactImage)
	rl.SetMaterialTexture(model.Materials, rl.MapEmission, artifactTexture)
	burnOverlayImage := image.NewRGBA(surfaceTexture.BaseImage.Bounds())
	burnOverlayTexture := loadTextureFromImage(burnOverlayImage)
	rl.SetMaterialTexture(model.Materials, rl.MapOcclusion, burnOverlayTexture)
	model.Materials.Shader.UpdateLocation(
		rl.ShaderLocMapHeight,
		rl.GetShaderLocation(model.Materials.Shader, "shadowMap"),
	)
	model.Materials.Shader.UpdateLocation(
		rl.ShaderLocMapEmission,
		rl.GetShaderLocation(model.Materials.Shader, "texture1"),
	)

	chunk.SurfaceMesh = surfaceMesh
	chunk.SurfaceTexture = surfaceTexture
	chunk.Model = &model
	chunk.Mesh = mesh
	chunk.BaseTexture = baseTexture
	chunk.ArtifactImage = artifactImage
	chunk.ArtifactData = artifactData
	chunk.ArtifactTexture = artifactTexture
	chunk.BurnOverlayImage = burnOverlayImage
	chunk.BurnOverlayTexture = burnOverlayTexture

	manager.registerTerrainChunkEntity(chunk)
	return chunk
}

func (manager *ChunkManager) BurnAtWorldPosition(worldX, worldZ float32) bool {
	chunk, ok := manager.ChunkForWorldPosition(worldX, worldZ)
	if !ok {
		return false
	}

	return chunk.AddBurnMark(worldX, worldZ)
}

func (manager *ChunkManager) registerTerrainChunkEntity(chunk *TerrainChunk) {
	if manager == nil || manager.terrainChunkMap == nil || chunk == nil {
		return
	}

	chunk.Entity = manager.terrainChunkMap.NewEntity(&TerrainChunkComponent{
		Chunk: chunk,
	})
}

func (manager *ChunkManager) loadChunkData(coords ChunkCoords, chunkName string) terrain.ChunkData {
	if chunk, ok := manager.chunkData[coords]; ok {
		return chunk
	}

	chunkPath := path.Join("terrain_chunks", chunkName+".json")
	chunk, err := terrain.LoadChunkData(manager.assets.assetFS, chunkPath)
	if err != nil {
		panic(fmt.Errorf("load terrain chunk %q: %w", chunkPath, err))
	}

	manager.chunkData[coords] = chunk
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
