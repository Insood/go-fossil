package main

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

func TestArtifactFragmentRiseSystemRaisesFragmentAndLeavesItReady(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	start := rl.NewVector3(1, 2, 3)
	raised := rl.NewVector3(1, 2+artifactFragmentPickupRiseHeight, 3)
	entity := ecs.NewMap2[Position3, ArtifactFragmentRiseComponent](world).NewEntity(
		(*Position3)(&start),
		&ArtifactFragmentRiseComponent{
			startPosition:  start,
			raisedPosition: raised,
		},
	)
	system := &ArtifactFragmentRiseSystem{}
	system.Initialize(&Game{world: world})

	system.update(artifactFragmentPickupRiseDuration / 2)
	position := ecs.NewMap[Position3](world).Get(entity)
	if rl.Vector3(*position) == start || rl.Vector3(*position) == raised {
		t.Fatal("fragment did not move to an intermediate rise position")
	}
	if !system.riseMap.Has(entity) {
		t.Fatal("rise component was removed before the animation completed")
	}

	system.update(artifactFragmentPickupRiseDuration / 2)
	position = ecs.NewMap[Position3](world).Get(entity)
	assertVector3Close(t, rl.Vector3(*position), raised)
	if system.riseMap.Has(entity) {
		t.Fatal("rise component remains after the animation completed")
	}

	system.update(1)
	position = ecs.NewMap[Position3](world).Get(entity)
	assertVector3Close(t, rl.Vector3(*position), raised)
	if !world.Alive(entity) {
		t.Fatal("ready fragment entity was removed")
	}
}
