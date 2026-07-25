package main

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

func TestArtifactFragmentDropOffMovesAtConstantSpeed(t *testing.T) {
	t.Parallel()

	position := Position3{X: 0, Y: 0, Z: 0}
	target := rl.NewVector3(3, 0, 4)
	movement := &MovementAnimationComponent{
		startPosition:  rl.Vector3(position),
		targetPosition: target,
		duration:       artifactFragmentDropOffDuration(rl.Vector3(position), target),
		easing:         MovementAnimationLinear,
	}

	if complete := updateMovementAnimation(movement, &position, target, 0.5); complete {
		t.Fatal("drop-off completed before reaching its target")
	}
	if got, want := rl.Vector3Distance(rl.Vector3{}, rl.Vector3(position)), float32(2); got != want {
		t.Fatalf("distance moved = %v, want %v", got, want)
	}

	if complete := updateMovementAnimation(movement, &position, target, 1); !complete {
		t.Fatal("drop-off did not complete after reaching its target")
	}
	assertVector3Close(t, rl.Vector3(position), target)
}

func TestArtifactFragmentDropOffSystemEjectsCollectedFragmentsInOrder(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	droneMapper := ecs.NewMap2[Position3, Drone](world)
	droneEntity := droneMapper.NewEntity(
		&Position3{X: 1.5, Y: 3, Z: 6.5},
		&Drone{},
	)
	chargingPadMapper := ecs.NewMap2[Position3, ChargingPad](world)
	chargingPadMapper.NewEntity(
		&Position3{X: 1.5, Y: 1, Z: 6.5},
		&ChargingPad{},
	)

	manager := NewArtifactManager()
	manager.fragments[2] = &ArtifactFragment{ID: 2, Collected: true}
	manager.fragments[1] = &ArtifactFragment{ID: 1, Collected: true}
	manager.fragments[3] = &ArtifactFragment{ID: 3}

	modelCount := 0
	system := &ArtifactFragmentDropOffSystem{
		newModel: func(*ArtifactFragment) *rl.Model {
			modelCount++
			return &rl.Model{}
		},
		unloadModel: func(rl.Model) {},
	}
	game := &Game{
		world:           world,
		artifactManager: manager,
	}
	system.Initialize(game)

	system.Update(game)
	if !manager.fragments[1].DroppedOff {
		t.Fatal("oldest collected fragment was not ejected immediately")
	}
	if manager.fragments[2].DroppedOff {
		t.Fatal("second fragment was ejected without the configured delay")
	}

	dronePosition, _ := droneMapper.Get(droneEntity)
	dronePosition.X = 7
	game.FrameTime = artifactFragmentDropOffDelay - 0.01
	system.Update(game)
	if manager.fragments[2].DroppedOff {
		t.Fatal("second fragment was ejected before the configured delay")
	}

	game.FrameTime = 0.01
	system.Update(game)
	if manager.fragments[2].DroppedOff {
		t.Fatal("second fragment ejected while the drone was away from the charging pad")
	}

	dronePosition.X = 1.5
	game.FrameTime = 0
	system.Update(game)
	if !manager.fragments[2].DroppedOff {
		t.Fatal("oldest available fragment did not eject after the drone returned")
	}
	if manager.fragments[3].DroppedOff {
		t.Fatal("uncollected fragment was ejected")
	}
	if got, want := modelCount, 2; got != want {
		t.Fatalf("created model count = %d, want %d", got, want)
	}
	if game.TotalScore != 0 {
		t.Fatalf("total score after ejection = %d, want 0 before pad arrival", game.TotalScore)
	}

	dropOffs := artifactFragmentDropOffs(system)
	if got, want := len(dropOffs), 2; got != want {
		t.Fatalf("drop-off entity count = %d, want %d", got, want)
	}
	if got, want := dropOffs[0].component.fragment.ID, int32(1); got != want {
		t.Fatalf("first drop-off fragment id = %d, want %d", got, want)
	}
	if got, want := dropOffs[1].component.fragment.ID, int32(2); got != want {
		t.Fatalf("second drop-off fragment id = %d, want %d", got, want)
	}
	if got, want := dropOffs[0].renderable.scale, float32(1); got != want {
		t.Fatalf("first drop-off scale = %v, want %v", got, want)
	}
	if got, want := rl.Vector3(dropOffs[1].position), rl.NewVector3(1.5, 3, 6.5); got != want {
		t.Fatalf("second drop-off started at %v, want %v", got, want)
	}
}

