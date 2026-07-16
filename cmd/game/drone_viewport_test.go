package main

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestTerrainRayTargetHitsFlatLoadedTerrain(t *testing.T) {
	t.Parallel()

	ray := rl.NewRay(
		rl.NewVector3(10, 10, 10),
		rl.NewVector3(0, -1, 0),
	)

	target, ok := terrainRayTarget(
		ray,
		func(worldX, worldZ float32) float32 { return 0 },
		func(worldX, worldZ float32) bool { return true },
	)
	if !ok {
		t.Fatal("terrainRayTarget returned false")
	}
	if target.Y < -0.001 || target.Y > 0.001 {
		t.Fatalf("target y = %v, want near 0", target.Y)
	}
}

func TestTerrainRayTargetRejectsUnloadedTerrain(t *testing.T) {
	t.Parallel()

	ray := rl.NewRay(
		rl.NewVector3(10, 10, 10),
		rl.NewVector3(0, -1, 0),
	)

	if _, ok := terrainRayTarget(
		ray,
		func(worldX, worldZ float32) float32 { return 0 },
		func(worldX, worldZ float32) bool { return false },
	); ok {
		t.Fatal("terrainRayTarget returned true for unloaded terrain")
	}
}
