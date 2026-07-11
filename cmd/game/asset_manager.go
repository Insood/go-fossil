package main

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"path"
	"path/filepath"
	"slices"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"

	"go-fossil/internal/terrain"
)

//go:embed assets/shaders/*
var shaderAssets embed.FS

//go:embed assets/textures/*
var textureAssets embed.FS

//go:embed assets/levels/*
var levelAssets embed.FS

type AssetManager struct {
	models   map[string]*rl.Model
	shaders  map[string]rl.Shader
	textures map[string]rl.Texture2D
	levels   map[string]terrain.LevelData
	terrains map[string]*TerrainAsset
}

type TerrainAsset struct {
	Model       *rl.Model
	Mesh        rl.Mesh
	BaseTexture rl.Texture2D
	Surface     *terrain.Surface
}

func NewAssetManager() *AssetManager {
	return &AssetManager{
		models:   make(map[string]*rl.Model),
		shaders:  make(map[string]rl.Shader),
		textures: make(map[string]rl.Texture2D),
		levels:   make(map[string]terrain.LevelData),
		terrains: make(map[string]*TerrainAsset),
	}
}

func (assets *AssetManager) Load() {
	assets.loadShaders()
	assets.loadTextures()
	assets.loadLevels()
	assets.loadModels()
}

func (assets *AssetManager) Model(name string) *rl.Model {
	return assets.models[name]
}

func (assets *AssetManager) Shader(name string) rl.Shader {
	return assets.shaders[name]
}

func (assets *AssetManager) Texture(name string) rl.Texture2D {
	return assets.textures[name]
}

func (assets *AssetManager) Level(name string) terrain.LevelData {
	level, ok := assets.levels[name]
	if !ok {
		panic(fmt.Errorf("level asset %q not loaded", name))
	}
	return level
}

func (assets *AssetManager) Terrain(name string) *TerrainAsset {
	terrainAsset, ok := assets.terrains[name]
	if !ok {
		panic(fmt.Errorf("terrain asset %q not loaded", name))
	}
	return terrainAsset
}

func (assets *AssetManager) Unload() {
	for _, terrainAsset := range assets.terrains {
		rl.UnloadTexture(terrainAsset.BaseTexture)
		rl.UnloadMesh(&terrainAsset.Mesh)
	}

	for _, model := range assets.models {
		if assets.isTerrainModel(model) {
			continue
		}
		rl.UnloadModel(*model)
	}
	for _, texture := range assets.textures {
		rl.UnloadTexture(texture)
	}
	for _, shader := range assets.shaders {
		rl.UnloadShader(shader)
	}
}

func (assets *AssetManager) loadShaders() {
	for name, sources := range shaderAssetSources() {
		assets.shaders[name] = rl.LoadShaderFromMemory(sources.vertex, sources.fragment)
	}
}

func (assets *AssetManager) loadTextures() {
	const textureDir = "assets/textures"

	assets.textures["white"] = loadSolidTexture(rl.White)

	for _, fileName := range embeddedAssetNames(textureAssets, textureDir, ".png", ".jpg", ".jpeg") {
		assetPath := path.Join(textureDir, fileName)
		textureName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		texture := loadTextureAsset(assetPath)
		rl.SetTextureWrap(texture, rl.WrapRepeat)
		assets.textures[textureName] = texture
	}
}

func (assets *AssetManager) loadModels() {
	terrainAsset := assets.loadGroundAsset(defaultLevelName)
	assets.terrains[defaultLevelName] = terrainAsset
	assets.models["ground"] = terrainAsset.Model
	assets.models["drone"] = assets.loadDroneModel()
	assets.models["prop_cube"] = assets.loadUnitCubeModel()
	assets.models["prop_sphere"] = assets.loadUnitSphereModel()
}

func (assets *AssetManager) loadLevels() {
	const levelDir = "assets/levels"

	for _, fileName := range embeddedAssetNames(levelAssets, levelDir, ".json") {
		levelPath := path.Join(levelDir, fileName)
		levelName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		level, err := terrain.LoadLevel(levelAssets, levelPath)
		if err != nil {
			panic(fmt.Errorf("load level asset %q: %w", levelPath, err))
		}

		assets.levels[levelName] = level
	}
}

func (assets *AssetManager) loadGroundAsset(levelName string) *TerrainAsset {
	level := assets.Level(levelName)
	tileImages := make(map[string]image.Image, len(level.TileDefinitions))
	for _, tileDefinition := range level.TileDefinitions {
		tileImage, err := loadTextureImageAsset(path.Join("assets/textures", tileDefinition))
		if err != nil {
			panic(fmt.Errorf("load terrain tile image %q: %w", tileDefinition, err))
		}
		tileImages[tileDefinition] = tileImage
	}

	surface, err := terrain.BuildSurface(level, tileImages, terrainTexturePixelsPerTile)
	if err != nil {
		panic(fmt.Errorf("build terrain surface for level %q: %w", levelName, err))
	}

	mesh := newTerrainMesh(surface)
	rl.UploadMesh(&mesh, false)
	ground := rl.LoadModelFromMesh(mesh)
	ground.Materials.Shader = assets.shaders["shadow_receiver"]
	baseTexture := loadTextureFromImage(surface.BaseImage)
	rl.SetMaterialTexture(ground.Materials, rl.MapAlbedo, baseTexture)
	ground.Materials.Shader.UpdateLocation(
		rl.ShaderLocMapHeight,
		rl.GetShaderLocation(ground.Materials.Shader, "shadowMap"),
	)

	return &TerrainAsset{
		Model:       &ground,
		Mesh:        mesh,
		BaseTexture: baseTexture,
		Surface:     surface,
	}
}

