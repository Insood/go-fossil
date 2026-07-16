package main

import (
	"image"
	"image/color"
	"testing"
)

func TestPaintBurnMarkSetsOnlyExactPixel(t *testing.T) {
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
			if x == 2 && y == 3 {
				want = color.RGBA{A: 255}
			}
			if got != want {
				t.Fatalf("pixel (%d,%d) = %#v, want %#v", x, y, got, want)
			}
		}
	}
}

