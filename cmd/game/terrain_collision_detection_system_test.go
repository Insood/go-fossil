package main

import (
	"testing"

	ecs "github.com/mlange-42/ark/ecs"
)

func TestTerrainCollisionDetectionSnapsEntityBelowTerrain(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	chunk := newTestTerrainChunk(ChunkCoords{X: 0, Z: 0}, 0, 0, 2)
	manager := &ChunkManager{chunks: map[ChunkCoords]*TerrainChunk{chunk.Coords: chunk}}
	mapper := ecs.NewMap2[Position3, Velocity3](world)
	below := mapper.NewEntity(
		&Position3{X: 4, Y: 1.5, Z: 4},
		&Velocity3{X: 1, Y: -0.5, Z: 2},
	)
	above := mapper.NewEntity(
		&Position3{X: 4, Y: 3, Z: 4},
		&Velocity3{Y: -0.5},
	)
	game := &Game{world: world, chunkManager: manager}
	system := &TerrainCollisionDetectionSystem{}
	system.Initialize(game)

	system.Update(game)

	belowPosition := ecs.NewMap[Position3](world).Get(below)
	belowVelocity := ecs.NewMap[Velocity3](world).Get(below)
	if got, want := belowPosition.Y, float32(2); got != want {
		t.Fatalf("below entity Y = %v, want %v", got, want)
	}
	if got, want := belowVelocity.Y, float32(0); got != want {
		t.Fatalf("below entity Y velocity = %v, want %v", got, want)
	}
	if belowVelocity.X != 1 || belowVelocity.Z != 2 {
		t.Fatalf("below entity X/Z velocity = %v/%v, want 1/2", belowVelocity.X, belowVelocity.Z)
	}

	abovePosition := ecs.NewMap[Position3](world).Get(above)
	aboveVelocity := ecs.NewMap[Velocity3](world).Get(above)
	if got, want := abovePosition.Y, float32(3); got != want {
		t.Fatalf("above entity Y = %v, want %v", got, want)
	}
	if got, want := aboveVelocity.Y, float32(-0.5); got != want {
		t.Fatalf("above entity Y velocity = %v, want %v", got, want)
	}
}

func TestResolveTerrainCollisionIgnoresEntityAtTerrainLevel(t *testing.T) {
	t.Parallel()

	position := Position3{Y: 2}
	velocity := Velocity3{Y: -0.5}
	if resolveTerrainCollision(&position, &velocity, 2) {
		t.Fatal("entity at terrain level unexpectedly collided")
	}
	if position.Y != 2 || velocity.Y != -0.5 {
		t.Fatalf("position/velocity = %v/%v, want 2/-0.5", position.Y, velocity.Y)
	}
}
