package main

import (
	"image"
	"math"
)

// Positive orientation rotates clockwise in the image plane.
func rotateImageClockwiseNearest(src image.Image, degrees float64) *image.RGBA {
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()
	if srcWidth == 0 || srcHeight == 0 {
		return image.NewRGBA(image.Rect(0, 0, 0, 0))
	}

	radians := degrees * math.Pi / 180
	sin, cos := math.Sincos(radians)
	if math.Abs(sin) < 1e-9 {
		sin = 0
	}
	if math.Abs(cos) < 1e-9 {
		cos = 0
	}
	outWidth := int(math.Ceil(math.Abs(float64(srcWidth)*cos) + math.Abs(float64(srcHeight)*sin)))
	outHeight := int(math.Ceil(math.Abs(float64(srcWidth)*sin) + math.Abs(float64(srcHeight)*cos)))
	if outWidth <= 0 || outHeight <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 0, 0))
	}

	dst := image.NewRGBA(image.Rect(0, 0, outWidth, outHeight))
	srcCenterX := float64(srcBounds.Min.X) + float64(srcWidth-1)/2
	srcCenterY := float64(srcBounds.Min.Y) + float64(srcHeight-1)/2
	dstCenterX := float64(outWidth-1) / 2
	dstCenterY := float64(outHeight-1) / 2

	for y := 0; y < outHeight; y++ {
		for x := 0; x < outWidth; x++ {
			dx := float64(x) - dstCenterX
			dy := float64(y) - dstCenterY

			srcX := cos*dx - sin*dy + srcCenterX
			srcY := sin*dx + cos*dy + srcCenterY

			sampleX := int(math.Round(srcX))
			sampleY := int(math.Round(srcY))
			if sampleX < srcBounds.Min.X || sampleX >= srcBounds.Max.X || sampleY < srcBounds.Min.Y || sampleY >= srcBounds.Max.Y {
				continue
			}

			dst.Set(x, y, src.At(sampleX, sampleY))
		}
	}

	return dst
}

func minFloat32(a, b float32) float32 {
	if a < b {
		return a
	}

	return b
}
