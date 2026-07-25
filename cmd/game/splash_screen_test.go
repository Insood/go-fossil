package main

import (
	"testing"

	ecs "github.com/mlange-42/ark/ecs"
)

func TestSplashSystemsInitializeAndUnloadInOrder(t *testing.T) {
	t.Parallel()

	initializeOrder := make([]int, 0, 3)
	unloadOrder := make([]int, 0, 3)
	screen := &SplashScreen{}
	screen.systems = []SplashSystem{
		&stubSplashSystem{id: 1, initializeOrder: &initializeOrder, unloadOrder: &unloadOrder},
		&stubSplashSystem{id: 2, initializeOrder: &initializeOrder, unloadOrder: &unloadOrder},
		&stubSplashSystem{id: 3, initializeOrder: &initializeOrder, unloadOrder: &unloadOrder},
	}

	screen.InitializeSystems()
	assertIntSlice(t, initializeOrder, []int{1, 2, 3})

	screen.Unload()
	assertIntSlice(t, unloadOrder, []int{3, 2, 1})
	if screen.systems != nil {
		t.Fatal("splash systems were not released")
	}
	if screen.world != nil {
		t.Fatal("splash world was not released")
	}
}

func TestSplashStartRequested(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		spacePressed   bool
		gamepadPressed bool
		want           bool
	}{
		{name: "no input"},
		{name: "space", spacePressed: true, want: true},
		{name: "gamepad A", gamepadPressed: true, want: true},
		{name: "both", spacePressed: true, gamepadPressed: true, want: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := splashStartRequested(test.spacePressed, test.gamepadPressed); got != test.want {
				t.Fatalf("splashStartRequested(%v, %v) = %v, want %v",
					test.spacePressed, test.gamepadPressed, got, test.want)
			}
		})
	}
}

func TestSplashScreenCreatesRequiredText(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	screen := &SplashScreen{world: world}
	screen.spawnText()

	texts := map[string]SplashText{}
	filter := ecs.NewFilter1[SplashText](world)
	query := filter.Query()
	for query.Next() {
		text := *query.Get()
		texts[text.Text] = text
	}
	query.Close()

	if got := texts[splashTitleText]; got.FontSize != splashTitleFontSize || got.Y != splashTitleY {
		t.Fatalf("title text = %+v, want font size %d and Y %d", got, splashTitleFontSize, splashTitleY)
	}
	if got := texts[splashStartPromptText]; got.FontSize != splashStartPromptFontSize || got.Y != splashStartPromptY {
		t.Fatalf("start prompt = %+v, want font size %d and Y %d", got, splashStartPromptFontSize, splashStartPromptY)
	}
}

type stubSplashSystem struct {
	id              int
	initializeOrder *[]int
	unloadOrder     *[]int
}

func (system *stubSplashSystem) Initialize(*SplashScreen) {
	*system.initializeOrder = append(*system.initializeOrder, system.id)
}

func (system *stubSplashSystem) Update(*SplashScreen) {}

func (system *stubSplashSystem) Unload() {
	*system.unloadOrder = append(*system.unloadOrder, system.id)
}

func assertIntSlice(t *testing.T, got, want []int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("slice length = %d, want %d: got %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("slice[%d] = %d, want %d: got %v", i, got[i], want[i], got)
		}
	}
}
