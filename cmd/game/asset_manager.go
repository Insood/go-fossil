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
)

type AssetManager struct {
	artifactDefinitions map[string]*ArtifactDefinition
	images              map[string]image.Image
	models              map[string]*rl.Model
	shaders             map[string]rl.Shader
	textures            map[string]rl.Texture2D
	assetRoot           string
	assetFS             fs.FS
}

func NewAssetManager() *AssetManager {
	assetRoot := runtimeAssetRoot()
	return &AssetManager{
		artifactDefinitions: make(map[string]*ArtifactDefinition),
		images:              make(map[string]image.Image),
		models:              make(map[string]*rl.Model),
		shaders:             make(map[string]rl.Shader),
		textures:            make(map[string]rl.Texture2D),
		assetRoot:           assetRoot,
		assetFS:             os.DirFS(assetRoot),
	}
}

func (assets *AssetManager) Load() {
	assets.loadShaders()
	assets.loadImages()
	assets.loadArtifactDefinitions()
	assets.loadTextures()
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
	img, ok := assets.images[normalizedPath]
	if !ok {
		img = assets.mustLoadImage(normalizedPath)
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

func (assets *AssetManager) Unload() {
	for _, model := range assets.models {
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
		assetPath := path.Join(textureDir, fileName)
		textureName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		texture := loadTextureFromImage(assets.Image(assetPath))
		rl.SetTextureWrap(texture, rl.WrapRepeat)
		assets.textures[textureName] = texture
	}
}

func (assets *AssetManager) loadModels() {
	assets.models["drone"] = assets.loadDroneModel()
	assets.models["prop_cube"] = assets.loadUnitCubeModel()
	assets.models["prop_sphere"] = assets.loadUnitSphereModel()
}

func (assets *AssetManager) mustLoadImage(assetPath string) image.Image {
	img, err := loadTextureImageAsset(assets.assetPath(assetPath))
	if err != nil {
		panic(fmt.Errorf("load image asset %q: %w", assetPath, err))
	}

	assets.images[assetPath] = img
	return img
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

func (assets *AssetManager) assetPath(parts ...string) string {
	return filepath.Join(append([]string{assets.assetRoot}, parts...)...)
}
