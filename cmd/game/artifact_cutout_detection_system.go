package main

import (
	"fmt"
	"image"
	"image/color"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type ArtifactCutoutDetectionSystem struct {
	filter          *ecs.Filter2[TerrainChunkComponent, TerrainChunkDamaged]
	damageMap       *ecs.Map[TerrainChunkDamaged]
	fragmentSpawner *ArtifactFragmentSpawnerFactory
}

type artifactRegion struct {
	tag    int32
	size   int
	points []image.Point
	bounds image.Rectangle
}

type artifactRegionCutout struct {
	chunk  *TerrainChunk
	region artifactRegion
}

func (system *ArtifactCutoutDetectionSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter2[TerrainChunkComponent, TerrainChunkDamaged](game.world)
	system.damageMap = ecs.NewMap[TerrainChunkDamaged](game.world)
	system.fragmentSpawner = NewArtifactFragmentSpawnerFactory(game.world)
}

func (system *ArtifactCutoutDetectionSystem) Update(game *Game) {
	if game.Tick%artifactCutoutDetectionScanTicks != artifactCutoutDetectionScanTicks-1 {
		return
	}

	query := system.filter.Query()
	cutouts := make([]artifactRegionCutout, 0)
	for query.Next() {
		chunkComponent, _ := query.Get()
		fmt.Fprintf(os.Stdout, "terrain chunk damaged: %s\n", chunkComponent.Chunk.Coords.String())
		for _, region := range detectArtifactRegions(chunkComponent.Chunk.ArtifactData) {
			fmt.Fprintf(os.Stdout, "artifact region %d size: %d\n", region.tag, region.size)
			if region.size < MaximumRegionSize {
				cutouts = append(cutouts, artifactRegionCutout{
					chunk:  chunkComponent.Chunk,
					region: region,
				})
			}
		}
	}
	query.Close()

	for _, cutout := range cutouts {
		applyArtifactRegion(game.artifactManager, system.fragmentSpawner, cutout.chunk, cutout.region)
	}

	system.damageMap.RemoveBatch(system.filter.Batch(), nil)
}

func (system *ArtifactCutoutDetectionSystem) Unload() {}

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

			region := floodFillArtifactRegionWithPoints(clone, x, y, nextTag)
			regions = append(regions, region)
			nextTag--
		}
	}

	return regions
}

func floodFillArtifactRegionWithPoints(data *ArtifactData, startX, startY int, tag int32) artifactRegion {
	stack := []image.Point{{X: startX, Y: startY}}
	bounds := data.Bounds()
	region := artifactRegion{
		tag:    tag,
		points: make([]image.Point, 0),
		bounds: image.Rect(startX, startY, startX+1, startY+1),
	}

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
		region.points = append(region.points, point)
		region.size++
		if point.X < region.bounds.Min.X {
			region.bounds.Min.X = point.X
		}
		if point.Y < region.bounds.Min.Y {
			region.bounds.Min.Y = point.Y
		}
		if point.X+1 > region.bounds.Max.X {
			region.bounds.Max.X = point.X + 1
		}
		if point.Y+1 > region.bounds.Max.Y {
			region.bounds.Max.Y = point.Y + 1
		}

		stack = append(stack,
			image.Point{X: point.X - 1, Y: point.Y},
			image.Point{X: point.X + 1, Y: point.Y},
			image.Point{X: point.X, Y: point.Y - 1},
			image.Point{X: point.X, Y: point.Y + 1},
		)
	}

	return region
}

func applyArtifactRegion(
	manager *ArtifactManager,
	fragmentSpawner *ArtifactFragmentSpawnerFactory,
	chunk *TerrainChunk,
	region artifactRegion,
) {
	score, grade := scoreArtifactRegion(manager, chunk, region)
	fragment := manager.CreateFragmentFromRegionWithScore(chunk.SurfaceTexture.BaseImage, chunk.ArtifactImage, region.bounds, region.points, score, grade)
	if fragment != nil {
		fragmentSpawner.Spawn(fragment, artifactFragmentPickupPosition(chunk, region.bounds))
	}

	for _, point := range region.points {
		chunk.ArtifactData.SetID(point.X, point.Y, -1)
		chunk.ArtifactImage.SetRGBA(point.X, point.Y, color.RGBA{})
		chunk.BurnOverlayImage.SetRGBA(point.X, point.Y, color.RGBA{A: dugOutOverlayAlpha})
	}

	chunk.uploadArtifactImageRect(region.bounds)
	chunk.uploadBurnOverlayRect(region.bounds)
}

func artifactFragmentPickupPosition(chunk *TerrainChunk, bounds image.Rectangle) rl.Vector3 {
	worldX := chunk.OriginX + float32(bounds.Min.X+bounds.Max.X)/(2*float32(terrainTexturePixelsPerTile))
	worldZ := chunk.OriginZ + float32(bounds.Min.Y+bounds.Max.Y)/(2*float32(terrainTexturePixelsPerTile))
	worldY := chunk.HeightAtWorldPosition(worldX, worldZ) + artifactFragmentPickupGroundLift
	return rl.NewVector3(worldX, worldY, worldZ)
}

func scoreArtifactRegion(manager *ArtifactManager, chunk *TerrainChunk, region artifactRegion) (float64, float64) {
	if manager == nil || chunk == nil || chunk.ArtifactData == nil || chunk.ArtifactImage == nil || len(region.points) == 0 {
		return 0, 0
	}

	cutPixelsByArtifactID := make(map[int32]int)
	for _, point := range region.points {
		if chunk.ArtifactImage.RGBAAt(point.X, point.Y).A == 0 {
			continue
		}

		artifactID := chunk.ArtifactData.IDAt(point.X, point.Y)
		if artifactID <= 0 {
			continue
		}

		cutPixelsByArtifactID[artifactID]++
	}

	score := 0.0
	recoveredPixels := 0
	artifactPixels := 0
	for artifactID, cutPixels := range cutPixelsByArtifactID {
		artifact, ok := manager.Lookup(artifactID)
		if !ok {
			panic(fmt.Errorf("score artifact region: missing artifact %d", artifactID))
		}
		if artifact.Size <= 0 {
			panic(fmt.Errorf("score artifact region: artifact %d has invalid size %d", artifactID, artifact.Size))
		}

		score += float64(cutPixels) / float64(artifact.Size) * float64(artifact.Value)
		recoveredPixels += min(cutPixels, artifact.Size)
		artifactPixels += artifact.Size
	}

	if artifactPixels == 0 {
		return score, 0
	}

	return score, float64(recoveredPixels) / float64(artifactPixels)
}
