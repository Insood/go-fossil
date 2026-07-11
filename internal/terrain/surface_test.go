package terrain

import (
	"image"
	"image/color"
	"testing"
)

func TestBuildSurfaceGeometry(t *testing.T) {
	t.Parallel()

	level := LevelData{
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

	surface, err := BuildSurface(level, map[string]image.Image{
		"red.png":  solidImage(2, 2, color.RGBA{R: 255, A: 255}),
		"blue.png": solidImage(2, 2, color.RGBA{B: 255, A: 255}),
	}, 4)
	if err != nil {
		t.Fatalf("BuildSurface() error = %v", err)
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

	level := LevelData{
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

	surface, err := BuildSurface(level, map[string]image.Image{
		"left.png":  solidImage(1, 1, color.RGBA{R: 255, A: 255}),
		"right.png": solidImage(1, 1, color.RGBA{G: 255, A: 255}),
	}, 2)
	if err != nil {
		t.Fatalf("BuildSurface() error = %v", err)
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
	if got := color.RGBAModel.Convert(surface.OverlayImage.At(1, 1)).(color.RGBA); got.A != 0 {
		t.Fatalf("overlay pixel alpha = %d, want transparent", got.A)
	}
}

func TestWorldToUVAndOverlayPixel(t *testing.T) {
	t.Parallel()

	surface := &Surface{
		Width:         8,
		Height:        10,
		PixelsPerTile: 4,
		OverlayImage:  image.NewRGBA(image.Rect(0, 0, 32, 40)),
	}

	u, v := surface.WorldToUV(0, 0)
	assertClose(t, u, 0)
	assertClose(t, v, 0)

	u, v = surface.WorldToUV(4, 5)
	assertClose(t, u, 0.5)
	assertClose(t, v, 0.5)

	u, v = surface.WorldToUV(8, 10)
	assertClose(t, u, 1)
	assertClose(t, v, 1)

	x, y := surface.WorldToOverlayPixel(8, 10)
	if x != 31 || y != 39 {
		t.Fatalf("edge overlay pixel = (%d, %d), want (31, 39)", x, y)
	}
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
