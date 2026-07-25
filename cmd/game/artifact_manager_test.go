package main

import (
	"image"
	"image/color"
	"testing"
)

func TestCreateFragmentFromRegionSkipsSmallFragments(t *testing.T) {
	t.Parallel()

	manager := NewArtifactManager()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	fragment := manager.CreateFragmentFromRegion(
		img,
		nil,
		img.Bounds(),
		[]image.Point{{X: 0, Y: 0}, {X: 1, Y: 1}},
	)

	if fragment != nil {
		t.Fatalf("fragment = %#v, want nil for region below minimum pixel size", fragment)
	}
	if _, ok := manager.LookupFragment(1); ok {
		t.Fatal("undersized fragment was recorded")
	}
	if manager.nextFragmentID != 0 {
		t.Fatalf("next fragment ID = %d, want 0", manager.nextFragmentID)
	}
}

func TestArtifactManagerArtifactsReturnsArtifactsSortedByID(t *testing.T) {
	t.Parallel()

	manager := NewArtifactManager()
	manager.artifacts[3] = &Artifact{ID: 3}
	manager.artifacts[1] = &Artifact{ID: 1}
	manager.artifacts[2] = &Artifact{ID: 2}

	artifacts := manager.Artifacts()
	want := []int32{1, 2, 3}
	if len(artifacts) != len(want) {
		t.Fatalf("artifact count = %d, want %d", len(artifacts), len(want))
	}
	for i, artifact := range artifacts {
		if artifact.ID != want[i] {
			t.Fatalf("artifact %d ID = %d, want %d", i, artifact.ID, want[i])
		}
	}
}

func TestArtifactManagerFragmentCount(t *testing.T) {
	t.Parallel()

	manager := NewArtifactManager()
	if got, want := manager.FragmentCount(), 0; got != want {
		t.Fatalf("fragment count = %d, want %d", got, want)
	}

	manager.fragments[1] = &ArtifactFragment{ID: 1}
	if got, want := manager.FragmentCount(), 1; got != want {
		t.Fatalf("fragment count = %d, want %d", got, want)
	}
	if got, want := manager.CollectedFragmentCount(), 0; got != want {
		t.Fatalf("collected fragment count = %d, want %d", got, want)
	}

	manager.fragments[1].Collected = true
	if got, want := manager.CollectedFragmentCount(), 1; got != want {
		t.Fatalf("collected fragment count = %d, want %d", got, want)
	}
}

func TestCreateFragmentFromRegionUsesExactPixels(t *testing.T) {
	t.Parallel()

	manager := NewArtifactManager()

	background := image.NewRGBA(image.Rect(0, 0, 6, 4))
	background.SetRGBA(0, 0, color.RGBA{G: 255, A: 255})
	background.SetRGBA(1, 0, color.RGBA{B: 255, A: 255})
	background.SetRGBA(0, 1, color.RGBA{G: 255, A: 255})
	background.SetRGBA(1, 1, color.RGBA{B: 255, A: 255})

	foreground := image.NewRGBA(image.Rect(0, 0, 6, 4))
	foreground.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	foreground.SetRGBA(1, 1, color.RGBA{})

	points := make([]image.Point, 0, 20)
	for y := 0; y < 4; y++ {
		for x := 0; x < 6; x++ {
			if (x == 1 && y == 0) || (x == 0 && y == 1) {
				continue
			}
			if len(points) == 20 {
				break
			}
			points = append(points, image.Point{X: x, Y: y})
		}
	}

	fragment := manager.CreateFragmentFromRegion(
		background,
		foreground,
		image.Rect(0, 0, 6, 4),
		points,
	)

	if fragment == nil {
		t.Fatal("fragment = nil, want recorded fragment at minimum pixel size")
	}
	if got, want := fragment.Weight, 20; got != want {
		t.Fatalf("fragment weight = %d, want %d", got, want)
	}
	if got, want := fragment.Score, 0; got != want {
		t.Fatalf("fragment score = %d, want %d", got, want)
	}

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