func (assets *AssetManager) loadDroneModel() *rl.Model {
	drone := rl.LoadModelFromMesh(rl.GenMeshCube(droneWidth, droneHeight, droneDepth))
	configureShadowReceiverMaterial(&drone, assets)
	return &drone
}

func (assets *AssetManager) loadUnitCubeModel() *rl.Model {
	cube := rl.LoadModelFromMesh(rl.GenMeshCube(1.0, 1.0, 1.0))
	configureShadowReceiverMaterial(&cube, assets)
	return &cube
}

func (assets *AssetManager) loadUnitSphereModel() *rl.Model {
	sphere := rl.LoadModelFromMesh(rl.GenMeshSphere(0.5, 24, 24))
	configureShadowReceiverMaterial(&sphere, assets)
	return &sphere
}

func configureShadowReceiverMaterial(model *rl.Model, assets *AssetManager) {
	model.Materials.Shader = assets.shaders["shadow_receiver"]
	rl.SetMaterialTexture(model.Materials, rl.MapAlbedo, assets.Texture("white"))
	model.Materials.Shader.UpdateLocation(
		rl.ShaderLocMapHeight,
		rl.GetShaderLocation(model.Materials.Shader, "shadowMap"),
	)
}

type shaderFiles struct {
	vertex   string
	fragment string
}

func embeddedAssetNames(fs embed.FS, dir string, extensions ...string) []string {
	entries, err := fs.ReadDir(dir)
	if err != nil {
		panic(fmt.Errorf("read asset dir %q: %w", dir, err))
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if len(extensions) > 0 && !slices.Contains(extensions, ext) {
			continue
		}

		names = append(names, entry.Name())
	}

	slices.Sort(names)
	return names
}

func shaderAssetSources() map[string]shaderFiles {
	const shaderDir = "assets/shaders"

	sources := make(map[string]shaderFiles)
	for _, fileName := range embeddedAssetNames(shaderAssets, shaderDir, ".vs", ".vert", ".fs", ".frag") {
		ext := strings.ToLower(filepath.Ext(fileName))
		stem := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		sourcePath := path.Join(shaderDir, fileName)
		sourceBytes, err := shaderAssets.ReadFile(sourcePath)
		if err != nil {
			panic(fmt.Errorf("read shader source %q: %w", sourcePath, err))
		}
		source := string(sourceBytes)

		switch ext {
		case ".vs", ".vert":
			current := sources[stem]
			current.vertex = source
			sources[stem] = current
		case ".fs", ".frag":
			current := sources[stem]
			current.fragment = source
			sources[stem] = current
		}
	}

	for name, sources := range sources {
		if sources.vertex == "" || sources.fragment == "" {
			panic(fmt.Errorf("shader %q is missing a vertex or fragment file", name))
		}
	}

	return sources
}

func loadTextureAsset(assetPath string) rl.Texture2D {
	data, err := textureAssets.ReadFile(assetPath)
	if err != nil {
		panic(fmt.Errorf("read texture asset %q: %w", assetPath, err))
	}

	image := rl.LoadImageFromMemory(".png", data, int32(len(data)))
	if image == nil {
		panic(fmt.Errorf("load texture image %q from memory", assetPath))
	}
	defer rl.UnloadImage(image)

	return rl.LoadTextureFromImage(image)
}

func loadTextureImageAsset(assetPath string) (image.Image, error) {
	data, err := textureAssets.ReadFile(assetPath)
	if err != nil {
		return nil, fmt.Errorf("read texture asset %q: %w", assetPath, err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode texture asset %q: %w", assetPath, err)
	}

	return img, nil
}

func loadSolidTexture(color rl.Color) rl.Texture2D {
	image := rl.GenImageColor(1, 1, color)
	if image == nil {
		panic("generate solid texture image")
	}
	defer rl.UnloadImage(image)

	return rl.LoadTextureFromImage(image)
}

func loadTextureFromImage(src image.Image) rl.Texture2D {
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, src); err != nil {
		panic(fmt.Errorf("encode texture image: %w", err))
	}

	image := rl.LoadImageFromMemory(".png", encoded.Bytes(), int32(encoded.Len()))
	if image == nil {
		panic("load texture image from encoded memory")
	}
	defer rl.UnloadImage(image)

	return rl.LoadTextureFromImage(image)
}

func newTerrainMesh(surface *terrain.Surface) rl.Mesh {
	mesh := rl.Mesh{
		VertexCount:   int32(len(surface.Vertices) / 3),
		TriangleCount: int32(len(surface.Indices) / 3),
		Vertices:      &surface.Vertices[0],
		Texcoords:     &surface.Texcoords[0],
		Normals:       &surface.Normals[0],
		Indices:       &surface.Indices[0],
	}
	return mesh
}

func (assets *AssetManager) isTerrainModel(model *rl.Model) bool {
	for _, terrainAsset := range assets.terrains {
		if terrainAsset.Model == model {
			return true
		}
	}
	return false
}
