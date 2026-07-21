package main

import (
	"testing"

	ecs "github.com/mlange-42/ark/ecs"
)

func TestTutorialSystemStartsAtMoveDroneStep(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	mapper := ecs.NewMap2[Position3, Drone](world)
	mapper.NewEntity(
		&Position3{X: 1, Y: 2, Z: 3},
		&Drone{},
	)

	game := &Game{world: world}
	system := &TutorialSystem{}
	system.Initialize(game)

	if got, want := system.state.currentStep, tutorialStepMoveDrone; got != want {
		t.Fatalf("current step = %d, want %d", got, want)
	}
}

func TestTutorialStepOneIgnoresDroneYChanges(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	mapper := ecs.NewMap2[Position3, Drone](world)
	entity := mapper.NewEntity(
		&Position3{X: 1, Y: 2, Z: 3},
		&Drone{},
	)

	game := &Game{world: world}
	system := &TutorialSystem{}
	system.Initialize(game)

	position, _ := mapper.Get(entity)
	position.Y += 1
	system.Update(game)

	if got, want := system.state.currentStep, tutorialStepMoveDrone; got != want {
		t.Fatalf("current step = %d, want %d", got, want)
	}
}

func TestTutorialStepOneCompletesWhenDroneMovesOnXZ(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	mapper := ecs.NewMap2[Position3, Drone](world)
	entity := mapper.NewEntity(
		&Position3{X: 1, Y: 2, Z: 3},
		&Drone{},
	)

	game := &Game{world: world}
	system := &TutorialSystem{}
	system.Initialize(game)

	position, _ := mapper.Get(entity)
	position.X += 0.001
	system.Update(game)

	if got, want := system.state.currentStep, tutorialStepComplete; got != want {
		t.Fatalf("current step = %d, want %d", got, want)
	}
}
