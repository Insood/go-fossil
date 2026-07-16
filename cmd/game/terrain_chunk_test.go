package main

import (
	"image"
	"image/color"
	"testing"
)

func TestPaintBurnMarkSets2x2Mask(t *testing.T) {
	t.Parallel()

	chunk := &TerrainChunk{
		BurnOverlayImage: image.NewRGBA(image.Rect(0, 0, 5, 5)),
	}

	if !chunk.paintBurnMark(2, 3) {
		t.Fatal("paintBurnMark returned false")
	}

	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			got := chunk.BurnOverlayImage.RGBAAt(x, y)
			want := color.RGBA{}
			if x >= 2 && x <= 3 && y >= 3 && y <= 4 {
				want = color.RGBA{A: burnOverlayAlpha}
			}
			if got != want {
				t.Fatalf("pixel (%d,%d) = %#v, want %#v", x, y, got, want)
			}
		}
	}
}

func TestPaintBurnMarkMarksArtifactDataAsBurnedInSameMask(t *testing.T) {
	t.Parallel()

	artifactData := NewArtifactData(image.Rect(0, 0, 5, 5))
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			artifactData.SetID(x, y, int32(7))
		}
	}

	chunk := &TerrainChunk{
		BurnOverlayImage: image.NewRGBA(image.Rect(0, 0, 5, 5)),
		ArtifactData:     artifactData,
	}

	if !chunk.paintBurnMark(2, 3) {
		t.Fatal("paintBurnMark returned false")
	}

	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			got := artifactData.IDAt(x, y)
			want := int32(7)
			if x >= 2 && x <= 3 && y >= 3 && y <= 4 {
				want = int32(-1)
			}
			if got != want {
				t.Fatalf("artifact id at (%d,%d) = %d, want %d", x, y, got, want)
			}
		}
	}
}
