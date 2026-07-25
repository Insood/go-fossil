package main

import "testing"

func TestReturnToMenuRequestedRequiresGameOverAndStartInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		gameOver       bool
		spacePressed   bool
		gamepadPressed bool
		want           bool
	}{
		{name: "no game over with space", spacePressed: true},
		{name: "no game over with gamepad A", gamepadPressed: true},
		{name: "game over without input", gameOver: true},
		{name: "game over with space", gameOver: true, spacePressed: true, want: true},
		{name: "game over with gamepad A", gameOver: true, gamepadPressed: true, want: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := returnToMenuRequested(test.gameOver, test.spacePressed, test.gamepadPressed); got != test.want {
				t.Fatalf(
					"returnToMenuRequested(%v, %v, %v) = %v, want %v",
					test.gameOver,
					test.spacePressed,
					test.gamepadPressed,
					got,
					test.want,
				)
			}
		})
	}
}

func TestHandleGameInputKeepsQuitDistinctFromMenuReturn(t *testing.T) {
	t.Parallel()

	game := &Game{Running: true}
	handleGameInput(game, true, true, true, true)

	if game.Running {
		t.Fatal("game still running after quit input")
	}
	if game.ReturnToMenuRequested {
		t.Fatal("quit input unexpectedly requested the main menu")
	}
}
