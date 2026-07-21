package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io/fs"
	"math"
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
	sounds              map[string]rl.Music
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
		sounds:              make(map[string]rl.Music),
		textures:            make(map[string]rl.Texture2D),
		assetRoot:           assetRoot,
		assetFS:             os.DirFS(assetRoot),
	}
}

func (assets *AssetManager) Load() {
	assets.loadImages()
	assets.loadArtifactDefinitions()
	assets.loadTextures()
	assets.loadShaders()
	assets.loadModels()
	assets.loadSounds()
}

func (assets *AssetManager) LookupArtifactDefinition(name string) (*ArtifactDefinition, bool) {
	definition, ok := assets.artifactDefinitions[name]
	return definition, ok
}

func (assets *AssetManager) ArtifactDefinitions() []*ArtifactDefinition {
	if len(assets.artifactDefinitions) == 0 {
		return nil
	}

	definitions := make([]*ArtifactDefinition, 0, len(assets.artifactDefinitions))
	for _, definition := range assets.artifactDefinitions {
		definitions = append(definitions, definition)
	}

	return definitions
}

func (assets *AssetManager) LookupImage(assetPath string) (image.Image, bool) {
	normalizedPath := path.Clean(assetPath)
	img, ok := assets.images[normalizedPath]
	return img, ok
}

func (assets *AssetManager) LookupModel(name string) (*rl.Model, bool) {
	model, ok := assets.models[name]
	return model, ok
}

func (assets *AssetManager) LookupShader(name string) (rl.Shader, bool) {
	shader, ok := assets.shaders[name]
	return shader, ok
}

func (assets *AssetManager) LookupSound(name string) (rl.Music, bool) {
	sound, ok := assets.sounds[name]
	return sound, ok
}

func (assets *AssetManager) SoundNames() []string {
	if len(assets.sounds) == 0 {
		return nil
	}

	names := make([]string, 0, len(assets.sounds))
	for name := range assets.sounds {
		names = append(names, name)
	}

	slices.Sort(names)
	return names
}

func (assets *AssetManager) LookupTexture(name string) (rl.Texture2D, bool) {
	texture, ok := assets.textures[name]
	return texture, ok
}

func (assets *AssetManager) Unload() {
	sharedTextureIDs := assets.textureIDs()
	modelTextures := assets.collectModelTextures(sharedTextureIDs)

	// raylib's model teardown does not release textures referenced by material
	// maps, so we collect the model-owned textures first. Shared asset textures
	// stay owned by assets.textures and are skipped here to avoid double frees.
	for _, texture := range modelTextures {
		rl.UnloadTexture(texture)
	}

	for _, model := range assets.models {
		rl.UnloadModel(*model)
	}
	for _, texture := range assets.textures {
		rl.UnloadTexture(texture)
	}
	for _, shader := range assets.shaders {
		rl.UnloadShader(shader)
	}
	for _, sound := range assets.sounds {
		rl.UnloadMusicStream(sound)
	}
}

func (assets *AssetManager) textureIDs() map[uint32]struct{} {
	textureIDs := make(map[uint32]struct{}, len(assets.textures))
	for _, texture := range assets.textures {
		if texture.ID == 0 {
			continue
		}

		textureIDs[texture.ID] = struct{}{}
	}

	return textureIDs
}

func (assets *AssetManager) collectModelTextures(sharedTextureIDs map[uint32]struct{}) map[uint32]rl.Texture2D {
	ownedTextures := make(map[uint32]rl.Texture2D)

	for _, model := range assets.models {
		if model == nil || model.MaterialCount == 0 || model.Materials == nil {
			continue
		}

		for _, material := range model.GetMaterials() {
			if material.Maps == nil {
				continue
			}

			for mapType := int32(0); mapType < rl.MaxMaterialMaps; mapType++ {
				texture := material.GetMap(mapType).Texture
				if !shouldUnloadTexture(texture, sharedTextureIDs) {
					continue
				}

				ownedTextures[texture.ID] = texture
			}
		}
	}

	return ownedTextures
}

