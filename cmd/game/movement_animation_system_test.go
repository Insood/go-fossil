package main

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

func TestMovementAnimationEasing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		easing MovementAnimationEasing
		wantX  float32
	}{
		{name: "linear", easing: MovementAnimationLinear, wantX: 4},
		{name: "ease in cubic", easing: MovementAnimationEaseInCubic, wantX: 1},
		{name: "ease out cubic", easing: MovementAnimationEaseOutCubic, wantX: 7},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			position := Position3{}
			movement := &MovementAnimationComponent{
				startPosition:  rl.Vector3{},
				targetPosition: rl.NewVector3(8, 0, 0),
				duration:       1,
				easing:         test.easing,
			}

			if complete := updateMovementAnimation(movement, &position, movement.targetPosition, 0.5); complete {
				t.Fatal("movement completed at its midpoint")
			}
			assertVector3Close(t, rl.Vector3(position), rl.NewVector3(test.wantX, 0, 0))
		})
	}
}

func TestMovementAnimationSystemCompletesAndRemovesComponent(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	mapper := ecs.NewMap2[Position3, MovementAnimationComponent](world)
	entity := mapper.NewEntity(
		&Position3{},
		&MovementAnimationComponent{
			targetPosition: rl.NewVector3(2, 3, 4),
			duration:       0,
		},
	)
	system := &MovementAnimationSystem{}
	system.Initialize(&Game{world: world})

	system.update(0)

	position := ecs.NewMap[Position3](world).Get(entity)
	assertVector3Close(t, rl.Vector3(*position), rl.NewVector3(2, 3, 4))
	if system.movementMap.Has(entity) {
		t.Fatal("completed movement component was not removed")
	}
	if !world.Alive(entity) {
		t.Fatal("movement system removed the entity")
	}
}

func TestMovementAnimationSystemTracksLiveTargetThenUsesFallback(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	positionMap := ecs.NewMap[Position3](world)
	target := positionMap.NewEntity(&Position3{X: 20})
	mover := ecs.NewMap2[Position3, MovementAnimationComponent](world).NewEntity(
		&Position3{},
		&MovementAnimationComponent{
			targetPosition: rl.NewVector3(10, 0, 0),
			targetEntity:   target,
			duration:       1,
		},
	)
	system := &MovementAnimationSystem{}
	system.Initialize(&Game{world: world})

	system.update(0.5)
	position := positionMap.Get(mover)
	assertVector3Close(t, rl.Vector3(*position), rl.NewVector3(10, 0, 0))

	world.RemoveEntity(target)
	system.update(0.5)
	position = positionMap.Get(mover)
	assertVector3Close(t, rl.Vector3(*position), rl.NewVector3(10, 0, 0))
	if system.movementMap.Has(mover) {
		t.Fatal("movement remained after reaching its fallback target")
	}
}

func TestArtifactFragmentRiseMovementLeavesFragmentReadyAtRaisedPosition(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	start := rl.NewVector3(1, 2, 3)
	entity := ecs.NewMap3[Position3, ArtifactFragmentComponent, MovementAnimationComponent](world).NewEntity(
		(*Position3)(&start),
		&ArtifactFragmentComponent{fragment: &ArtifactFragment{ID: 1}},
		artifactFragmentRiseMovement(start),
	)
	system := &MovementAnimationSystem{}
	system.Initialize(&Game{world: world})

	system.update(artifactFragmentPickupRiseDuration)

	position := ecs.NewMap[Position3](world).Get(entity)
	assertVector3Close(t, rl.Vector3(*position), rl.NewVector3(1, 2+artifactFragmentPickupRiseHeight, 3))
	if system.movementMap.Has(entity) {
		t.Fatal("rise movement remains after completion")
	}
	if !world.Alive(entity) {
		t.Fatal("ready fragment entity was removed")
	}
}
