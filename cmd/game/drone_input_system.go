package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type DroneInputSystem struct {
	filter *ecs.Filter3[Position3, Velocity3, Drone]
}

func (system *DroneInputSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter3[Position3, Velocity3, Drone](game.world)
}

func (system *DroneInputSystem) Update(game *Game) {
	dt := rl.GetFrameTime()
	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		position, velocity, _ := query.Get()

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
			input.X += applyGamepadDeadzone(rl.GetGamepadAxisMovement(droneGamepadIndex, droneGamepadMoveAxisX))
			input.Z += applyGamepadDeadzone(rl.GetGamepadAxisMovement(droneGamepadIndex, droneGamepadMoveAxisZ))
		}

		input.X = clampFloat32(input.X, -1, 1)
		input.Z = clampFloat32(input.Z, -1, 1)

		if length := rl.Vector3Length(input); length > 1 {
			input = rl.Vector3Scale(rl.Vector3Normalize(input), droneTopSpeed)
		} else if length > 0 {
			input = rl.Vector3Scale(input, droneTopSpeed)
		}

		desiredVelocity := Velocity3{X: input.X, Y: input.Y, Z: input.Z}
		desiredVelocity = clampDroneVelocityToTerrainBounds(*position, desiredVelocity, dt, game.chunkManager)

		velocity.X = desiredVelocity.X
		velocity.Y = desiredVelocity.Y
		velocity.Z = desiredVelocity.Z
	}
}

func (system *DroneInputSystem) Unload() {}

func clampDroneVelocityToTerrainBounds(position Position3, velocity Velocity3, dt float32, chunkManager *ChunkManager) Velocity3 {
	if dt <= 0 {
		return velocity
	}

	nextX := position.X + velocity.X*dt
	if !chunkManager.DroneFitsAtWorldPosition(nextX, position.Z) {
		velocity.X = 0
	}

	nextZ := position.Z + velocity.Z*dt
	if !chunkManager.DroneFitsAtWorldPosition(position.X, nextZ) {
		velocity.Z = 0
	}

	return velocity
}

func applyGamepadDeadzone(value float32) float32 {
	if float32(math.Abs(float64(value))) < droneGamepadDeadzone {
		return 0
	}

	return value
}
