package main

import (
	"math"
	"math/rand"
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

func TestSplashDroneIsSimulatedAndRenderable(t *testing.T) {
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
	velocity := ecs.NewMap[Velocity3](world).Get(entity)
	if velocity == nil || *velocity != (Velocity3{}) {
		t.Fatalf("initial splash drone velocity = %+v, want zero", velocity)
	}
	if ecs.NewMap[Laser](world).Get(entity) == nil {
		t.Fatal("splash drone is missing Laser")
	}
	if ecs.NewMap[DroneFireTargets](world).Get(entity) == nil {
		t.Fatal("splash drone is missing DroneFireTargets")
	}
	battery := ecs.NewMap[Battery](world).Get(entity)
	if battery == nil || battery.charge != splashDroneBatteryCharge {
		t.Fatalf("splash drone battery = %+v, want charge %v", battery, splashDroneBatteryCharge)
	}

	for name, missing := range map[string]bool{
		"PlayerFireInput": ecs.NewMap[PlayerFireInput](world).Get(entity) == nil,
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

	if got, want := len(scene.systems), 8; got != want {
		t.Fatalf("splash scene system count = %d, want %d", got, want)
	}
	if _, ok := scene.systems[0].(*SplashScreenDroneControlSystem); !ok {
		t.Fatalf("splash scene system 0 = %T, want *SplashScreenDroneControlSystem", scene.systems[0])
	}
	if _, ok := scene.systems[1].(*PhysicsSystem); !ok {
		t.Fatalf("splash scene system 1 = %T, want *PhysicsSystem", scene.systems[1])
	}
	if _, ok := scene.systems[2].(*DroneHeightSystem); !ok {
		t.Fatalf("splash scene system 2 = %T, want *DroneHeightSystem", scene.systems[2])
	}
	if _, ok := scene.systems[3].(*LightSystem); !ok {
		t.Fatalf("splash scene system 3 = %T, want *LightSystem", scene.systems[3])
	}
	if _, ok := scene.systems[4].(*SplashScreenDroneFireTargetSystem); !ok {
		t.Fatalf("splash scene system 4 = %T, want *SplashScreenDroneFireTargetSystem", scene.systems[4])
	}
	if _, ok := scene.systems[5].(*LaserSystem); !ok {
		t.Fatalf("splash scene system 5 = %T, want *LaserSystem", scene.systems[5])
	}
	if _, ok := scene.systems[6].(*ParticleSystem); !ok {
		t.Fatalf("splash scene system 6 = %T, want *ParticleSystem", scene.systems[6])
	}
	renderSystem, ok := scene.systems[7].(*RenderSystem3D)
	if !ok {
		t.Fatalf("splash scene system 7 = %T, want *RenderSystem3D", scene.systems[7])
	}
	if !renderSystem.skipDroneViewport {
		t.Fatal("splash render system did not skip the drone viewport")
	}
}

func TestSplashDroneDirectionTimerChoosesImmediatelyAndEverySecond(t *testing.T) {
	t.Parallel()

	elapsed, choose := advanceSplashDroneDirectionTimer(splashDroneDirectionDuration, 0)
	if !choose || elapsed != 0 {
		t.Fatalf("initial timer = (%v, %v), want (0, true)", elapsed, choose)
	}

	elapsed, choose = advanceSplashDroneDirectionTimer(elapsed, 0.5)
	if choose || elapsed != 0.5 {
		t.Fatalf("half-second timer = (%v, %v), want (0.5, false)", elapsed, choose)
	}

	elapsed, choose = advanceSplashDroneDirectionTimer(elapsed, 0.5)
	if !choose || elapsed != 0 {
		t.Fatalf("one-second timer = (%v, %v), want (0, true)", elapsed, choose)
	}
}

func TestRandomSplashDroneVelocityUsesFullHorizontalSpeed(t *testing.T) {
	t.Parallel()

	velocity := randomSplashDroneVelocity(rand.New(rand.NewSource(1)))
	speed := math.Sqrt(float64(velocity.X*velocity.X + velocity.Z*velocity.Z))

	if math.Abs(speed-float64(droneTopSpeed)) > 0.00001 {
		t.Fatalf("random splash drone speed = %v, want %v", speed, droneTopSpeed)
	}
	if velocity.Y != 0 {
		t.Fatalf("random splash drone Y velocity = %v, want 0", velocity.Y)
	}
}

func TestClampSplashDroneVelocityToCenter(t *testing.T) {
	t.Parallel()

	center := rl.NewVector3(6, 0, 6)
	tests := []struct {
		name     string
		position Position3
		velocity Velocity3
		want     Velocity3
	}{
		{
			name:     "blocks positive X beyond limit",
			position: Position3{X: 11, Z: 6},
			velocity: Velocity3{X: 2, Z: 1},
			want:     Velocity3{Z: 1},
		},
		{
			name:     "blocks negative X beyond limit",
			position: Position3{X: 1, Z: 6},
			velocity: Velocity3{X: -2},
			want:     Velocity3{},
		},
		{
			name:     "blocks positive Z beyond limit",
			position: Position3{X: 6, Z: 11},
			velocity: Velocity3{X: 1, Z: 2},
			want:     Velocity3{X: 1},
		},
		{
			name:     "allows inward movement",
			position: Position3{X: 11, Z: 6},
			velocity: Velocity3{X: -2},
			want:     Velocity3{X: -2},
		},
		{
			name:     "allows movement to exact limit",
			position: Position3{X: 11, Z: 6},
			velocity: Velocity3{X: 1},
			want:     Velocity3{X: 1},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := clampSplashDroneVelocityToCenter(
				test.position,
				test.velocity,
				1,
				center,
				splashDroneMaximumMoveDistance,
			)
			if got != test.want {
				t.Fatalf("clamped velocity = %+v, want %+v", got, test.want)
			}
		})
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
