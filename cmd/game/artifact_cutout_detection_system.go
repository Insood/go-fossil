package main

import (
	"fmt"
	"os"

	ecs "github.com/mlange-42/ark/ecs"
)

type ArtifactCutoutDetectionSystem struct {
	filter    *ecs.Filter2[TerrainChunkComponent, TerrainChunkDamaged]
	damageMap *ecs.Map[TerrainChunkDamaged]
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
	}

	system.damageMap.RemoveBatch(system.filter.Batch(), nil)
}
