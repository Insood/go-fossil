package terrain

import (
	"fmt"
	"image"
	"math"
)

type Surface struct {
	Width         int
	Height        int
	PixelsPerTile int
	Vertices      []float32
	Normals       []float32
	Texcoords     []float32
	Indices       []uint16
	BaseImage     *image.RGBA
	OverlayImage  *image.RGBA
}

func BuildSurface(level LevelData, tileImages map[string]image.Image, pixelsPerTile int) (*Surface, error) {
	if pixelsPerTile <= 0 {
		return nil, fmt.Errorf("pixelsPerTile must be positive, got %d", pixelsPerTile)
	}

	vertexCount := (level.Width + 1) * (level.Height + 1)
	if vertexCount > math.MaxUint16+1 {
		return nil, fmt.Errorf("level %q has %d vertices, exceeds uint16 mesh index limit", level.Name, vertexCount)
	}

	surface := &Surface{
		Width:         level.Width,
		Height:        level.Height,
		PixelsPerTile: pixelsPerTile,
		Vertices:      make([]float32, 0, vertexCount*3),
		Normals:       make([]float32, vertexCount*3),
		Texcoords:     make([]float32, 0, vertexCount*2),
		Indices:       make([]uint16, 0, level.Width*level.Height*6),
		BaseImage:     image.NewRGBA(image.Rect(0, 0, level.Width*pixelsPerTile, level.Height*pixelsPerTile)),
		OverlayImage:  image.NewRGBA(image.Rect(0, 0, level.Width*pixelsPerTile, level.Height*pixelsPerTile)),
	}

	halfWidth := float32(level.Width) / 2
	halfHeight := float32(level.Height) / 2
	for z := 0; z <= level.Height; z++ {
		for x := 0; x <= level.Width; x++ {
			surface.Vertices = append(surface.Vertices,
				float32(x)-halfWidth,
				level.HeightSamples[z][x],
				float32(z)-halfHeight,
			)
			surface.Texcoords = append(surface.Texcoords,
				float32(x)/float32(level.Width),
				float32(z)/float32(level.Height),
			)
		}
	}

	for z := 0; z < level.Height; z++ {
		for x := 0; x < level.Width; x++ {
			topLeft := uint16(z*(level.Width+1) + x)
			bottomLeft := uint16((z+1)*(level.Width+1) + x)
			topRight := topLeft + 1
			bottomRight := bottomLeft + 1

			surface.Indices = append(surface.Indices,
				topLeft, bottomLeft, topRight,
				topRight, bottomLeft, bottomRight,
			)

			tileDefinition := level.TileDefinitions[level.Tiles[z][x]]
			tileImage, ok := tileImages[tileDefinition]
			if !ok {
				return nil, fmt.Errorf("level %q: missing tile image %q", level.Name, tileDefinition)
			}

			drawScaledTile(surface.BaseImage, x*pixelsPerTile, z*pixelsPerTile, pixelsPerTile, tileImage)
		}
	}

	accumulateNormals(surface)
	return surface, nil
}

func (surface *Surface) WorldToUV(worldX, worldZ float32) (float32, float32) {
	return clampUnit(worldX / float32(surface.Width)), clampUnit(worldZ / float32(surface.Height))
}

func (surface *Surface) WorldToOverlayPixel(worldX, worldZ float32) (int, int) {
	u, v := surface.WorldToUV(worldX, worldZ)
	maxX := surface.OverlayImage.Bounds().Dx() - 1
	maxY := surface.OverlayImage.Bounds().Dy() - 1
	return int(math.Round(float64(u * float32(maxX)))), int(math.Round(float64(v * float32(maxY))))
}

func clampUnit(value float32) float32 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
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

func accumulateNormals(surface *Surface) {
	for i := 0; i < len(surface.Indices); i += 3 {
		a := int(surface.Indices[i]) * 3
		b := int(surface.Indices[i+1]) * 3
		c := int(surface.Indices[i+2]) * 3

		ax, ay, az := surface.Vertices[a], surface.Vertices[a+1], surface.Vertices[a+2]
		bx, by, bz := surface.Vertices[b], surface.Vertices[b+1], surface.Vertices[b+2]
		cx, cy, cz := surface.Vertices[c], surface.Vertices[c+1], surface.Vertices[c+2]

		abx, aby, abz := bx-ax, by-ay, bz-az
		acx, acy, acz := cx-ax, cy-ay, cz-az
		nx := aby*acz - abz*acy
		ny := abz*acx - abx*acz
		nz := abx*acy - aby*acx

		surface.Normals[a] += nx
		surface.Normals[a+1] += ny
		surface.Normals[a+2] += nz
		surface.Normals[b] += nx
		surface.Normals[b+1] += ny
		surface.Normals[b+2] += nz
		surface.Normals[c] += nx
		surface.Normals[c+1] += ny
		surface.Normals[c+2] += nz
	}

	for i := 0; i < len(surface.Normals); i += 3 {
		nx := surface.Normals[i]
		ny := surface.Normals[i+1]
		nz := surface.Normals[i+2]
		length := float32(math.Sqrt(float64(nx*nx + ny*ny + nz*nz)))
		if length == 0 {
			surface.Normals[i+1] = 1
			continue
		}

		surface.Normals[i] /= length
		surface.Normals[i+1] /= length
		surface.Normals[i+2] /= length
	}
}
