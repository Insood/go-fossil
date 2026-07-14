package terrain

import (
	"image"
	"image/color"
	"testing"
)

func TestBuildSurfaceGeometry(t *testing.T) {
	t.Parallel()

	chunk := ChunkData{
		Name:   "Geometry Test",
		Width:  2,
		Height: 1,
		Tiles: [][]int{
			{0, 1},
		},
		TileDefinitions: []string{"red.png", "blue.png"},
		HeightSamples: [][]float32{
			{0, 0.5, 1.0},
			{0.25, 0.75, 1.25},
		},
	}

	surface, err := BuildSurfaceMesh(chunk)
	if err != nil {
		t.Fatalf("BuildSurfaceMesh() error = %v", err)
	}

	if got, want := len(surface.Vertices)/3, 6; got != want {
		t.Fatalf("vertex count = %d, want %d", got, want)
	}
	if got, want := len(surface.Indices)/3, 4; got != want {
		t.Fatalf("triangle count = %d, want %d", got, want)
	}

	if got := surface.Vertices[1]; got != 0 {
		t.Fatalf("vertex[0].y = %v, want 0", got)
	}
	if got := surface.Vertices[4]; got != 0.5 {
		t.Fatalf("vertex[1].y = %v, want 0.5", got)
	}
	if got := surface.Vertices[16]; got != 1.25 {
		t.Fatalf("vertex[last].y = %v, want 1.25", got)
	}

	assertClose(t, surface.Texcoords[0], 0)
	assertClose(t, surface.Texcoords[1], 0)
	assertClose(t, surface.Texcoords[4], 1)
	assertClose(t, surface.Texcoords[5], 0)
	assertClose(t, surface.Texcoords[10], 1)
	assertClose(t, surface.Texcoords[11], 1)
}

func TestBuildSurfaceTextureComposition(t *testing.T) {
	t.Parallel()

	chunk := ChunkData{
		Name:   "Texture Test",
		Width:  2,
		Height: 1,
		Tiles: [][]int{
			{0, 1},
		},
		TileDefinitions: []string{"left.png", "right.png"},
		HeightSamples: [][]float32{
			{0, 0, 0},
			{0, 0, 0},
		},
	}

	surface, err := BuildSurfaceTexture(chunk, map[string]image.Image{
		"left.png":  solidImage(1, 1, color.RGBA{R: 255, A: 255}),
		"right.png": solidImage(1, 1, color.RGBA{G: 255, A: 255}),
	}, 2)
	if err != nil {
		t.Fatalf("BuildSurfaceTexture() error = %v", err)
	}

	if got, want := surface.BaseImage.Bounds().Dx(), 4; got != want {
		t.Fatalf("base image width = %d, want %d", got, want)
	}
	if got, want := surface.BaseImage.Bounds().Dy(), 2; got != want {
		t.Fatalf("base image height = %d, want %d", got, want)
	}

	if got := color.RGBAModel.Convert(surface.BaseImage.At(0, 0)).(color.RGBA); got.R != 255 || got.G != 0 {
		t.Fatalf("left tile pixel = %#v, want red tile", got)
	}
	if got := color.RGBAModel.Convert(surface.BaseImage.At(3, 0)).(color.RGBA); got.G != 255 || got.R != 0 {
		t.Fatalf("right tile pixel = %#v, want green tile", got)
	}
}

func TestSampleHeight(t *testing.T) {
	t.Parallel()

	surface := &SurfaceMesh{
		Width:  2,
		Height: 2,
		HeightSamples: [][]float32{
			{0, 1, 2},
			{1, 2, 3},
			{2, 3, 4},
		},
	}

	assertClose(t, surface.SampleHeight(0, 0), 0)
	assertClose(t, surface.SampleHeight(2, 2), 4)
	assertClose(t, surface.SampleHeight(0.5, 0.5), 1)
	assertClose(t, surface.SampleHeight(1.25, 0.75), 2)
	assertClose(t, surface.SampleHeight(-1, 3), 2)
}

func solidImage(width, height int, fill color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, fill)
		}
	}
	return img
}

func assertClose(t *testing.T, got, want float32) {
	t.Helper()
	if !roughlyEqual(got, want) {
		t.Fatalf("value = %v, want %v", got, want)
	}
}
