package main

import (
	"image"
	"math"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

func TestArtifactFragmentPickupPlaneDimensionsPreserveWorldAspectRatio(t *testing.T) {
	t.Parallel()

	fragment := &ArtifactFragment{Image: image.NewRGBA(image.Rect(0, 0, 128, 64))}
	width, length := artifactFragmentPlaneDimensions(fragment)

	if width != 2 {
		t.Fatalf("plane width = %v, want 2", width)
	}
	if length != 1 {
		t.Fatalf("plane length = %v, want 1", length)
	}
}

func TestArtifactFragmentPickupPositionUsesRegionCenterAndTerrainHeight(t *testing.T) {
	t.Parallel()

	chunk := newTestTerrainChunk(ChunkCoords{X: 1, Z: -1}, 8, -8, 2)
	position := artifactFragmentPickupPosition(chunk, image.Rect(64, 128, 192, 256))

	assertVector3(t, position, 10, 2+artifactFragmentPickupGroundLift, -5)
}

func TestUpdateArtifactFragmentPickupPresentationTiltsAndShrinks(t *testing.T) {
	t.Parallel()

	movement := &MovementAnimationComponent{
		duration: artifactFragmentPickupHomingDuration,
		elapsed:  artifactFragmentPickupHomingDuration / 2,
		easing:   MovementAnimationEaseInCubic,
	}
	model := &rl.Model{}
	position := Position3{X: 5, Y: 4, Z: 3}
	renderable := &Renderable{model: model, scale: 1}
	camera := rl.NewCamera3D(
		rl.NewVector3(20, 20, 20),
		rl.Vector3{},
		rl.NewVector3(0, 1, 0),
		45,
		rl.CameraPerspective,
	)

	updateArtifactFragmentPickupPresentation(movement, &position, renderable, camera)

	if got, want := renderable.scale, float32(0.875); got != want {
		t.Fatalf("scale midway through homing = %v, want %v", got, want)
	}
	if model.Transform == rl.MatrixIdentity() {
		t.Fatal("model did not tilt during homing")
	}
}

func TestArtifactFragmentPickupSystemStartsOneNearestReadyFragment(t *testing.T) {
	t.Parallel()

	world, game, system := newArtifactFragmentPickupTest(t)
	farther := addTestWorldFragment(world, &ArtifactFragment{ID: 1, Weight: 10}, rl.NewVector3(0.4, 1, 0))
	tieHigherID := addTestWorldFragment(world, &ArtifactFragment{ID: 3, Weight: 10}, rl.NewVector3(0.2, 1, 0))
	tieLowerID := addTestWorldFragment(world, &ArtifactFragment{ID: 2, Weight: 10}, rl.NewVector3(-0.2, 1, 0))

	system.Update(game)

	if system.pickupMap.Has(farther) {
		t.Fatal("farther fragment started homing")
	}
	if system.pickupMap.Has(tieHigherID) {
		t.Fatal("higher-ID equidistant fragment started homing")
	}
	if !system.pickupMap.Has(tieLowerID) {
		t.Fatal("nearest fragment with lowest tie-breaking ID did not start homing")
	}
}

func TestArtifactFragmentPickupSystemTracksMovingDrone(t *testing.T) {
	t.Parallel()

	world, game, pickupSystem := newArtifactFragmentPickupTest(t)
	start := rl.NewVector3(0.1, 1, 0)
	fragmentEntity := addTestWorldFragment(world, &ArtifactFragment{ID: 1, Weight: 10}, start)
	droneEntity, _, ok := pickupSystem.drone()
	if !ok {
		t.Fatal("test drone was not found")
	}

	pickupSystem.Update(game)
	movementSystem := &MovementAnimationSystem{}
	movementSystem.Initialize(game)
	dronePosition := ecs.NewMap[Position3](world).Get(droneEntity)
	*dronePosition = Position3{X: 0.4}
	game.FrameTime = artifactFragmentPickupHomingDuration / 2
	movementSystem.Update(game)

	fragmentPosition := ecs.NewMap[Position3](world).Get(fragmentEntity)
	want := rl.Vector3Lerp(start, rl.NewVector3(0.4, 0, 0), movementAnimationEasedProgress(0.5, MovementAnimationEaseInCubic))
	assertVector3Close(t, rl.Vector3(*fragmentPosition), want)

	*dronePosition = Position3{X: 0.5}
	game.FrameTime = 0
	movementSystem.Update(game)
	fragmentPosition = ecs.NewMap[Position3](world).Get(fragmentEntity)
	want = rl.Vector3Lerp(start, rl.NewVector3(0.5, 0, 0), movementAnimationEasedProgress(0.5, MovementAnimationEaseInCubic))
	assertVector3Close(t, rl.Vector3(*fragmentPosition), want)
}

func TestArtifactFragmentPickupSystemWaitsForRiseRangeAndCapacity(t *testing.T) {
	t.Parallel()

	world, game, system := newArtifactFragmentPickupTest(t)
	game.artifactManager.fragments[100] = &ArtifactFragment{
		ID:        100,
		Weight:    droneMaximumCarryWeight - 100,
		Collected: true,
	}
	exactFit := addTestWorldFragment(world, &ArtifactFragment{ID: 1, Weight: 100}, rl.NewVector3(0.25, 1, 0))
	tooHeavy := addTestWorldFragment(world, &ArtifactFragment{ID: 2, Weight: 101}, rl.NewVector3(0.1, 1, 0))
	outOfRange := addTestWorldFragment(world, &ArtifactFragment{ID: 3, Weight: 1}, rl.NewVector3(1, 1, 0))
	rising := addTestWorldFragment(world, &ArtifactFragment{ID: 4, Weight: 1}, rl.NewVector3(0.05, 1, 0))
	ecs.NewMap[MovementAnimationComponent](world).Add(rising, &MovementAnimationComponent{duration: 1})

	system.Update(game)

	if !system.pickupMap.Has(exactFit) {
		t.Fatal("exact-fit fragment did not start homing")
	}
	for entity, name := range map[ecs.Entity]string{
		tooHeavy:   "too-heavy",
		outOfRange: "out-of-range",
		rising:     "rising",
	} {
		if system.pickupMap.Has(entity) {
			t.Fatalf("%s fragment started homing", name)
		}
	}
}

func TestArtifactFragmentPickupSystemReservesHomingWeight(t *testing.T) {
	t.Parallel()

	world, game, system := newArtifactFragmentPickupTest(t)
	active := addTestWorldFragment(world, &ArtifactFragment{ID: 1, Weight: droneMaximumCarryWeight}, rl.NewVector3(0.1, 1, 0))
	system.pickupMoverMap.Add(
		active,
		&ArtifactFragmentPickupComponent{},
		&MovementAnimationComponent{
			startPosition: rl.NewVector3(0.1, 1, 0),
			duration:      artifactFragmentPickupHomingDuration,
		},
	)
	waiting := addTestWorldFragment(world, &ArtifactFragment{ID: 2, Weight: 1}, rl.NewVector3(0.2, 1, 0))

	system.Update(game)

	if system.pickupMap.Has(waiting) {
		t.Fatal("waiting fragment ignored weight reserved by homing fragment")
	}
}

func TestArtifactFragmentPickupSystemRetriesAfterDropOffFreesCapacity(t *testing.T) {
	t.Parallel()

	world, game, system := newArtifactFragmentPickupTest(t)
	carried := &ArtifactFragment{ID: 1, Weight: droneMaximumCarryWeight, Collected: true}
	game.artifactManager.fragments[carried.ID] = carried
	waiting := addTestWorldFragment(world, &ArtifactFragment{ID: 2, Weight: 1}, rl.NewVector3(0.2, 1, 0))

	system.Update(game)
	if system.pickupMap.Has(waiting) {
		t.Fatal("fragment started while drone was full")
	}

	carried.DroppedOff = true
	system.Update(game)
	if !system.pickupMap.Has(waiting) {
		t.Fatal("fragment did not start after drop-off freed capacity")
	}
}

func TestArtifactFragmentPickupSystemCollectsWithoutScoringAndRemovesEntity(t *testing.T) {
	t.Parallel()

	world, game, system := newArtifactFragmentPickupTest(t)
	game.TotalScore = 5
	fragment := &ArtifactFragment{ID: 1, Weight: 10, Score: 17}
	entity := addTestWorldFragment(world, fragment, rl.NewVector3(0.1, 1, 0))

	unloadCount := 0
	system.unloadModel = func(rl.Model) {
		unloadCount++
	}
	system.Update(game)
	movementSystem := &MovementAnimationSystem{}
	movementSystem.Initialize(game)
	game.FrameTime = artifactFragmentPickupHomingDuration
	movementSystem.Update(game)
	system.Update(game)

	if world.Alive(entity) {
		t.Fatal("pickup entity is still alive after completion")
	}
	if !fragment.Collected {
		t.Fatal("fragment was not marked collected")
	}
	if got, want := game.TotalScore, 5; got != want {
		t.Fatalf("total score = %d, want %d", got, want)
	}
	if got, want := unloadCount, 1; got != want {
		t.Fatalf("model unload count = %d, want %d", got, want)
	}

	collectArtifactFragment(fragment)
	if got, want := game.TotalScore, 5; got != want {
		t.Fatalf("total score after duplicate collection = %d, want %d", got, want)
	}
}

func TestArtifactFragmentPickupSystemUnloadReleasesAllWorldFragmentModels(t *testing.T) {
	t.Parallel()

	world, _, system := newArtifactFragmentPickupTest(t)
	ready := addTestWorldFragment(world, &ArtifactFragment{ID: 1}, rl.Vector3{})
	rising := addTestWorldFragment(world, &ArtifactFragment{ID: 2}, rl.Vector3{})
	homing := addTestWorldFragment(world, &ArtifactFragment{ID: 3}, rl.Vector3{})
	ecs.NewMap[MovementAnimationComponent](world).Add(rising, &MovementAnimationComponent{duration: 1})
	system.pickupMoverMap.Add(
		homing,
		&ArtifactFragmentPickupComponent{},
		&MovementAnimationComponent{duration: 1},
	)

	unloadCount := 0
	system.unloadModel = func(rl.Model) {
		unloadCount++
	}
	system.Unload()

	for _, entity := range []ecs.Entity{ready, rising, homing} {
		if world.Alive(entity) {
			t.Fatal("fragment entity is still alive after system unload")
		}
	}
	if got, want := unloadCount, 3; got != want {
		t.Fatalf("model unload count = %d, want %d", got, want)
	}
}

func newArtifactFragmentPickupTest(t *testing.T) (*ecs.World, *Game, *ArtifactFragmentPickupSystem) {
	t.Helper()

	world := ecs.NewWorld()
	ecs.NewMap2[Position3, Drone](world).NewEntity(&Position3{}, &Drone{})
	game := &Game{
		world:           world,
		artifactManager: NewArtifactManager(),
		camera: rl.NewCamera3D(
			rl.NewVector3(10, 10, 10),
			rl.Vector3{},
			rl.NewVector3(0, 1, 0),
			45,
			rl.CameraPerspective,
		),
	}
	system := &ArtifactFragmentPickupSystem{
		unloadModel: func(rl.Model) {},
	}
	system.Initialize(game)
	return world, game, system
}

func addTestWorldFragment(world *ecs.World, fragment *ArtifactFragment, position rl.Vector3) ecs.Entity {
	return ecs.NewMap3[Position3, Renderable, ArtifactFragmentComponent](world).NewEntity(
		(*Position3)(&position),
		&Renderable{model: &rl.Model{}, scale: 1},
		&ArtifactFragmentComponent{fragment: fragment},
	)
}

func assertVector3Close(t *testing.T, got, want rl.Vector3) {
	t.Helper()

	const tolerance = 0.0001
	if math.Abs(float64(got.X-want.X)) > tolerance ||
		math.Abs(float64(got.Y-want.Y)) > tolerance ||
		math.Abs(float64(got.Z-want.Z)) > tolerance {
		t.Fatalf("vector = (%v, %v, %v), want (%v, %v, %v)", got.X, got.Y, got.Z, want.X, want.Y, want.Z)
	}
}
