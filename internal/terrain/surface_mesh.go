package terrain

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
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
	return buildSurfaceMesh(chunk, 1)
}

func BuildSubdividedSurfaceMesh(chunk ChunkData, subdivisionsPerTile int) (*SurfaceMesh, error) {
	return buildSurfaceMesh(chunk, subdivisionsPerTile)
}

func buildSurfaceMesh(chunk ChunkData, subdivisionsPerTile int) (*SurfaceMesh, error) {
	if subdivisionsPerTile < 1 {
		subdivisionsPerTile = 1
	}

	meshWidth := chunk.Width * subdivisionsPerTile
	meshHeight := chunk.Height * subdivisionsPerTile
	vertexCount := (meshWidth + 1) * (meshHeight + 1)
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
		Indices:       make([]uint16, 0, meshWidth*meshHeight*6),
	}

	halfWidth := float32(chunk.Width) / 2
	halfHeight := float32(chunk.Height) / 2
	for z := 0; z <= meshHeight; z++ {
		for x := 0; x <= meshWidth; x++ {
			localX := float32(x) / float32(subdivisionsPerTile)
			localZ := float32(z) / float32(subdivisionsPerTile)
			surface.Vertices = append(surface.Vertices,
				localX-halfWidth,
				sampleHeight(chunk.HeightSamples, chunk.Width, chunk.Height, localX, localZ),
				localZ-halfHeight,
			)
			surface.Texcoords = append(surface.Texcoords,
				localX/float32(chunk.Width),
				localZ/float32(chunk.Height),
			)
		}
	}

	for z := 0; z < meshHeight; z++ {
		for x := 0; x < meshWidth; x++ {
			topLeft := uint16(z*(meshWidth+1) + x)
			bottomLeft := uint16((z+1)*(meshWidth+1) + x)
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

	return sampleHeight(surface.HeightSamples, surface.Width, surface.Height, worldX, worldZ)
}

func sampleHeight(heightSamples [][]float32, width, height int, sampleX, sampleZ float32) float32 {
	x := clampRange(sampleX, 0, float32(width))
	z := clampRange(sampleZ, 0, float32(height))

	x0 := int(math.Floor(float64(x)))
	z0 := int(math.Floor(float64(z)))
	x1 := minInt(x0+1, width)
	z1 := minInt(z0+1, height)

	tx := x - float32(x0)
	tz := z - float32(z0)

	h00 := heightSamples[z0][x0]
	h10 := heightSamples[z0][x1]
	h01 := heightSamples[z1][x0]
	h11 := heightSamples[z1][x1]

	h0 := h00 + (h10-h00)*tx
	h1 := h01 + (h11-h01)*tx
	return h0 + (h1-h0)*tz
}

func accumulateNormals(surface *SurfaceMesh) {
	for i := 0; i < len(surface.Indices); i += 3 {
		a := int(surface.Indices[i]) * 3
		b := int(surface.Indices[i+1]) * 3
		c := int(surface.Indices[i+2]) * 3

		va := vector3FromSlice(surface.Vertices, a)
		vb := vector3FromSlice(surface.Vertices, b)
		vc := vector3FromSlice(surface.Vertices, c)
		normal := rl.Vector3CrossProduct(
			rl.Vector3Subtract(vb, va),
			rl.Vector3Subtract(vc, va),
		)

		addVector3ToSlice(surface.Normals, a, normal)
		addVector3ToSlice(surface.Normals, b, normal)
		addVector3ToSlice(surface.Normals, c, normal)
	}

	for i := 0; i < len(surface.Normals); i += 3 {
		normal := vector3FromSlice(surface.Normals, i)
		if normal == (rl.Vector3{}) {
			surface.Normals[i+1] = 1
			continue
		}

		setVector3InSlice(surface.Normals, i, rl.Vector3Normalize(normal))
	}
}

func vector3FromSlice(values []float32, index int) rl.Vector3 {
	return rl.NewVector3(values[index], values[index+1], values[index+2])
}

func addVector3ToSlice(values []float32, index int, vector rl.Vector3) {
	values[index] += vector.X
	values[index+1] += vector.Y
	values[index+2] += vector.Z
}

func setVector3InSlice(values []float32, index int, vector rl.Vector3) {
	values[index] = vector.X
	values[index+1] = vector.Y
	values[index+2] = vector.Z
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