func TestArtifactFragmentDropOffSystemCompletesAndUnloadsFlights(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	drone := ecs.NewMap2[Position3, Drone](world).NewEntity(&Position3{}, &Drone{})
	mapper := ecs.NewMap4[Position3, Renderable, ArtifactFragmentDropOffComponent, MovementAnimationComponent](world)
	first := mapper.NewEntity(
		&Position3{},
		&Renderable{model: &rl.Model{}, scale: 1},
		&ArtifactFragmentDropOffComponent{
			fragment: &ArtifactFragment{ID: 1, Score: 17, DroppedOff: true},
		},
		&MovementAnimationComponent{
			targetPosition: rl.NewVector3(1, 0, 0),
			duration:       0.25,
		},
	)
	second := mapper.NewEntity(
		&Position3{X: 10},
		&Renderable{model: &rl.Model{}, scale: 1},
		&ArtifactFragmentDropOffComponent{
			fragment: &ArtifactFragment{ID: 2, Score: 23, DroppedOff: true},
		},
		&MovementAnimationComponent{
			startPosition:  rl.NewVector3(10, 0, 0),
			targetPosition: rl.NewVector3(20, 0, 0),
			duration:       2.5,
		},
	)

	unloadCount := 0
	system := &ArtifactFragmentDropOffSystem{
		unloadModel: func(rl.Model) {
			unloadCount++
		},
	}
	game := &Game{world: world, TotalScore: 5, FrameTime: 0.25}
	system.Initialize(game)
	movementSystem := &MovementAnimationSystem{}
	movementSystem.Initialize(game)
	movementSystem.Update(game)
	system.completeFlights(game)

	if world.Alive(first) {
		t.Fatal("completed drop-off entity is still alive")
	}
	if !world.Alive(second) {
		t.Fatal("in-flight drop-off entity was removed too early")
	}
	if got, want := unloadCount, 1; got != want {
		t.Fatalf("model unload count after arrival = %d, want %d", got, want)
	}
	if got, want := game.TotalScore, 22; got != want {
		t.Fatalf("total score after first arrival = %d, want %d", got, want)
	}
	if got, want := soundPlaybackRequestCount(world), 1; got != want {
		t.Fatalf("score sound request count after first arrival = %d, want %d", got, want)
	}
	recharge := ecs.NewMap[BatteryRecharge](world).Get(drone)
	if recharge == nil {
		t.Fatal("battery recharge was not added to the drone")
	}
	if got, want := recharge.Charge, float32(17)*batteryScoreModifier; got != want {
		t.Fatalf("battery recharge after first arrival = %v, want %v", got, want)
	}

	ecs.NewMap3[Position3, Renderable, ArtifactFragmentDropOffComponent](world).NewEntity(
		&Position3{},
		&Renderable{model: &rl.Model{}, scale: 1},
		&ArtifactFragmentDropOffComponent{
			fragment: &ArtifactFragment{ID: 3, Score: 100, DroppedOff: true},
		},
	)
	system.completeFlights(game)
	if got, want := recharge.Charge, float32(117)*batteryScoreModifier; got != want {
		t.Fatalf("accumulated battery recharge = %v, want %v", got, want)
	}
	if got, want := game.TotalScore, 122; got != want {
		t.Fatalf("total score after second arrival = %d, want %d", got, want)
	}
	if got, want := soundPlaybackRequestCount(world), 2; got != want {
		t.Fatalf("score sound request count after second arrival = %d, want %d", got, want)
	}

	system.Unload()
	if world.Alive(second) {
		t.Fatal("in-flight drop-off entity is still alive after unload")
	}
	if got, want := unloadCount, 3; got != want {
		t.Fatalf("model unload count after system unload = %d, want %d", got, want)
	}
	if got, want := game.TotalScore, 122; got != want {
		t.Fatalf("total score after unloading in-flight fragment = %d, want %d", got, want)
	}
}

type artifactFragmentDropOffSnapshot struct {
	position   Position3
	renderable Renderable
	component  ArtifactFragmentDropOffComponent
}

func artifactFragmentDropOffs(system *ArtifactFragmentDropOffSystem) []artifactFragmentDropOffSnapshot {
	query := system.filter.Query()
	defer query.Close()

	dropOffs := make([]artifactFragmentDropOffSnapshot, 0)
	for query.Next() {
		position, renderable, component := query.Get()
		dropOffs = append(dropOffs, artifactFragmentDropOffSnapshot{
			position:   *position,
			renderable: *renderable,
			component:  *component,
		})
	}
	return dropOffs
}
