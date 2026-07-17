package main

import (
	"fmt"
	"image"
	"math/rand"
	"path"
	"slices"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"

	"go-fossil/internal/terrain"
)

type ChunkManager struct {
	world           *ecs.World
	assets          *AssetManager
	artifactManager *ArtifactManager
	rng             *rand.Rand
	terrainChunkMap *ecs.Map1[TerrainChunkComponent]
	chunkData       map[ChunkCoords]terrain.ChunkData
	chunks          map[ChunkCoords]*TerrainChunk
}

func NewChunkManager(world *ecs.World, assets *AssetManager, artifactManager *ArtifactManager) *ChunkManager {
	return &ChunkManager{
		world:           world,
		assets:          assets,
		artifactManager: artifactManager,
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())),
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

	chunkData := generator.GenerateFlat(coords)
	manager.addGeneratedChunkArtifacts(coords, &chunkData)
	chunk := manager.buildChunk(coords, chunkData)
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

func (manager *ChunkManager) addGeneratedChunkArtifacts(coords ChunkCoords, chunkData *terrain.ChunkData) {
	if manager == nil || chunkData == nil || manager.assets == nil {
		return
	}

	definitions := manager.assets.ArtifactDefinitions()
	if len(definitions) == 0 {
		return
	}

	placements := randomGeneratedArtifactPlacements(chunkData.Width, chunkData.Height, definitions, manager.rng)
	chunkData.Artifacts = append(chunkData.Artifacts, placements...)
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

func randomGeneratedArtifactPlacements(
	width int,
	height int,
	definitions []*ArtifactDefinition,
	rng *rand.Rand,
) []terrain.ArtifactPlacement {
	if len(definitions) == 0 {
		return nil
	}
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	maxCount := generatedArtifactMaxCount
	if len(definitions) < maxCount {
		maxCount = len(definitions)
	}
	if maxCount <= 0 {
		return nil
	}

	count := 1 + rng.Intn(maxCount)
	order := rng.Perm(len(definitions))

	bounds := image.Rect(0, 0, width*terrainTexturePixelsPerTile, height*terrainTexturePixelsPerTile)
	minX := float32(bounds.Min.X + generatedArtifactEdgeMarginPixels)
	maxX := float32(bounds.Max.X - generatedArtifactEdgeMarginPixels)
	minZ := float32(bounds.Min.Y + generatedArtifactEdgeMarginPixels)
	maxZ := float32(bounds.Max.Y - generatedArtifactEdgeMarginPixels)
	if maxX <= minX || maxZ <= minZ {
		return nil
	}

	placements := make([]terrain.ArtifactPlacement, 0, count)
	centers := make([][2]float32, 0, count)
	for _, index := range order {
		if len(placements) >= count {
			break
		}

		definition := definitions[index]
		if definition == nil {
			continue
		}

		placement, ok := randomArtifactPlacement(definition.Name, minX, maxX, minZ, maxZ, centers, rng)
		if !ok {
			continue
		}

		placements = append(placements, placement)
		centers = append(centers, [2]float32{placement.X, placement.Z})
	}

	return placements
}

func randomArtifactPlacement(
	name string,
	minX, maxX, minZ, maxZ float32,
	centers [][2]float32,
	rng *rand.Rand,
) (terrain.ArtifactPlacement, bool) {
	for attempt := 0; attempt < generatedArtifactPlacementAttempts; attempt++ {
		x := minX + float32(rng.Float64())*(maxX-minX)
		z := minZ + float32(rng.Float64())*(maxZ-minZ)
		if tooCloseToAnyCenter(x, z, centers) {
			continue
		}

		return terrain.ArtifactPlacement{
			Name:        name,
			X:           x,
			Z:           z,
			Orientation: float32(rng.Intn(360)),
		}, true
	}

	return terrain.ArtifactPlacement{}, false
}

func tooCloseToAnyCenter(x, z float32, centers [][2]float32) bool {
	minGapSquared := float32(generatedArtifactMinCenterGapPixels * generatedArtifactMinCenterGapPixels)
	for _, center := range centers {
		dx := x - center[0]
		dz := z - center[1]
		if dx*dx+dz*dz < minGapSquared {
			return true
		}
	}

	return false
}
