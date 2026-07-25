package main

import (
	"math"
	"testing"

	ecs "github.com/mlange-42/ark/ecs"
)

func TestDroneBatteryDrainPerSecondUsesBaseAndCargoModifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cargoWeight int
		want        float32
	}{
		{
			name: "empty and idle",
			want: droneBatteryBaseDrainPerSecond,
		},
		{
			name:        "half cargo",
			cargoWeight: droneMaximumCarryWeight / 2,
			want:        droneBatteryBaseDrainPerSecond + droneBatteryCargoDrainModifier/2,
		},
		{
			name:        "full cargo",
			cargoWeight: droneMaximumCarryWeight,
			want:        droneBatteryBaseDrainPerSecond + droneBatteryCargoDrainModifier,
		},
		{
			name:        "cargo over capacity is clamped",
			cargoWeight: droneMaximumCarryWeight * 2,
			want:        droneBatteryBaseDrainPerSecond + droneBatteryCargoDrainModifier,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := droneBatteryDrainPerSecond(test.cargoWeight); got != test.want {
				t.Fatalf("drain per second = %v, want %v", got, test.want)
			}
		})
	}
}

func TestBatteryDrainSystemUsesElapsedTimeAndClampsAtZero(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()

	manager := NewArtifactManager()
	manager.fragments[1] = &ArtifactFragment{
		Weight:    droneMaximumCarryWeight,
		Collected: true,
	}
	entity := ecs.NewMap3[Velocity3, Drone, Battery](world).NewEntity(
		&Velocity3{X: droneTopSpeed},
		&Drone{},
		&Battery{charge: 0.5},
	)
	game := &Game{world: world, artifactManager: manager, FrameTime: 0.25}
	system := &BatteryDrainSystem{}
	system.Initialize(game)

	system.Update(game)

	battery := ecs.NewMap[Battery](world).Get(entity)
	if got, want := battery.charge, float32(0.1875); got != want {
		t.Fatalf("battery charge = %v, want %v", got, want)
	}

	game.FrameTime = 1
	system.Update(game)
	if battery.charge != 0 {
		t.Fatalf("battery charge = %v, want 0", battery.charge)
	}
}

func TestBatteryDrainSystemAppliesAndExhaustsRecharge(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	entity := ecs.NewMap4[Velocity3, Drone, Battery, BatteryRecharge](world).NewEntity(
		&Velocity3{},
		&Drone{},
		&Battery{charge: 50},
		&BatteryRecharge{Charge: 0.025},
	)
	game := &Game{
		world:           world,
		artifactManager: NewArtifactManager(),
	}
	system := &BatteryDrainSystem{}
	system.Initialize(game)

	system.Update(game)
	system.Update(game)
	system.Update(game)

	battery := ecs.NewMap[Battery](world).Get(entity)
	if got, want := battery.charge, float32(50.025); math.Abs(float64(got-want)) > 0.00001 {
		t.Fatalf("battery charge = %v, want %v", got, want)
	}
	if system.rechargeMap.Has(entity) {
		t.Fatal("exhausted battery recharge component was not removed")
	}
}

func TestBatteryDrainSystemClampsRechargeAndDiscardsOverflow(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	entity := ecs.NewMap4[Velocity3, Drone, Battery, BatteryRecharge](world).NewEntity(
		&Velocity3{},
		&Drone{},
		&Battery{charge: droneBatteryCharge - 0.005},
		&BatteryRecharge{Charge: 1},
	)
	game := &Game{
		world:           world,
		artifactManager: NewArtifactManager(),
	}
	system := &BatteryDrainSystem{}
	system.Initialize(game)

	system.Update(game)

	battery := ecs.NewMap[Battery](world).Get(entity)
	if got, want := battery.charge, float32(droneBatteryCharge); got != want {
		t.Fatalf("battery charge = %v, want %v", got, want)
	}
	if system.rechargeMap.Has(entity) {
		t.Fatal("overflowing battery recharge component was not removed")
	}
}
