package main

import (
	"image"
	"image/color"

	"github.com/disintegration/gift"
	rl "github.com/gen2brain/raylib-go/raylib"
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

func clampFloat32(value, min, max float32) float32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}

	return value
}

func xzVector(position rl.Vector3) rl.Vector2 {
	return rl.NewVector2(position.X, position.Z)
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
