package main

import (
	"embed"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

//go:embed assets/shaders/*
var shaderAssets embed.FS

type AssetManager struct {
	models  map[string]*rl.Model
	shaders map[string]rl.Shader
}

func NewAssetManager() *AssetManager {
	return &AssetManager{
		models:  make(map[string]*rl.Model),
		shaders: make(map[string]rl.Shader),
	}
}

func (assets *AssetManager) Load() {
	assets.loadShaders()
	assets.loadModels()
}

func (assets *AssetManager) Model(name string) *rl.Model {
	return assets.models[name]
}

func (assets *AssetManager) Shader(name string) rl.Shader {
	return assets.shaders[name]
}

func (assets *AssetManager) Unload() {
	for _, model := range assets.models {
		rl.UnloadModel(*model)
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

func (assets *AssetManager) loadModels() {
	assets.models["ground"] = assets.loadGroundModel()
	assets.models["drone"] = assets.loadDroneModel()
}

func (assets *AssetManager) loadGroundModel() *rl.Model {
	ground := rl.LoadModelFromMesh(rl.GenMeshPlane(gridSize, gridSize, gridSubdivisions, gridSubdivisions))
	ground.Materials.Shader = assets.shaders["shadow_receiver"]
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

func shaderAssetSources() map[string]shaderFiles {
	const shaderDir = "assets/shaders"

	entries, err := shaderAssets.ReadDir(shaderDir)
	if err != nil {
		panic(fmt.Errorf("read shader dir %q: %w", shaderDir, err))
	}

	sources := make(map[string]shaderFiles)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		stem := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		sourcePath := path.Join(shaderDir, entry.Name())
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
