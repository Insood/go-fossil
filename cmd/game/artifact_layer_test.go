package main

import (
	"image"
	"image/color"
	"testing"

	"go-fossil/internal/terrain"
)

func TestBuildArtifactLayerPlacesAndRotatesArtifact(t *testing.T) {
	t.Parallel()

	assets := fakeArtifactSource{
		definitions: map[string]*ArtifactDefinition{
			"phone": {
				Name:      "phone",
				ImagePath: "textures/phone.png",
				Width:     2,
				Height:    1,
				Value:     10,
			},
		},
		images: map[string]image.Image{
			"textures/phone.png": solidArtifactImage([][]color.RGBA{
				{{R: 255, A: 255}, {G: 255, A: 255}},
			}),
		},
	}

	layer := buildArtifactLayer(
		terrain.ChunkData{
			Name: "Artifact Test",
			Artifacts: []terrain.ArtifactPlacement{
				{Name: "phone", X: 2, Z: 2, Orientation: 90},
			},
		},
		assets,
		image.Rect(0, 0, 4, 4),
	)

	if got := color.RGBAModel.Convert(layer.At(2, 1)).(color.RGBA); got.G != 255 || got.R != 0 {
		t.Fatalf("rotated top pixel = %#v, want green", got)
	}
	if got := color.RGBAModel.Convert(layer.At(2, 2)).(color.RGBA); got.R != 255 || got.G != 0 {
		t.Fatalf("rotated bottom pixel = %#v, want red", got)
	}
}

func TestBuildArtifactLayerRejectsUnknownArtifact(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("buildArtifactLayer() did not panic, want missing artifact failure")
		}
	}()

	_ = buildArtifactLayer(
		terrain.ChunkData{
			Name: "Artifact Test",
			Artifacts: []terrain.ArtifactPlacement{
				{Name: "missing", X: 2, Z: 2, Orientation: 0},
			},
		},
		fakeArtifactSource{},
		image.Rect(0, 0, 4, 4),
	)
}

type fakeArtifactSource struct {
	definitions map[string]*ArtifactDefinition
	images      map[string]image.Image
}

func (source fakeArtifactSource) LookupArtifactDefinition(name string) (*ArtifactDefinition, bool) {
	definition, ok := source.definitions[name]
	return definition, ok
}

func (source fakeArtifactSource) LookupImage(assetPath string) (image.Image, bool) {
	img, ok := source.images[assetPath]
	return img, ok
}

func solidArtifactImage(rows [][]color.RGBA) image.Image {
	height := len(rows)
	width := 0
	if height > 0 {
		width = len(rows[0])
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y, row := range rows {
		for x, fill := range row {
			img.SetRGBA(x, y, fill)
		}
	}

	return img
}
