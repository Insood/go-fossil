package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveFragmentImageWritesPNG(t *testing.T) {
	t.Parallel()

	manager := NewArtifactManager()
	manager.fragmentOutputDir = t.TempDir()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	img.SetRGBA(1, 1, color.RGBA{B: 255, A: 128})

	manager.saveFragmentImage(1, img)

	file, err := os.Open(filepath.Join(manager.fragmentOutputDir, "fragment-000001.png"))
	if err != nil {
		t.Fatalf("open fragment png: %v", err)
	}
	defer file.Close()

	decoded, err := png.Decode(file)
	if err != nil {
		t.Fatalf("decode fragment png: %v", err)
	}

	if got := color.RGBAModel.Convert(decoded.At(0, 0)).(color.RGBA); got != (color.RGBA{R: 255, A: 255}) {
		t.Fatalf("pixel (0,0) = %#v, want red opaque", got)
	}
	if got := color.NRGBAModel.Convert(decoded.At(1, 1)).(color.NRGBA); got.A != 128 || got.B < 250 || got.R != 0 || got.G != 0 {
		t.Fatalf("pixel (1,1) = %#v, want blue semi-transparent", got)
	}
}

func TestCreateFragmentFromRegionUsesExactPixels(t *testing.T) {
	t.Parallel()

	manager := NewArtifactManager()
	manager.fragmentOutputDir = t.TempDir()

	background := image.NewRGBA(image.Rect(0, 0, 2, 2))
	background.SetRGBA(0, 0, color.RGBA{G: 255, A: 255})
	background.SetRGBA(1, 0, color.RGBA{B: 255, A: 255})
	background.SetRGBA(0, 1, color.RGBA{G: 255, A: 255})
	background.SetRGBA(1, 1, color.RGBA{B: 255, A: 255})

	foreground := image.NewRGBA(image.Rect(0, 0, 2, 2))
	foreground.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	foreground.SetRGBA(1, 1, color.RGBA{})

	fragment := manager.CreateFragmentFromRegion(
		background,
		foreground,
		image.Rect(0, 0, 2, 2),
		[]image.Point{{X: 0, Y: 0}, {X: 1, Y: 1}},
	)

	if got := color.NRGBAModel.Convert(fragment.Image.At(0, 0)).(color.NRGBA); got.R < 250 || got.G != 0 || got.B != 0 || got.A != 255 {
		t.Fatalf("pixel (0,0) = %#v, want opaque red from foreground", got)
	}
	if got := color.NRGBAModel.Convert(fragment.Image.At(1, 1)).(color.NRGBA); got.B < 250 || got.A != 255 {
		t.Fatalf("pixel (1,1) = %#v, want opaque blue from background", got)
	}
	if got := color.NRGBAModel.Convert(fragment.Image.At(1, 0)).(color.NRGBA); got.A != 0 {
		t.Fatalf("pixel (1,0) = %#v, want transparent", got)
	}
	if got := color.NRGBAModel.Convert(fragment.Image.At(0, 1)).(color.NRGBA); got.A != 0 {
		t.Fatalf("pixel (0,1) = %#v, want transparent", got)
	}
}
