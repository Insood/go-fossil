package main

import (
	"fmt"
	"image"
	"os"

	ecs "github.com/mlange-42/ark/ecs"
)

type ArtifactCutoutDetectionSystem struct {
	filter    *ecs.Filter2[TerrainChunkComponent, TerrainChunkDamaged]
	damageMap *ecs.Map[TerrainChunkDamaged]
}

type artifactRegion struct {
	tag  int32
	size int
}

func (system *ArtifactCutoutDetectionSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter2[TerrainChunkComponent, TerrainChunkDamaged](game.world)
	system.damageMap = ecs.NewMap[TerrainChunkDamaged](game.world)
}

func (system *ArtifactCutoutDetectionSystem) Update(game *Game) {
	if game.Tick%artifactCutoutDetectionScanTicks != artifactCutoutDetectionScanTicks-1 {
		return
	}

	query := system.filter.Query()
	defer query.Close()
	for query.Next() {
		chunkComponent, _ := query.Get()
		fmt.Fprintf(os.Stdout, "terrain chunk damaged: %s\n", chunkComponent.Chunk.Coords.String())
		for _, region := range detectArtifactRegions(chunkComponent.Chunk.ArtifactData) {
			fmt.Fprintf(os.Stdout, "artifact region %d size: %d\n", region.tag, region.size)
		}
	}

	system.damageMap.RemoveBatch(system.filter.Batch(), nil)
}

func detectArtifactRegions(data *ArtifactData) []artifactRegion {
	clone := data.Clone()
	bounds := clone.Bounds()
	regions := make([]artifactRegion, 0)
	nextTag := int32(-2)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if clone.IDAt(x, y) <= 0 {
				continue
			}

			regions = append(regions, artifactRegion{
				tag:  nextTag,
				size: floodFillArtifactRegion(clone, x, y, nextTag),
			})
			nextTag--
		}
	}

	return regions
}

func floodFillArtifactRegion(data *ArtifactData, startX, startY int, tag int32) int {
	stack := []image.Point{{X: startX, Y: startY}}
	bounds := data.Bounds()
	size := 0

	for len(stack) > 0 {
		point := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if !point.In(bounds) {
			continue
		}
		if data.IDAt(point.X, point.Y) < 0 {
			continue
		}

		data.SetID(point.X, point.Y, tag)
		size++

		stack = append(stack,
			image.Point{X: point.X - 1, Y: point.Y},
			image.Point{X: point.X + 1, Y: point.Y},
			image.Point{X: point.X, Y: point.Y - 1},
			image.Point{X: point.X, Y: point.Y + 1},
		)
	}

	return size
}
