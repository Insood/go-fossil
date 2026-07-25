package main

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
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

func TestSplashChunkCoordsLoadCenterCardinalsThenDiagonals(t *testing.T) {
	t.Parallel()

	got := splashChunkCoords()
	want := []ChunkCoords{
		{X: 0, Z: 0},
		{X: 0, Z: -1},
		{X: 1, Z: 0},
		{X: 0, Z: 1},
		{X: -1, Z: 0},
		{X: -1, Z: -1},
		{X: 1, Z: -1},
		{X: 1, Z: 1},
		{X: -1, Z: 1},
	}

	if len(got) != len(want) {
		t.Fatalf("splash chunk count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("splash chunk %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestSplashDroneIsStaticAndRenderable(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	position := Position3{X: 4, Y: 3, Z: 4}
	entity := spawnSplashDrone(world, &rl.Model{}, position)

	dronePosition := ecs.NewMap[Position3](world).Get(entity)
	if dronePosition == nil || *dronePosition != position {
		t.Fatalf("splash drone position = %+v, want %+v", dronePosition, position)
	}
	if ecs.NewMap[Renderable](world).Get(entity) == nil {
		t.Fatal("splash drone is missing Renderable")
	}
	if ecs.NewMap[Drone](world).Get(entity) == nil {
		t.Fatal("splash drone is missing Drone")
	}
	if ecs.NewMap[HoverMotion](world).Get(entity) == nil {
		t.Fatal("splash drone is missing HoverMotion")
	}

	for name, missing := range map[string]bool{
		"Velocity3":        ecs.NewMap[Velocity3](world).Get(entity) == nil,
		"Battery":          ecs.NewMap[Battery](world).Get(entity) == nil,
		"Laser":            ecs.NewMap[Laser](world).Get(entity) == nil,
		"DroneFireControl": ecs.NewMap[DroneFireControl](world).Get(entity) == nil,
	} {
		if !missing {
			t.Fatalf("splash drone unexpectedly has %s", name)
		}
	}
}

func TestSplashRegistersOnlyScenePresentationSystems(t *testing.T) {
	t.Parallel()

	scene := &Game{}
	screen := &SplashScreen{scene: scene}
	screen.registerSceneSystems()

	if got, want := len(scene.systems), 3; got != want {
		t.Fatalf("splash scene system count = %d, want %d", got, want)
	}
	if _, ok := scene.systems[0].(*DroneHeightSystem); !ok {
		t.Fatalf("splash scene system 0 = %T, want *DroneHeightSystem", scene.systems[0])
	}
	if _, ok := scene.systems[1].(*LightSystem); !ok {
		t.Fatalf("splash scene system 1 = %T, want *LightSystem", scene.systems[1])
	}
	renderSystem, ok := scene.systems[2].(*RenderSystem3D)
	if !ok {
		t.Fatalf("splash scene system 2 = %T, want *RenderSystem3D", scene.systems[2])
	}
	if !renderSystem.skipDroneViewport {
		t.Fatal("splash render system did not skip the drone viewport")
	}
}

func TestSplashSceneBorrowsAssetManager(t *testing.T) {
	t.Parallel()

	assets := &AssetManager{}
	screen := &SplashScreen{scene: &Game{assets: assets}}

	if got := screen.scene.assets; got != assets {
		t.Fatalf("borrowed assets = %p, want %p", got, assets)
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
