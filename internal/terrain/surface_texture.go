package terrain

import (
	"fmt"
	"image"
)

type SurfaceTexture struct {
	Width         int
	Height        int
	PixelsPerTile int
	BaseImage     *image.RGBA
}

func BuildSurfaceTexture(level LevelData, tileImages map[string]image.Image, pixelsPerTile int) (*SurfaceTexture, error) {
	if pixelsPerTile <= 0 {
		return nil, fmt.Errorf("pixelsPerTile must be positive, got %d", pixelsPerTile)
	}

	surface := &SurfaceTexture{
		Width:         level.Width,
		Height:        level.Height,
		PixelsPerTile: pixelsPerTile,
		BaseImage:     image.NewRGBA(image.Rect(0, 0, level.Width*pixelsPerTile, level.Height*pixelsPerTile)),
	}

	for z := 0; z < level.Height; z++ {
		for x := 0; x < level.Width; x++ {
			tileDefinition := level.TileDefinitions[level.Tiles[z][x]]
			tileImage, ok := tileImages[tileDefinition]
			if !ok {
				return nil, fmt.Errorf("level %q: missing tile image %q", level.Name, tileDefinition)
			}

			drawScaledTile(surface.BaseImage, x*pixelsPerTile, z*pixelsPerTile, pixelsPerTile, tileImage)
		}
	}

	return surface, nil
}

func drawScaledTile(dst *image.RGBA, dstX, dstY, size int, src image.Image) {
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	for y := 0; y < size; y++ {
		srcY := srcBounds.Min.Y + (y*srcHeight)/size
		for x := 0; x < size; x++ {
			srcX := srcBounds.Min.X + (x*srcWidth)/size
			dst.Set(dstX+x, dstY+y, src.At(srcX, srcY))
		}
	}
}
