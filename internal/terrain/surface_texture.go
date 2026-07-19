package terrain

import (
	"fmt"
	"image"

	"github.com/disintegration/gift"
)

type SurfaceTexture struct {
	Width         int
	Height        int
	PixelsPerTile int
	BaseImage     *image.RGBA
}

func BuildSurfaceTexture(chunk ChunkData, tileImages map[string]image.Image, pixelsPerTile int) (*SurfaceTexture, error) {
	if pixelsPerTile <= 0 {
		return nil, fmt.Errorf("pixelsPerTile must be positive, got %d", pixelsPerTile)
	}

	surface := &SurfaceTexture{
		Width:         chunk.Width,
		Height:        chunk.Height,
		PixelsPerTile: pixelsPerTile,
		BaseImage:     image.NewRGBA(image.Rect(0, 0, chunk.Width*pixelsPerTile, chunk.Height*pixelsPerTile)),
	}

	for z := 0; z < chunk.Height; z++ {
		for x := 0; x < chunk.Width; x++ {
			tileDefinition := chunk.TileDefinitions[chunk.Tiles[z][x]]
			tileImage, ok := tileImages[tileDefinition]
			if !ok {
				return nil, fmt.Errorf("chunk %q: missing tile image %q", chunk.Name, tileDefinition)
			}

			drawScaledTile(surface.BaseImage, x*pixelsPerTile, z*pixelsPerTile, pixelsPerTile, tileImage)
		}
	}

	return surface, nil
}

func drawScaledTile(dst *image.RGBA, dstX, dstY, size int, src image.Image) {
	filter := gift.New(gift.Resize(size, size, gift.NearestNeighborResampling))
	filter.DrawAt(dst, src, image.Pt(dstX, dstY), gift.CopyOperator)
}
