package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"

	"go-fossil/internal/terrain"
)

type AssetManager struct {
	artifactDefinitions map[string]*ArtifactDefinition
	models    map[string]*rl.Model
	shaders   map[string]rl.Shader
	textures  map[string]rl.Texture2D
	levels    map[string]terrain.LevelData
	terrains  map[string]*TerrainAsset
	assetRoot string
	assetFS   fs.FS
}

type TerrainAsset struct {
	Model       *rl.Model
	Mesh        rl.Mesh
	BaseTexture rl.Texture2D
	Surface     *terrain.Surface
}

func NewAssetManager() *AssetManager {
	assetRoot := runtimeAssetRoot()
	return &AssetManager{
		artifactDefinitions: make(map[string]*ArtifactDefinition),
		images:              make(map[string]image.Image),
		shaders:   make(map[string]rl.Shader),
		textures:  make(map[string]rl.Texture2D),
		levels:    make(map[string]terrain.LevelData),
		terrains:  make(map[string]*TerrainAsset),
		assetRoot: assetRoot,
		assetFS:   os.DirFS(assetRoot),
	}
}

func (assets *AssetManager) Load() {
	assets.loadShaders()
	assets.loadImages()
	assets.loadArtifactDefinitions()
	assets.loadTextures()
	assets.loadLevels()
	assets.loadModels()
}

func (assets *AssetManager) ArtifactDefinition(name string) *ArtifactDefinition {
	definition, ok := assets.artifactDefinitions[name]
	if !ok {
		panic(fmt.Errorf("artifact definition %q not loaded", name))
	}
	return definition
}

func (assets *AssetManager) Image(assetPath string) image.Image {
	normalizedPath := path.Clean(assetPath)
	if !ok {
		panic(fmt.Errorf("image asset %q not loaded", name))
	}
	return img
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
	for name, sources := range assets.shaderAssetSources() {
		assets.shaders[name] = rl.LoadShaderFromMemory(sources.vertex, sources.fragment)
	}
}

func (assets *AssetManager) loadImages() {
	const textureDir = "textures"

	for _, fileName := range assetFileNames(assets.assetFS, textureDir, ".png", ".jpg", ".jpeg") {
		assetPath := path.Join(textureDir, fileName)
		assets.images[assetPath] = assets.mustLoadImage(assetPath)
	}
}

func (assets *AssetManager) loadArtifactDefinitions() {
	const artifactDir = "artifacts"

	for _, fileName := range assetFileNames(assets.assetFS, artifactDir, ".json") {
		definitionPath := path.Join(artifactDir, fileName)
		definition, err := loadArtifactDefinitionAsset(assets.assetFS, definitionPath)
		if err != nil {
			panic(fmt.Errorf("load artifact definition asset %q: %w", definitionPath, err))
		}
		if _, exists := assets.artifactDefinitions[definition.Name]; exists {
			panic(fmt.Errorf("artifact definition %q declared more than once", definition.Name))
		}

		assets.Image(definition.ImagePath)
		definitionCopy := definition
		assets.artifactDefinitions[definitionCopy.Name] = &definitionCopy
	}
}

func (assets *AssetManager) loadTextures() {
	const textureDir = "textures"

	assets.textures["white"] = loadSolidTexture(rl.White)

	for _, fileName := range assetFileNames(assets.assetFS, textureDir, ".png", ".jpg", ".jpeg") {
		textureName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		texture := loadTextureFromImage(assets.Image(textureName))
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
	const levelDir = "levels"

	for _, fileName := range assetFileNames(assets.assetFS, levelDir, ".json") {
		levelPath := path.Join(levelDir, fileName)
		levelName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		level, err := terrain.LoadLevel(assets.assetFS, levelPath)
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
		tileImage := assets.Image(strings.TrimSuffix(tileDefinition, filepath.Ext(tileDefinition)))
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
	drone := rl.LoadModel(assets.assetPath("models", "drone.glb"))
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
	materials := model.GetMaterials()
	for i := range materials {
		materials[i].Shader = assets.shaders["shadow_receiver"]
		materials[i].GetMap(rl.MapAlbedo).Color = rl.White
		if materials[i].GetMap(rl.MapAlbedo).Texture.ID == 0 {
			rl.SetMaterialTexture(&materials[i], rl.MapAlbedo, assets.Texture("white"))
		}
		materials[i].Shader.UpdateLocation(
			rl.ShaderLocMapHeight,
			rl.GetShaderLocation(materials[i].Shader, "shadowMap"),
		)
	}
}

type shaderFiles struct {
	vertex   string
	fragment string
}

func runtimeAssetRoot() string {
	executablePath, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("resolve executable path: %w", err))
	}

	resolvedPath, err := filepath.EvalSymlinks(executablePath)
	if err == nil {
		executablePath = resolvedPath
	}

	return filepath.Join(filepath.Dir(executablePath), "assets")
}

func assetFileNames(assetFS fs.FS, dir string, extensions ...string) []string {
	entries, err := fs.ReadDir(assetFS, dir)
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

func (assets *AssetManager) shaderAssetSources() map[string]shaderFiles {
	const shaderDir = "shaders"

	sources := make(map[string]shaderFiles)
	for _, fileName := range assetFileNames(assets.assetFS, shaderDir, ".vs", ".vert", ".fs", ".frag") {
		ext := strings.ToLower(filepath.Ext(fileName))
		stem := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		sourcePath := path.Join(shaderDir, fileName)
		sourceBytes, err := fs.ReadFile(assets.assetFS, sourcePath)
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

func loadTextureImageAsset(assetPath string) (image.Image, error) {
	data, err := os.ReadFile(assetPath)
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

func (assets *AssetManager) assetPath(parts ...string) string {
	return filepath.Join(append([]string{assets.assetRoot}, parts...)...)
}