func shouldUnloadTexture(texture rl.Texture2D, sharedTextureIDs map[uint32]struct{}) bool {
	if texture.ID == 0 || texture.ID == rl.GetTextureIdDefault() {
		return false
	}

	_, shared := sharedTextureIDs[texture.ID]
	return !shared
}

func (assets *AssetManager) loadShaders() {
	sources := assets.shaderAssetSources()
	for name, source := range sources {
		assets.shaders[name] = rl.LoadShaderFromMemory(source.vertex, source.fragment)
	}
}

func (assets *AssetManager) loadImages() {
	const textureDir = "textures"

	fileNames := assetFileNames(assets.assetFS, textureDir, ".png", ".jpg", ".jpeg")
	for _, fileName := range fileNames {
		assetPath := path.Join(textureDir, fileName)
		assets.images[assetPath] = loadTextureImageAsset(assets.assetFS, assetPath)
	}
}

func (assets *AssetManager) loadArtifactDefinitions() {
	const artifactDir = "artifacts"

	fileNames := assetFileNames(assets.assetFS, artifactDir, ".json")
	for _, fileName := range fileNames {
		definitionPath := path.Join(artifactDir, fileName)
		definition, err := loadArtifactDefinitionAsset(assets.assetFS, definitionPath)
		if err != nil {
			panic(fmt.Errorf("load artifact definition asset %q: %w", definitionPath, err))
		}
		if _, exists := assets.artifactDefinitions[definition.Name]; exists {
			panic(fmt.Errorf("artifact definition %q declared more than once", definition.Name))
		}

		_, ok := assets.LookupImage(definition.ImagePath)
		if !ok {
			panic(fmt.Errorf("artifact definition %q references missing image %q", definition.Name, definition.ImagePath))
		}
		definitionCopy := definition
		assets.artifactDefinitions[definitionCopy.Name] = &definitionCopy
	}
}

func countNonTransparentPixels(img image.Image) int {
	bounds := img.Bounds()
	count := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if color.RGBAModel.Convert(img.At(x, y)).(color.RGBA).A == 0 {
				continue
			}
			count++
		}
	}

	return count
}

func (assets *AssetManager) loadTextures() {
	const textureDir = "textures"

	assets.textures["white"] = loadSolidTexture(rl.White)
	assets.textures["blank"] = loadSolidTexture(rl.Blank)

	fileNames := assetFileNames(assets.assetFS, textureDir, ".png", ".jpg", ".jpeg")
	for _, fileName := range fileNames {
		assetPath := path.Join(textureDir, fileName)
		textureName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		img, ok := assets.LookupImage(assetPath)
		if !ok {
			panic(fmt.Errorf("texture %q references missing image %q", textureName, assetPath))
		}

		texture := loadTextureFromImageAsset(img)
		rl.SetTextureWrap(texture, rl.WrapRepeat)
		if _, exists := assets.textures[textureName]; exists {
			panic(fmt.Errorf("texture %q declared more than once", textureName))
		}
		assets.textures[textureName] = texture
	}
}

func (assets *AssetManager) loadModels() {
	assets.models["drone"] = assets.loadDroneModel()
	assets.models["particle_cube"] = assets.loadUnitParticleCubeModel()
	assets.models["prop_cube"] = assets.loadUnitCubeModel()
	assets.models["prop_sphere"] = assets.loadUnitSphereModel()
	assets.models[tutorialArtifactMarkerModelName] = assets.loadTutorialConeModel()
}

