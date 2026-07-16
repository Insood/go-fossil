package main

import "testing"

func TestGameTickAdvancesOncePerUpdate(t *testing.T) {
	t.Parallel()

	game := &Game{
		systems: []System{
			&stubSystem{},
			&stubSystem{},
		},
	}

	if got, want := game.Tick, 0; got != want {
		t.Fatalf("initial tick = %d, want %d", got, want)
	}

	game.UpdateSystems()
	if got, want := game.Tick, 1; got != want {
		t.Fatalf("tick after first update = %d, want %d", got, want)
	}

	game.UpdateSystems()
	if got, want := game.Tick, 2; got != want {
		t.Fatalf("tick after second update = %d, want %d", got, want)
	}
}

type stubSystem struct{}

func (system *stubSystem) Initialize(*Game) {}

func (system *stubSystem) Update(*Game) {}

