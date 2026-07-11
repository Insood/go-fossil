package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type DroneInputSystem struct {
	filter *ecs.Filter2[Velocity3, Drone]
}

func (system *DroneInputSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter2[Velocity3, Drone](game.world)
}

func (system *DroneInputSystem) Update(game *Game) {
	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		velocity, _ := query.Get()

		input := rl.NewVector3(0, 0, 0)

		if rl.IsKeyDown(rl.KeyD) {
			input.X += 1
		}
		if rl.IsKeyDown(rl.KeyA) {
			input.X -= 1
		}
		if rl.IsKeyDown(rl.KeyW) {
			input.Z -= 1
		}
		if rl.IsKeyDown(rl.KeyS) {
			input.Z += 1
		}

		if rl.IsGamepadAvailable(droneGamepadIndex) {
			input.X += applyGamepadDeadzone(rl.GetGamepadAxisMovement(droneGamepadIndex, droneGamepadAxisX))
			input.Z += applyGamepadDeadzone(rl.GetGamepadAxisMovement(droneGamepadIndex, droneGamepadAxisZ))
		}

		input.X = clampFloat32(input.X, -1, 1)
		input.Z = clampFloat32(input.Z, -1, 1)

		if length := rl.Vector3Length(input); length > 1 {
			input = rl.Vector3Scale(rl.Vector3Normalize(input), droneTopSpeed)
		} else if length > 0 {
			input = rl.Vector3Scale(input, droneTopSpeed)
		}

		velocity.X = input.X
		velocity.Y = input.Y
		velocity.Z = input.Z
	}
}

func applyGamepadDeadzone(value float32) float32 {
	if float32(math.Abs(float64(value))) < droneGamepadDeadzone {
		return 0
	}

	return value
}

func clampFloat32(value, min, max float32) float32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}

	return value
}