func (assets *AssetManager) loadSounds() {
	const soundDir = "sounds"

	fileNames := assetFileNames(assets.assetFS, soundDir, ".mp3", ".ogg", ".wav")
	for _, fileName := range fileNames {
		soundName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		if _, exists := assets.sounds[soundName]; exists {
			panic(fmt.Errorf("sound %q declared more than once", soundName))
		}

		sound := rl.LoadMusicStream(assets.assetPath(soundDir, fileName))
		if !rl.IsMusicValid(sound) {
			panic(fmt.Errorf("load sound stream asset %q", path.Join(soundDir, fileName)))
		}
		assets.sounds[soundName] = sound
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

func (assets *AssetManager) loadUnitParticleCubeModel() *rl.Model {
	cube := rl.LoadModelFromMesh(rl.GenMeshCube(1.0, 1.0, 1.0))
	return &cube
}

func (assets *AssetManager) loadUnitSphereModel() *rl.Model {
	sphere := rl.LoadModelFromMesh(rl.GenMeshSphere(0.5, 24, 24))
	configureShadowReceiverMaterial(&sphere, assets)
	return &sphere
}

func (assets *AssetManager) loadTutorialConeModel() *rl.Model {
	cone := rl.LoadModelFromMesh(rl.GenMeshCone(
		tutorialArtifactMarkerRadius,
		tutorialArtifactMarkerHeight,
		tutorialArtifactMarkerSlices,
	))
	cone.Transform = rl.MatrixRotateX(float32(math.Pi))
	materials := cone.GetMaterials()
	for i := range materials {
		materials[i].Shader = Must(assets.LookupShader("tutorial_marker"))
	}
	return &cone
}

func configureShadowReceiverMaterial(model *rl.Model, assets *AssetManager) {
	materials := model.GetMaterials()
	for i := range materials {
		materials[i].Shader = Must(assets.LookupShader("shadow_receiver"))
		materials[i].GetMap(rl.MapAlbedo).Color = rl.White
		if materials[i].GetMap(rl.MapAlbedo).Texture.ID == 0 {
			rl.SetMaterialTexture(&materials[i], rl.MapAlbedo, Must(assets.LookupTexture("white")))
		}
		materials[i].GetMap(rl.MapEmission).Color = rl.White
		if materials[i].GetMap(rl.MapEmission).Texture.ID == 0 {
			rl.SetMaterialTexture(&materials[i], rl.MapEmission, Must(assets.LookupTexture("blank")))
		}
		materials[i].GetMap(rl.MapOcclusion).Color = rl.White
		if materials[i].GetMap(rl.MapOcclusion).Texture.ID == 0 {
			rl.SetMaterialTexture(&materials[i], rl.MapOcclusion, Must(assets.LookupTexture("blank")))
		}
		materials[i].Shader.UpdateLocation(
			rl.ShaderLocMapHeight,
			rl.GetShaderLocation(materials[i].Shader, "shadowMap"),
		)
		materials[i].Shader.UpdateLocation(
			rl.ShaderLocMapEmission,
			rl.GetShaderLocation(materials[i].Shader, "texture1"),
		)
		materials[i].Shader.UpdateLocation(
			rl.ShaderLocMapOcclusion,
			rl.GetShaderLocation(materials[i].Shader, "texture2"),
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
	fileNames := assetFileNames(assets.assetFS, shaderDir, ".vs", ".vert", ".fs", ".frag")

	for _, fileName := range fileNames {
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
			if current.vertex != "" {
				panic(fmt.Errorf("shader %q has more than one vertex source", stem))
			}
			current.vertex = source
			sources[stem] = current
		case ".fs", ".frag":
			current := sources[stem]
			if current.fragment != "" {
				panic(fmt.Errorf("shader %q has more than one fragment source", stem))
			}
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

func loadTextureImageAsset(assetFS fs.FS, assetPath string) image.Image {
	data, err := fs.ReadFile(assetFS, assetPath)
	if err != nil {
		panic(fmt.Errorf("read texture asset %q: %w", assetPath, err))
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		panic(fmt.Errorf("decode texture asset %q: %w", assetPath, err))
	}

	return img
}

func loadSolidTexture(fill rl.Color) rl.Texture2D {
	solid := image.NewRGBA(image.Rect(0, 0, 1, 1))
	solid.SetRGBA(0, 0, color.RGBAModel.Convert(fill).(color.RGBA))
	return loadTextureFromImageAsset(solid)
}

func loadTextureFromImageAsset(src image.Image) rl.Texture2D {
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
