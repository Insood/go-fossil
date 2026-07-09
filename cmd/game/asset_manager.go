package main

import (
	"embed"
	"fmt"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

//go:embed assets/shaders/*
var shaderAssets embed.FS

//go:embed assets/textures/*
var textureAssets embed.FS

type AssetManager struct {
	models   map[string]*rl.Model
	shaders  map[string]rl.Shader
	textures map[string]rl.Texture2D
}

func NewAssetManager() *AssetManager {
	return &AssetManager{
		models:   make(map[string]*rl.Model),
		shaders:  make(map[string]rl.Shader),
		textures: make(map[string]rl.Texture2D),
	}
}

func (assets *AssetManager) Load() {
	assets.loadShaders()
	assets.loadTextures()
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
	for name, sources := range shaderAssetSources() {
		assets.shaders[name] = rl.LoadShaderFromMemory(sources.vertex, sources.fragment)
	}
}

func (assets *AssetManager) loadTextures() {
	const textureDir = "assets/textures"

	for _, fileName := range embeddedAssetNames(textureAssets, textureDir, ".png", ".jpg", ".jpeg") {
		assetPath := path.Join(textureDir, fileName)
		textureName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		texture := loadTextureAsset(assetPath)
		rl.SetTextureWrap(texture, rl.WrapRepeat)
		assets.textures[textureName] = texture
	}
}

func (assets *AssetManager) loadModels() {
	assets.models["ground"] = assets.loadGroundModel()
	assets.models["drone"] = assets.loadDroneModel()
}

func (assets *AssetManager) loadGroundModel() *rl.Model {
	mesh := rl.GenMeshPlane(gridSize, gridSize, gridSubdivisions, gridSubdivisions)
	tileMeshTexcoords(&mesh, float32(gridSize), float32(gridSize))

	ground := rl.LoadModelFromMesh(mesh)
	ground.Materials.Shader = assets.shaders["shadow_receiver"]
	rl.SetMaterialTexture(ground.Materials, rl.MapAlbedo, assets.Texture("ground_grid"))
	ground.Materials.Shader.UpdateLocation(
		rl.ShaderLocMapHeight,
		rl.GetShaderLocation(ground.Materials.Shader, "shadowMap"),
	)
	return &ground
}

func (assets *AssetManager) loadDroneModel() *rl.Model {
	drone := rl.LoadModelFromMesh(rl.GenMeshCube(droneWidth, droneHeight, droneDepth))
	return &drone
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

func tileMeshTexcoords(mesh *rl.Mesh, uScale, vScale float32) {
	if mesh.Texcoords == nil {
		return
	}

	texcoords := unsafe.Slice(mesh.Texcoords, mesh.VertexCount*2)
	for i := int32(0); i < mesh.VertexCount; i++ {
		texcoords[i*2] *= uScale
		texcoords[i*2+1] *= vScale
	}

	texcoordBytes := unsafe.Slice(
		(*byte)(unsafe.Pointer(mesh.Texcoords)),
		int(mesh.VertexCount*2)*int(unsafe.Sizeof(float32(0))),
	)
	rl.UpdateMeshBuffer(*mesh, 1, texcoordBytes, 0)
}
