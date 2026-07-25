package main

import (
	"testing"

	ecs "github.com/mlange-42/ark/ecs"
)

func TestGameOverDetectionDisablesAndDropsDepletedDrone(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	entity := ecs.NewMap5[Drone, PlayerControlled, HoverMotion, Velocity3, Battery](world).NewEntity(
		&Drone{},
		&PlayerControlled{},
		&HoverMotion{},
		&Velocity3{X: 2, Y: 3, Z: -4},
		&Battery{},
	)
	game := &Game{world: world}
	system := &GameOverDetectionSystem{}
	system.Initialize(game)

	system.Update(game)

	if ecs.NewMap[PlayerControlled](world).Get(entity) != nil {
		t.Fatal("depleted drone still has PlayerControlled")
	}
	if ecs.NewMap[HoverMotion](world).Get(entity) != nil {
		t.Fatal("depleted drone still has HoverMotion")
	}
	velocity := ecs.NewMap[Velocity3](world).Get(entity)
	if got, want := velocity.X, float32(0); got != want {
		t.Fatalf("velocity X = %v, want %v", got, want)
	}
	if got, want := velocity.Y, float32(droneGameOverFallSpeed); got != want {
		t.Fatalf("velocity Y = %v, want %v", got, want)
	}
	if got, want := velocity.Z, float32(0); got != want {
		t.Fatalf("velocity Z = %v, want %v", got, want)
	}
	if got := gameOverEntityCount(world); got != 1 {
		t.Fatalf("game-over entity count = %d, want 1", got)
	}

	system.Update(game)
	if got := gameOverEntityCount(world); got != 1 {
		t.Fatalf("game-over entity count after second update = %d, want 1", got)
	}
}

func TestGameOverDetectionLeavesChargedDroneControlled(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	entity := ecs.NewMap5[Drone, PlayerControlled, HoverMotion, Velocity3, Battery](world).NewEntity(
		&Drone{},
		&PlayerControlled{},
		&HoverMotion{},
		&Velocity3{X: 2, Y: 3, Z: -4},
		&Battery{charge: 1},
	)
	game := &Game{world: world}
	system := &GameOverDetectionSystem{}
	system.Initialize(game)

	system.Update(game)

	if ecs.NewMap[PlayerControlled](world).Get(entity) == nil {
		t.Fatal("charged drone lost PlayerControlled")
	}
	if ecs.NewMap[HoverMotion](world).Get(entity) == nil {
		t.Fatal("charged drone lost HoverMotion")
	}
	if got := gameOverEntityCount(world); got != 0 {
		t.Fatalf("game-over entity count = %d, want 0", got)
	}
}

func gameOverEntityCount(world *ecs.World) int {
	filter := ecs.NewFilter1[GameOver](world)
	query := filter.Query()
	defer query.Close()

	count := 0
	for query.Next() {
		count++
	}
	return count
}
