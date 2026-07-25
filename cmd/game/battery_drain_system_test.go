package main

import (
	"testing"

	ecs "github.com/mlange-42/ark/ecs"
)

func TestDroneBatteryDrainPerSecondUsesCargoAndMovementModifiers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cargoWeight int
		moving      bool
		want        float32
	}{
		{
			name: "empty and idle",
			want: droneBatteryIdleDrainPerSecond,
		},
		{
			name:        "full cargo and idle",
			cargoWeight: droneMaximumCarryWeight,
			want:        droneBatteryIdleDrainPerSecond * (1 + droneBatteryCargoDrainModifier),
		},
		{
			name:   "empty and moving",
			moving: true,
			want:   droneBatteryIdleDrainPerSecond * droneBatteryMovementDrainModifier,
		},
		{
			name:        "full cargo and moving",
			cargoWeight: droneMaximumCarryWeight,
			moving:      true,
			want: droneBatteryIdleDrainPerSecond *
				(1 + droneBatteryCargoDrainModifier) *
				droneBatteryMovementDrainModifier,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := droneBatteryDrainPerSecond(test.cargoWeight, test.moving); got != test.want {
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
	if got, want := battery.charge, float32(0.25); got != want {
		t.Fatalf("battery charge = %v, want %v", got, want)
	}

	game.FrameTime = 1
	system.Update(game)
	if battery.charge != 0 {
		t.Fatalf("battery charge = %v, want 0", battery.charge)
	}
}
