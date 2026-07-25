package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ArtifactManager struct {
	nextID         int32
	nextFragmentID int32
	artifacts      map[int32]*Artifact
	fragments      map[int32]*ArtifactFragment
}

func NewArtifactManager() *ArtifactManager {
	return &ArtifactManager{
		artifacts: make(map[int32]*Artifact),
		fragments: make(map[int32]*ArtifactFragment),
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

func (manager *ArtifactManager) FragmentCount() int {
	if manager == nil {
		return 0
	}

	return len(manager.fragments)
}

func (manager *ArtifactManager) CollectedFragmentCount() int {
	if manager == nil {
		return 0
	}

	count := 0
	for _, fragment := range manager.fragments {
		if fragment.Collected {
			count++
		}
	}

	return count
}

func (manager *ArtifactManager) CarriedFragmentWeight() int {
	if manager == nil {
		return 0
	}

	weight := 0
	for _, fragment := range manager.fragments {
		if fragment.Collected && !fragment.DroppedOff {
			weight += fragment.Weight
		}
	}

	return weight
}

func (manager *ArtifactManager) Artifacts() []*Artifact {
	if manager == nil || len(manager.artifacts) == 0 {
		return nil
	}

	artifacts := make([]*Artifact, 0, len(manager.artifacts))
	for _, artifact := range manager.artifacts {
		artifacts = append(artifacts, artifact)
	}

	slices.SortFunc(artifacts, func(left, right *Artifact) int {
		switch {
		case left.ID < right.ID:
			return -1
		case left.ID > right.ID:
			return 1
		default:
			return 0
		}
	})

	return artifacts
}

func (manager *ArtifactManager) RegisterChunkArtifact(
	chunk *TerrainChunk,
	name string,
	value int,
	size int,
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
		Size:        size,
		CenterX:     centerX,
		CenterZ:     centerZ,
		PixelBounds: pixelBounds,
		Chunk:       chunk,
	}

	manager.artifacts[artifactID] = artifact
	return artifact
}

func (manager *ArtifactManager) CreateFragment(src image.Image) *ArtifactFragment {
	return manager.createFragmentFromRegion(nil, src, src.Bounds(), nil, 0, 0)
}

func (manager *ArtifactManager) CreateFragmentFromRect(src image.Image, bounds image.Rectangle) *ArtifactFragment {
	return manager.createFragmentFromRegion(nil, src, bounds, nil, 0, 0)
}

func (manager *ArtifactManager) CreateFragmentFromLayers(background image.Image, foreground image.Image, bounds image.Rectangle) *ArtifactFragment {
	return manager.createFragmentFromRegion(background, foreground, bounds, nil, 0, 0)
}

func (manager *ArtifactManager) CreateFragmentFromRegion(background image.Image, foreground image.Image, bounds image.Rectangle, points []image.Point) *ArtifactFragment {
	return manager.createFragmentFromRegion(background, foreground, bounds, points, 0, 0)
}

func (manager *ArtifactManager) CreateFragmentFromRegionWithScore(background image.Image, foreground image.Image, bounds image.Rectangle, points []image.Point, score float64, grade float64) *ArtifactFragment {
	return manager.createFragmentFromRegion(background, foreground, bounds, points, score, grade)
}

func (manager *ArtifactManager) createFragmentFromRegion(background image.Image, foreground image.Image, bounds image.Rectangle, points []image.Point, score float64, grade float64) *ArtifactFragment {
	weight := artifactFragmentPixelSize(bounds, points)
	if weight < artifactFragmentMinPixels {
		return nil
	}

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
		Weight:  weight,
		Score:   int(math.Round(score)),
		Grade:   grade,
		Image:   fragmentImage,
		Texture: fragmentTexture,
	}

	manager.fragments[fragmentID] = fragment
	fmt.Printf("Created artifact fragment %d: weight=%d score=%d raw=%.3f grade=%.3f\n", fragment.ID, fragment.Weight, fragment.Score, score, fragment.Grade)
	return fragment
}

func artifactFragmentPixelSize(bounds image.Rectangle, points []image.Point) int {
	if len(points) > 0 {
		return len(points)
	}

	return bounds.Dx() * bounds.Dy()
}

func (manager *ArtifactManager) Unload() {
	for _, fragment := range manager.fragments {
		rl.UnloadTexture(fragment.Texture)
		fragment.Texture = rl.Texture2D{}
	}
}
