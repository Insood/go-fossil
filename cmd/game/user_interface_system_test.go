package main

import (
	"fmt"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestGameOverScreenLayoutCentersTitleScoreAndReturnPrompt(t *testing.T) {
	t.Parallel()

	titleX, titleY, scoreText, scoreX, scoreY, promptX, promptY := gameOverScreenLayout(1234)
	if got, want := scoreText, "Score: 1234"; got != want {
		t.Fatalf("score text = %q, want %q", got, want)
	}
	if got, want := titleX, int32((screenWidth-rl.MeasureText(gameOverTitleText, gameOverTitleFontSize))/2); got != want {
		t.Fatalf("title X = %d, want %d", got, want)
	}
	if got, want := scoreX, int32((screenWidth-rl.MeasureText(fmt.Sprintf("Score: %d", 1234), gameOverScoreFontSize))/2); got != want {
		t.Fatalf("score X = %d, want %d", got, want)
	}
	if got, want := promptX, int32((screenWidth-rl.MeasureText(gameOverReturnPromptText, gameOverReturnPromptFontSize))/2); got != want {
		t.Fatalf("prompt X = %d, want %d", got, want)
	}
	if got, want := scoreY-titleY, gameOverTitleFontSize+gameOverTextGap; got != want {
		t.Fatalf("title-to-score Y distance = %d, want %d", got, want)
	}
	if got, want := promptY-scoreY, gameOverScoreFontSize+gameOverTextGap; got != want {
		t.Fatalf("score-to-prompt Y distance = %d, want %d", got, want)
	}
}

func TestDroneStatusBarLayoutAlignsCargoAboveBattery(t *testing.T) {
	t.Parallel()

	viewport := rl.NewRectangle(100, 200, 256, 256)
	batteryBarX, batteryBarY, batteryBarWidth, batteryLabelX, _ := droneStatusBarLayout(viewport, 0)
	cargoBarX, cargoBarY, cargoBarWidth, cargoLabelX, _ := droneStatusBarLayout(viewport, 1)

	if cargoBarX != batteryBarX || cargoBarWidth != batteryBarWidth {
		t.Fatalf(
			"cargo bar x/width = %d/%d, want battery x/width %d/%d",
			cargoBarX,
			cargoBarWidth,
			batteryBarX,
			batteryBarWidth,
		)
	}
	if cargoLabelX != batteryLabelX {
		t.Fatalf("cargo label x = %d, want battery label x %d", cargoLabelX, batteryLabelX)
	}
	if got, want := batteryBarY-cargoBarY, int32(droneBatteryBarHeight+droneBatteryBarGap); got != want {
		t.Fatalf("status bar row distance = %d, want %d", got, want)
	}
	if got, want := batteryBarY, int32(viewport.Y)-droneBatteryBarGap-droneBatteryBarHeight; got != want {
		t.Fatalf("battery bar y = %d, want %d", got, want)
	}
}

func TestDroneCargoPercentUsesCollectedUndroppedWeight(t *testing.T) {
	t.Parallel()

	manager := NewArtifactManager()
	if got := droneCargoPercent(manager); got != 0 {
		t.Fatalf("empty cargo percent = %v, want 0", got)
	}

	manager.fragments[1] = &ArtifactFragment{ID: 1, Weight: droneMaximumCarryWeight / 4, Collected: true}
	manager.fragments[2] = &ArtifactFragment{ID: 2, Weight: droneMaximumCarryWeight, Collected: false}
	manager.fragments[3] = &ArtifactFragment{ID: 3, Weight: droneMaximumCarryWeight, Collected: true, DroppedOff: true}
	if got, want := droneCargoPercent(manager), float32(0.25); got != want {
		t.Fatalf("partial cargo percent = %v, want %v", got, want)
	}

	manager.fragments[4] = &ArtifactFragment{ID: 4, Weight: droneMaximumCarryWeight, Collected: true}
	if got := droneCargoPercent(manager); got != 1 {
		t.Fatalf("over-capacity cargo percent = %v, want 1", got)
	}
}

func TestDroneBatteryBarColorTurnsRedBelowTwentyPercent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		percent float32
		want    rl.Color
	}{
		{name: "empty", percent: 0, want: rl.Red},
		{name: "below threshold", percent: 0.199, want: rl.Red},
		{name: "at threshold", percent: 0.2, want: rl.Lime},
		{name: "above threshold", percent: 0.75, want: rl.Lime},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := droneBatteryBarColor(test.percent); got != test.want {
				t.Fatalf("battery bar color = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestArtifactFragmentGradeDisplay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		grade     float64
		wantText  string
		wantColor rl.Color
	}{
		{name: "perfect", grade: 1, wantText: "[SUPER]", wantColor: rl.Gold},
		{name: "a upper bound", grade: 0.999, wantText: "[A]", wantColor: rl.Green},
		{name: "a lower bound", grade: 0.9, wantText: "[A]", wantColor: rl.Green},
		{name: "b upper bound", grade: 0.899, wantText: "[B]", wantColor: rl.Yellow},
		{name: "b lower bound", grade: 0.8, wantText: "[B]", wantColor: rl.Yellow},
		{name: "c upper bound", grade: 0.799, wantText: "[C]", wantColor: rl.Orange},
		{name: "c lower bound", grade: 0.7, wantText: "[C]", wantColor: rl.Orange},
		{name: "f upper bound", grade: 0.699, wantText: "[F]", wantColor: rl.Red},
		{name: "zero", grade: 0, wantText: "[F]", wantColor: rl.Red},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			gotText, gotColor := artifactFragmentGradeDisplay(test.grade)
			if gotText != test.wantText {
				t.Fatalf("grade text = %q, want %q", gotText, test.wantText)
			}
			if gotColor != test.wantColor {
				t.Fatalf("grade color = %#v, want %#v", gotColor, test.wantColor)
			}
		})
	}
}

