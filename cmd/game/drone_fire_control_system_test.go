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

	if cursor.X != -1 || cursor.Y != 1 {
		t.Fatalf("cursor = (%.2f, %.2f), want normalized clamped corner", cursor.X, cursor.Y)
	}
}

func TestDroneFireControlCursorUsesGamepadWhenAxesMove(t *testing.T) {
	t.Parallel()

	viewport := rl.NewRectangle(100, 200, 256, 256)

	cursor, usingGamepad := droneFireControlCursor(rl.NewVector2(0, 0), 1, -1, viewport)
	if !usingGamepad {
		t.Fatal("expected gamepad control when axes move")
	}

	if cursor.X != 1 || cursor.Y != -1 {
		t.Fatalf("cursor = (%.2f, %.2f), want normalized viewport edge", cursor.X, cursor.Y)
	}
}

func TestDroneFireControlCursorUsesSmallGamepadAxesWithoutDeadzone(t *testing.T) {
	t.Parallel()

	viewport := rl.NewRectangle(100, 200, 256, 256)

	cursor, usingGamepad := droneFireControlCursor(rl.NewVector2(0, 0), 0.05, -0.05, viewport)
	if !usingGamepad {
		t.Fatal("expected gamepad control for small non-zero axes")
	}

	wantX := float32(0.05)
	wantY := float32(-0.05)
	if cursor.X != wantX || cursor.Y != wantY {
		t.Fatalf("cursor = (%.2f, %.2f), want (%.2f, %.2f)", cursor.X, cursor.Y, wantX, wantY)
	}
}
