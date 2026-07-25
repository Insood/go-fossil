package main

import (
	ecs "github.com/mlange-42/ark/ecs"
)

type BatteryDrainSystem struct {
	filter          *ecs.Filter3[Velocity3, Drone, Battery]
	artifactManager *ArtifactManager
}

func (system *BatteryDrainSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter3[Velocity3, Drone, Battery](game.world)
	system.artifactManager = game.artifactManager
}

func (system *BatteryDrainSystem) Update(game *Game) {
	if game.FrameTime <= 0 {
		return
	}

	cargoWeight := system.artifactManager.CarriedFragmentWeight()
	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		velocity, _, battery := query.Get()
		moving := velocity.X != 0 || velocity.Z != 0
		battery.charge -= droneBatteryDrainPerSecond(cargoWeight, moving) * game.FrameTime
		if battery.charge < 0 {
			battery.charge = 0
		}
	}
}

func (system *BatteryDrainSystem) Unload() {}

func droneBatteryDrainPerSecond(cargoWeight int, moving bool) float32 {
	cargoPercent := clampFloat32(
		float32(cargoWeight)/float32(droneMaximumCarryWeight),
		0,
		1,
	)
	weightModifier := 1 + cargoPercent*droneBatteryCargoDrainModifier
	movementModifier := float32(1)
	if moving {
		movementModifier = droneBatteryMovementDrainModifier
	}

	return droneBatteryIdleDrainPerSecond * weightModifier * movementModifier
}
