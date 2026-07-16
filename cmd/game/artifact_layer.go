package main

import (
	"fmt"
	"image"
	"image/draw"
	"math"

	"go-fossil/internal/terrain"
)

type artifactDefinitionSource interface {
	LookupArtifactDefinition(name string) (*ArtifactDefinition, bool)
	LookupImage(assetPath string) (image.Image, bool)
}

func buildArtifactLayer(chunk terrain.ChunkData, assets artifactDefinitionSource, layerBounds image.Rectangle) *image.RGBA {
	layer := image.NewRGBA(layerBounds)
	if len(chunk.Artifacts) == 0 {
		return layer
	}

	for _, artifact := range chunk.Artifacts {
		definition := Must(assets.LookupArtifactDefinition(artifact.Name))
		if definition.Width <= 0 || definition.Height <= 0 {
			panic(fmt.Errorf("chunk %q: artifact definition %q has invalid size %dx%d", chunk.Name, definition.Name, definition.Width, definition.Height))
		}

		sourceImage := Must(assets.LookupImage(definition.ImagePath))
		scaled := resizeImageNearest(sourceImage, definition.Width, definition.Height)
		rotated := rotateImageClockwiseNearest(scaled, float64(artifact.Orientation))
		blitCenteredImage(layer, rotated, artifact.X, artifact.Z)
	}

	return layer
}

func resizeImageNearest(src image.Image, width, height int) *image.RGBA {
	if width <= 0 || height <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 0, 0))
	}

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()
	if srcWidth == 0 || srcHeight == 0 {
		return dst
	}

	for y := 0; y < height; y++ {
		srcY := srcBounds.Min.Y + int(float64(y)*float64(srcHeight)/float64(height))
		if srcY >= srcBounds.Max.Y {
			srcY = srcBounds.Max.Y - 1
		}
		for x := 0; x < width; x++ {
			srcX := srcBounds.Min.X + int(float64(x)*float64(srcWidth)/float64(width))
			if srcX >= srcBounds.Max.X {
				srcX = srcBounds.Max.X - 1
			}

			dst.Set(x, y, src.At(srcX, srcY))
		}
	}

	return dst
}

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

func blitCenteredImage(dst *image.RGBA, src image.Image, centerX, centerY float32) {
	srcBounds := src.Bounds()
	if srcBounds.Empty() {
		return
	}

	left := int(math.Round(float64(centerX))) - srcBounds.Dx()/2
	top := int(math.Round(float64(centerY))) - srcBounds.Dy()/2
	dstRect := image.Rect(left, top, left+srcBounds.Dx(), top+srcBounds.Dy()).Intersect(dst.Bounds())
	if dstRect.Empty() {
		return
	}

	srcPoint := image.Point{
		X: dstRect.Min.X - left + srcBounds.Min.X,
		Y: dstRect.Min.Y - top + srcBounds.Min.Y,
	}

	draw.Draw(dst, dstRect, src, srcPoint, draw.Over)
}
