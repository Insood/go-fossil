package main

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

func TestLightSystemTracksDroneOverhead(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	droneMapper := ecs.NewMap2[Position3, Drone](world)
	droneMapper.NewEntity(
		&Position3{X: 6, Y: 2, Z: 7},
		&Drone{},
	)

	lightMapper := ecs.NewMap1[Light](world)
	lightMapper.NewEntity(&Light{
		origin:           rl.NewVector3(0, lightHeight, 0),
		target:           rl.NewVector3(0, 0, 0),
		up:               rl.NewVector3(0, 0, -1),
		orthographicSize: defaultLightSize,
	})

	game := &Game{world: world}
	system := &LightSystem{}
	system.Initialize(game)
	system.Update(game)

	lightFilter := ecs.NewFilter1[Light](world)
	query := lightFilter.Query()
	defer query.Close()
	if !query.Next() {
		t.Fatal("light not found")
	}

	light := query.Get()
	assertVector3(t, light.origin, 6, lightHeight, 7)
	assertVector3(t, light.target, 6, 0, 7)
	assertVector3(t, light.camera.Position, 6, lightHeight, 7)
	assertVector3(t, light.camera.Target, 6, 0, 7)
}

func assertVector3(t *testing.T, got rl.Vector3, wantX, wantY, wantZ float32) {
	t.Helper()

	if got.X != wantX || got.Y != wantY || got.Z != wantZ {
		t.Fatalf("vector = (%.2f, %.2f, %.2f), want (%.2f, %.2f, %.2f)", got.X, got.Y, got.Z, wantX, wantY, wantZ)
	}
}
