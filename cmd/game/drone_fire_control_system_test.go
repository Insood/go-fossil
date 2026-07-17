package main

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestDroneFireControlCursorUsesMouseWhenGamepadIsIdle(t *testing.T) {
	t.Parallel()

	viewport := rl.NewRectangle(100, 200, 256, 256)
	mouse := rl.NewVector2(20, 500)

	cursor, usingGamepad := droneFireControlCursor(mouse, 0, 0, viewport)
	if usingGamepad {
		t.Fatal("expected mouse control when gamepad is idle")
	}

	if cursor.X != viewport.X || cursor.Y != viewport.Y+viewport.Height {
		t.Fatalf("cursor = (%.2f, %.2f), want viewport-clamped corner", cursor.X, cursor.Y)
	}
}

func TestDroneFireControlCursorUsesGamepadWhenAxesMove(t *testing.T) {
	t.Parallel()

	viewport := rl.NewRectangle(100, 200, 256, 256)

	cursor, usingGamepad := droneFireControlCursor(rl.NewVector2(0, 0), 1, -1, viewport)
	if !usingGamepad {
		t.Fatal("expected gamepad control when axes move")
	}

	if cursor.X != viewport.X+viewport.Width || cursor.Y != viewport.Y {
		t.Fatalf("cursor = (%.2f, %.2f), want viewport edge", cursor.X, cursor.Y)
	}
}

func TestDroneFireControlCursorUsesSmallGamepadAxesWithoutDeadzone(t *testing.T) {
	t.Parallel()

	viewport := rl.NewRectangle(100, 200, 256, 256)

	cursor, usingGamepad := droneFireControlCursor(rl.NewVector2(0, 0), 0.05, -0.05, viewport)
	if !usingGamepad {
		t.Fatal("expected gamepad control for small non-zero axes")
	}

	wantX := viewport.X + (0.05+1)*0.5*viewport.Width
	wantY := viewport.Y + (-0.05+1)*0.5*viewport.Height
	if cursor.X != wantX || cursor.Y != wantY {
		t.Fatalf("cursor = (%.2f, %.2f), want (%.2f, %.2f)", cursor.X, cursor.Y, wantX, wantY)
	}
}
