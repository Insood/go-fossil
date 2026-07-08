package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

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
	for name, paths := range shaderAssetPaths() {
		assets.shaders[name] = rl.LoadShader(paths.vertex, paths.fragment)
	}
}

func (assets *AssetManager) loadModels() {
	assets.models["ground"] = assets.loadGroundModel()
	assets.models["drone"] = assets.loadDroneModel()
}

func (assets *AssetManager) loadGroundModel() *rl.Model {
	gridShader := assets.gridShader()
	assets.configureGridShader(gridShader)

	ground := rl.LoadModelFromMesh(rl.GenMeshPlane(gridSize, gridSize, gridSubdivisions, gridSubdivisions))
	ground.Materials.Shader = gridShader
	return &ground
}

func (assets *AssetManager) loadDroneModel() *rl.Model {
	drone := rl.LoadModelFromMesh(rl.GenMeshCube(droneWidth, droneHeight, droneDepth))
	return &drone
}

func (assets *AssetManager) gridShader() rl.Shader {
	return assets.shaders["grid"]
}

func (assets *AssetManager) configureGridShader(gridShader rl.Shader) {
	gridCellsLoc := rl.GetShaderLocation(gridShader, "gridCells")
	lineWidthLoc := rl.GetShaderLocation(gridShader, "lineWidth")
	rl.SetShaderValue(gridShader, gridCellsLoc, []float32{gridSubdivisions}, rl.ShaderUniformFloat)
	rl.SetShaderValue(gridShader, lineWidthLoc, []float32{gridLineWidth}, rl.ShaderUniformFloat)
}

type shaderFiles struct {
	vertex   string
	fragment string
}

func shaderAssetPaths() map[string]shaderFiles {
	shaderDir := gameAssetPath("assets", "shaders")
	entries, err := os.ReadDir(shaderDir)
	if err != nil {
		panic(fmt.Errorf("read shader dir %q: %w", shaderDir, err))
	}

	paths := make(map[string]shaderFiles)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		stem := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))

		switch ext {
		case ".vs", ".vert":
			current := paths[stem]
			current.vertex = filepath.Join(shaderDir, entry.Name())
			paths[stem] = current
		case ".fs", ".frag":
			current := paths[stem]
			current.fragment = filepath.Join(shaderDir, entry.Name())
			paths[stem] = current
		}
	}

	for name, paths := range paths {
		if paths.vertex == "" || paths.fragment == "" {
			panic(fmt.Errorf("shader %q is missing a vertex or fragment file", name))
		}
	}

	return paths
}

func gameAssetPath(parts ...string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join(parts...)
	}

	base := filepath.Dir(filename)
	segments := make([]string, 0, len(parts)+1)
	segments = append(segments, base)
	segments = append(segments, parts...)
	return filepath.Join(segments...)
}