func TestSortedArtifactFragmentsOrdersByID(t *testing.T) {
	t.Parallel()

	manager := NewArtifactManager()
	manager.fragments[3] = &ArtifactFragment{ID: 3, Collected: true}
	manager.fragments[1] = &ArtifactFragment{ID: 1, Collected: true}
	manager.fragments[2] = &ArtifactFragment{ID: 2, Collected: true}
	manager.fragments[4] = &ArtifactFragment{ID: 4}
	manager.fragments[5] = &ArtifactFragment{ID: 5, Collected: true, DroppedOff: true}

	fragments := sortedArtifactFragments(manager)
	if len(fragments) != 3 {
		t.Fatalf("fragment count = %d, want 3", len(fragments))
	}
	if got, want := fragments[0].ID, int32(1); got != want {
		t.Fatalf("first fragment id = %d, want %d", got, want)
	}
	if got, want := fragments[1].ID, int32(2); got != want {
		t.Fatalf("second fragment id = %d, want %d", got, want)
	}
	if got, want := fragments[2].ID, int32(3); got != want {
		t.Fatalf("third fragment id = %d, want %d", got, want)
	}
}

func TestRecentArtifactFragmentsReturnsNewestFirstWithLimit(t *testing.T) {
	t.Parallel()

	manager := NewArtifactManager()
	for id := int32(1); id <= 10; id++ {
		manager.fragments[id] = &ArtifactFragment{ID: id, Collected: true}
	}
	manager.fragments[11] = &ArtifactFragment{ID: 11}

	fragments := recentArtifactFragments(manager, 8)
	if len(fragments) != 8 {
		t.Fatalf("fragment count = %d, want 8", len(fragments))
	}

	want := []int32{10, 9, 8, 7, 6, 5, 4, 3}
	for i, fragment := range fragments {
		if got, wantID := fragment.ID, want[i]; got != wantID {
			t.Fatalf("fragment %d id = %d, want %d", i, got, wantID)
		}
	}
}
