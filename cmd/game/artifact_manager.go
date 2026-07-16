package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ArtifactManager struct {
	nextID            int32
	nextFragmentID    int32
	artifacts         map[int32]*Artifact
	fragments         map[int32]*ArtifactFragment
	fragmentOutputDir string
}

func NewArtifactManager() *ArtifactManager {
	return &ArtifactManager{
		artifacts:         make(map[int32]*Artifact),
		fragments:         make(map[int32]*ArtifactFragment),
		fragmentOutputDir: artifactFragmentOutputDir,
	}
}

func (manager *ArtifactManager) Lookup(id int32) (*Artifact, bool) {
	artifact, ok := manager.artifacts[id]
	return artifact, ok
}

func (manager *ArtifactManager) LookupFragment(id int32) (*ArtifactFragment, bool) {
	fragment, ok := manager.fragments[id]
	return fragment, ok
}

func (manager *ArtifactManager) RegisterChunkArtifact(
	chunk *TerrainChunk,
	name string,
	value int,
	centerX float32,
	centerZ float32,
	pixelBounds image.Rectangle,
) *Artifact {
	manager.nextID++
	artifactID := manager.nextID

	artifact := &Artifact{
		ID:          artifactID,
		Name:        name,
		Value:       value,
		CenterX:     centerX,
		CenterZ:     centerZ,
		PixelBounds: pixelBounds,
		Chunk:       chunk,
	}

	manager.artifacts[artifactID] = artifact
	return artifact
}

func (manager *ArtifactManager) CreateFragment(src image.Image) *ArtifactFragment {
	return manager.CreateFragmentFromLayers(nil, src, src.Bounds())
}

func (manager *ArtifactManager) CreateFragmentFromRect(src image.Image, bounds image.Rectangle) *ArtifactFragment {
	return manager.CreateFragmentFromLayers(nil, src, bounds)
}

func (manager *ArtifactManager) CreateFragmentFromLayers(background image.Image, foreground image.Image, bounds image.Rectangle) *ArtifactFragment {
	return manager.CreateFragmentFromRegion(background, foreground, bounds, nil)
}

func (manager *ArtifactManager) CreateFragmentFromRegion(background image.Image, foreground image.Image, bounds image.Rectangle, points []image.Point) *ArtifactFragment {
	manager.nextFragmentID++
	fragmentID := manager.nextFragmentID

	fragmentImage := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	if len(points) == 0 {
		if background != nil {
			draw.Draw(fragmentImage, fragmentImage.Bounds(), background, bounds.Min, draw.Src)
		}
		if foreground != nil {
			draw.Draw(fragmentImage, fragmentImage.Bounds(), foreground, bounds.Min, draw.Over)
		}
	} else {
		for _, point := range points {
			localX := point.X - bounds.Min.X
			localY := point.Y - bounds.Min.Y
			if localX < 0 || localY < 0 || localX >= bounds.Dx() || localY >= bounds.Dy() {
				continue
			}

			pixel := color.RGBA{}
			if background != nil {
				pixel = color.RGBAModel.Convert(background.At(point.X, point.Y)).(color.RGBA)
			}
			if foreground != nil {
				overlay := color.RGBAModel.Convert(foreground.At(point.X, point.Y)).(color.RGBA)
				if overlay.A != 0 {
					pixel = overlay
				}
			}
			fragmentImage.SetRGBA(localX, localY, pixel)
		}
	}
	fragmentTexture := loadTextureFromImage(fragmentImage)
	fragment := &ArtifactFragment{
		ID:      fragmentID,
		Image:   fragmentImage,
		Texture: fragmentTexture,
	}

	manager.fragments[fragmentID] = fragment
	manager.saveFragmentImage(fragmentID, fragmentImage)
	return fragment
}

func (manager *ArtifactManager) Unload() {
	for _, fragment := range manager.fragments {
		rl.UnloadTexture(fragment.Texture)
		fragment.Texture = rl.Texture2D{}
	}
}

func (manager *ArtifactManager) saveFragmentImage(id int32, img image.Image) {
	if err := os.MkdirAll(manager.fragmentOutputDir, 0o755); err != nil {
		panic(fmt.Errorf("create artifact fragment output dir %q: %w", manager.fragmentOutputDir, err))
	}

	path := filepath.Join(manager.fragmentOutputDir, fmt.Sprintf("fragment-%06d.png", id))
	file, err := os.Create(path)
	if err != nil {
		panic(fmt.Errorf("create artifact fragment file %q: %w", path, err))
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		panic(fmt.Errorf("encode artifact fragment png %q: %w", path, err))
	}
}
