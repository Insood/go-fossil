package main

import (
	"image"
	"image/color"

	"github.com/disintegration/gift"
)

// Positive orientation rotates clockwise in the image plane.
func rotateImageClockwiseNearest(src image.Image, degrees float64) *image.RGBA {
	filter := gift.New(gift.Rotate(float32(degrees), color.Transparent, gift.NearestNeighborInterpolation))
	dst := image.NewRGBA(filter.Bounds(src.Bounds()))
	filter.Draw(dst, src)
	return dst
}

func minFloat32(a, b float32) float32 {
	if a < b {
		return a
	}

	return b
}
