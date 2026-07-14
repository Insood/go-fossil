package terrain

import (
	"fmt"
	"math"
)

type SurfaceMesh struct {
	Width         int
	Height        int
	HeightSamples [][]float32
	Vertices      []float32
	Normals       []float32
	Texcoords     []float32
	Indices       []uint16
}

func BuildSurfaceMesh(chunk ChunkData) (*SurfaceMesh, error) {
	vertexCount := (chunk.Width + 1) * (chunk.Height + 1)
	if vertexCount > math.MaxUint16+1 {
		return nil, fmt.Errorf("chunk %q has %d vertices, exceeds uint16 mesh index limit", chunk.Name, vertexCount)
	}

	surface := &SurfaceMesh{
		Width:         chunk.Width,
		Height:        chunk.Height,
		HeightSamples: chunk.HeightSamples,
		Vertices:      make([]float32, 0, vertexCount*3),
		Normals:       make([]float32, vertexCount*3),
		Texcoords:     make([]float32, 0, vertexCount*2),
		Indices:       make([]uint16, 0, chunk.Width*chunk.Height*6),
	}

	halfWidth := float32(chunk.Width) / 2
	halfHeight := float32(chunk.Height) / 2
	for z := 0; z <= chunk.Height; z++ {
		for x := 0; x <= chunk.Width; x++ {
			surface.Vertices = append(surface.Vertices,
				float32(x)-halfWidth,
				chunk.HeightSamples[z][x],
				float32(z)-halfHeight,
			)
			surface.Texcoords = append(surface.Texcoords,
				float32(x)/float32(chunk.Width),
				float32(z)/float32(chunk.Height),
			)
		}
	}

	for z := 0; z < chunk.Height; z++ {
		for x := 0; x < chunk.Width; x++ {
			topLeft := uint16(z*(chunk.Width+1) + x)
			bottomLeft := uint16((z+1)*(chunk.Width+1) + x)
			topRight := topLeft + 1
			bottomRight := bottomLeft + 1

			surface.Indices = append(surface.Indices,
				topLeft, bottomLeft, topRight,
				topRight, bottomLeft, bottomRight,
			)
		}
	}

	accumulateNormals(surface)
	return surface, nil
}

func (surface *SurfaceMesh) SampleHeight(worldX, worldZ float32) float32 {
	if len(surface.HeightSamples) == 0 {
		return 0
	}

	x := clampRange(worldX, 0, float32(surface.Width))
	z := clampRange(worldZ, 0, float32(surface.Height))

	x0 := int(math.Floor(float64(x)))
	z0 := int(math.Floor(float64(z)))
	x1 := minInt(x0+1, surface.Width)
	z1 := minInt(z0+1, surface.Height)

	tx := x - float32(x0)
	tz := z - float32(z0)

	h00 := surface.HeightSamples[z0][x0]
	h10 := surface.HeightSamples[z0][x1]
	h01 := surface.HeightSamples[z1][x0]
	h11 := surface.HeightSamples[z1][x1]

	h0 := h00 + (h10-h00)*tx
	h1 := h01 + (h11-h01)*tx
	return h0 + (h1-h0)*tz
}

func accumulateNormals(surface *SurfaceMesh) {
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

func clampRange(value, minValue, maxValue float32) float32 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
