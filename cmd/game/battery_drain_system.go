package main

import (
	ecs "github.com/mlange-42/ark/ecs"
)

type BatteryDrainSystem struct {
	filter          *ecs.Filter3[Velocity3, Drone, Battery]
	rechargeMap     *ecs.Map[BatteryRecharge]
	artifactManager *ArtifactManager
}

func (system *BatteryDrainSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter3[Velocity3, Drone, Battery](game.world)
	system.rechargeMap = ecs.NewMap[BatteryRecharge](game.world)
	system.artifactManager = game.artifactManager
}

func (system *BatteryDrainSystem) Update(game *Game) {
	cargoWeight := system.artifactManager.CarriedFragmentWeight()
	query := system.filter.Query()
	completedRecharges := make([]ecs.Entity, 0)

	for query.Next() {
		_, _, battery := query.Get()
		if game.FrameTime > 0 {
			battery.charge -= droneBatteryDrainPerSecond(cargoWeight) * game.FrameTime
		}
		if battery.charge < 0 {
			battery.charge = 0
		}

		recharge := system.rechargeMap.Get(query.Entity())
		if recharge == nil {
			continue
		}

		charge := min(recharge.Charge, float32(batteryRechargePerTick))
		recharge.Charge -= charge
		battery.charge += charge
		if battery.charge > droneBatteryCharge {
			battery.charge = droneBatteryCharge
			completedRecharges = append(completedRecharges, query.Entity())
		} else if recharge.Charge <= 0 {
			completedRecharges = append(completedRecharges, query.Entity())
		}
	}
	query.Close()

	for _, entity := range completedRecharges {
		system.rechargeMap.Remove(entity)
	}
}

func (system *BatteryDrainSystem) Unload() {}

func droneBatteryDrainPerSecond(cargoWeight int) float32 {
	cargoPercent := clampFloat32(
		float32(cargoWeight)/float32(droneMaximumCarryWeight),
		0,
		1,
	)
	weightModifier := cargoPercent * droneBatteryCargoDrainModifier

	return droneBatteryBaseDrainPerSecond + weightModifier
}
