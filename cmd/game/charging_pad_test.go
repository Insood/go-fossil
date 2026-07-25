package main

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

func TestNearestChargingPadPositionUsesXZDistance(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	mapper := ecs.NewMap2[Position3, ChargingPad](world)
	mapper.NewEntity(
		&Position3{X: 1, Y: 100, Z: 1},
		&ChargingPad{},
	)
	mapper.NewEntity(
		&Position3{X: 4, Y: 0, Z: 4},
		&ChargingPad{},
	)

	filter := ecs.NewFilter2[Position3, ChargingPad](world)
	position, ok := nearestChargingPadPosition(filter, rl.NewVector3(0, 0, 0))
	if !ok {
		t.Fatal("nearest charging pad was not found")
	}
	assertVector3(t, position, 1, 100, 1)
}
