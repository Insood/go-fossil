package main

import (
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

		if rl.Vector3Length(input) > 0 {
			input = rl.Vector3Scale(rl.Vector3Normalize(input), droneTopSpeed)
		}

		velocity.X = input.X
		velocity.Y = input.Y
		velocity.Z = input.Z
	}
}
