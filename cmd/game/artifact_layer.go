package main

import (
	"image"
	"image/color"
	"math"

	"go-fossil/internal/terrain"
)

type artifactDefinitionSource interface {
	LookupArtifactDefinition(name string) (*ArtifactDefinition, bool)
	LookupImage(assetPath string) (image.Image, bool)
}

type artifactBoundsProvider interface {
	Bounds() image.Rectangle
}

func buildArtifactImageLayer(chunk terrain.ChunkData, assets artifactDefinitionSource, layerBounds image.Rectangle) *image.RGBA {
	layer := image.NewRGBA(layerBounds)
	for _, placement := range chunk.Artifacts {
		_, rotated := prepareArtifactPlacement(placement, assets)
		blitArtifactImage(layer, rotated, placement.X, placement.Z)
	}

	return layer
}

func buildArtifactDataLayer(manager *ArtifactManager, chunk *TerrainChunk, assets artifactDefinitionSource, layerBounds image.Rectangle) *ArtifactData {
	artifactData := NewArtifactData(layerBounds)
	for _, placement := range chunk.Data.Artifacts {
		definition, rotated := prepareArtifactPlacement(placement, assets)
		pixelBounds := artifactPlacementBounds(rotated, placement.X, placement.Z, layerBounds)
		artifact := manager.RegisterChunkArtifact(chunk, definition.Name, definition.Value, placement.X, placement.Z, pixelBounds)
		blitArtifactData(artifactData, rotated, artifact.ID, placement.X, placement.Z)
	}

	return artifactData
}

func prepareArtifactPlacement(placement terrain.ArtifactPlacement, assets artifactDefinitionSource) (*ArtifactDefinition, image.Image) {
	definition := Must(assets.LookupArtifactDefinition(placement.Name))

	sourceImage := Must(assets.LookupImage(definition.ImagePath))
	return definition, rotateImageClockwiseNearest(sourceImage, float64(placement.Orientation))
}

func blitArtifactImage(dst *image.RGBA, src image.Image, centerX, centerY float32) {
	dstRect := artifactPlacementRect(dst, src, centerX, centerY)
	if dstRect.Empty() {
		return
	}

	srcBounds := src.Bounds()
	left := int(math.Round(float64(centerX))) - srcBounds.Dx()/2
	top := int(math.Round(float64(centerY))) - srcBounds.Dy()/2
	for y := dstRect.Min.Y; y < dstRect.Max.Y; y++ {
		srcY := y - top + srcBounds.Min.Y
		for x := dstRect.Min.X; x < dstRect.Max.X; x++ {
			srcX := x - left + srcBounds.Min.X
			pixel := color.RGBAModel.Convert(src.At(srcX, srcY)).(color.RGBA)
			if pixel.A == 0 {
				continue
			}

			dst.SetRGBA(x, y, pixel)
		}
	}
}

func blitArtifactData(artifactData *ArtifactData, src image.Image, artifactID int32, centerX, centerY float32) {
	dstRect := artifactPlacementRect(artifactData, src, centerX, centerY)
	if dstRect.Empty() {
		return
	}

	srcBounds := src.Bounds()
	left := int(math.Round(float64(centerX))) - srcBounds.Dx()/2
	top := int(math.Round(float64(centerY))) - srcBounds.Dy()/2
	for y := dstRect.Min.Y; y < dstRect.Max.Y; y++ {
		srcY := y - top + srcBounds.Min.Y
		for x := dstRect.Min.X; x < dstRect.Max.X; x++ {
			srcX := x - left + srcBounds.Min.X
			pixel := color.RGBAModel.Convert(src.At(srcX, srcY)).(color.RGBA)
			if pixel.A == 0 {
				continue
			}

			artifactData.SetID(x, y, artifactID)
		}
	}
}

func artifactPlacementRect(dst artifactBoundsProvider, src image.Image, centerX, centerY float32) image.Rectangle {
	srcBounds := src.Bounds()
	if srcBounds.Empty() {
		return image.Rectangle{}
	}

	dstBounds := dst.Bounds()
	if dstBounds.Empty() {
		return image.Rectangle{}
	}

	left := int(math.Round(float64(centerX))) - srcBounds.Dx()/2
	top := int(math.Round(float64(centerY))) - srcBounds.Dy()/2
	return image.Rect(left, top, left+srcBounds.Dx(), top+srcBounds.Dy()).Intersect(dstBounds)
}

func artifactPlacementBounds(src image.Image, centerX, centerY float32, layerBounds image.Rectangle) image.Rectangle {
	return artifactPlacementRect(boundsRectangle{bounds: layerBounds}, src, centerX, centerY)
}

type boundsRectangle struct {
	bounds image.Rectangle
}

func (rect boundsRectangle) Bounds() image.Rectangle {
	return rect.bounds
}
