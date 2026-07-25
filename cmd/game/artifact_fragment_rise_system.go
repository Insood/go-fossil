package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type ArtifactFragmentRiseSystem struct {
	filter  *ecs.Filter2[Position3, ArtifactFragmentRiseComponent]
	riseMap *ecs.Map[ArtifactFragmentRiseComponent]
}

func (system *ArtifactFragmentRiseSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter2[Position3, ArtifactFragmentRiseComponent](game.world)
	system.riseMap = ecs.NewMap[ArtifactFragmentRiseComponent](game.world)
}

func (system *ArtifactFragmentRiseSystem) Update(game *Game) {
	system.update(rl.GetFrameTime())
}

func (system *ArtifactFragmentRiseSystem) update(dt float32) {
	query := system.filter.Query()
	completed := make([]ecs.Entity, 0)
	for query.Next() {
		position, rise := query.Get()
		if updateArtifactFragmentRise(rise, position, dt) {
			completed = append(completed, query.Entity())
		}
	}
	query.Close()

	for _, entity := range completed {
		system.riseMap.Remove(entity)
	}
}

func (system *ArtifactFragmentRiseSystem) Unload() {}

func updateArtifactFragmentRise(rise *ArtifactFragmentRiseComponent, position *Position3, dt float32) bool {
	rise.elapsed = minFloat32(rise.elapsed+dt, artifactFragmentPickupRiseDuration)
	progress := rise.elapsed / artifactFragmentPickupRiseDuration
	*position = Position3(rl.Vector3Lerp(rise.startPosition, rise.raisedPosition, easeOutCubic(progress)))
	return rise.elapsed >= artifactFragmentPickupRiseDuration
}
