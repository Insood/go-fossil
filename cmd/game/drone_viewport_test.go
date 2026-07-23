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

func TestDroneViewportCursorConversionsUseNormalizedCoordinates(t *testing.T) {
	t.Parallel()

	viewport := rl.NewRectangle(100, 200, 256, 256)

	cursor := droneViewportCursorFromMouse(rl.NewVector2(228, 328), viewport)
	if cursor.X != 0 || cursor.Y != 0 {
		t.Fatalf("cursor = (%.2f, %.2f), want normalized center", cursor.X, cursor.Y)
	}

	pixel := droneViewportCursorPixel(rl.NewVector2(1, -1), viewport)
	if pixel.X != viewport.X+viewport.Width || pixel.Y != viewport.Y {
		t.Fatalf("pixel = (%.2f, %.2f), want top-right viewport corner", pixel.X, pixel.Y)
	}
}
