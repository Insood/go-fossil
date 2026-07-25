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

func TestUpdateArtifactFragmentPickupTracksDrone(t *testing.T) {
	t.Parallel()

	start := rl.NewVector3(1, 2, 3)
	pickup := &ArtifactFragmentPickupComponent{startPosition: start}
	model := &rl.Model{}
	position := Position3(start)
	renderable := &Renderable{model: model, scale: 1}
	camera := rl.NewCamera3D(
		rl.NewVector3(20, 20, 20),
		rl.Vector3{},
		rl.NewVector3(0, 1, 0),
		45,
		rl.CameraPerspective,
	)
	dronePosition := rl.NewVector3(9, 6, 7)

	if complete := updateArtifactFragmentPickup(pickup, &position, renderable, dronePosition, camera, 0); complete {
		t.Fatal("pickup completed at start")
	}
	assertVector3Close(t, rl.Vector3(position), start)

	if complete := updateArtifactFragmentPickup(pickup, &position, renderable, dronePosition, camera, artifactFragmentPickupHomingDuration/2); complete {
		t.Fatal("pickup completed midway through homing")
	}
	wantMidpoint := rl.Vector3Lerp(start, dronePosition, easeInCubic(0.5))
	assertVector3Close(t, rl.Vector3(position), wantMidpoint)
	if renderable.scale >= 1 || renderable.scale <= 0 {
		t.Fatalf("scale midway through homing = %v, want between 0 and 1", renderable.scale)
	}
	if model.Transform == rl.MatrixIdentity() {
		t.Fatal("model did not tilt during homing")
	}

	movedDronePosition := rl.NewVector3(13, 10, 11)
	updateArtifactFragmentPickup(pickup, &position, renderable, movedDronePosition, camera, 0)
	wantTrackedPosition := rl.Vector3Lerp(start, movedDronePosition, easeInCubic(0.5))
	assertVector3Close(t, rl.Vector3(position), wantTrackedPosition)

	if complete := updateArtifactFragmentPickup(pickup, &position, renderable, movedDronePosition, camera, artifactFragmentPickupHomingDuration/2+0.001); !complete {
		t.Fatal("pickup did not complete")
	}
	assertVector3Close(t, rl.Vector3(position), movedDronePosition)
	if renderable.scale != 0 {
		t.Fatalf("scale at completion = %v, want 0", renderable.scale)
	}
}

func TestArtifactFragmentPickupSystemStartsOneNearestReadyFragment(t *testing.T) {
	t.Parallel()

	world, game, system := newArtifactFragmentPickupTest(t)
	farther := addTestWorldFragment(world, &ArtifactFragment{ID: 1, Weight: 10}, rl.NewVector3(0.4, 1, 0))
	tieHigherID := addTestWorldFragment(world, &ArtifactFragment{ID: 3, Weight: 10}, rl.NewVector3(0.2, 1, 0))
	tieLowerID := addTestWorldFragment(world, &ArtifactFragment{ID: 2, Weight: 10}, rl.NewVector3(-0.2, 1, 0))

	system.update(game, 0)

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
	ecs.NewMap[ArtifactFragmentRiseComponent](world).Add(rising, &ArtifactFragmentRiseComponent{})

	system.update(game, 0)

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
	system.pickupMap.Add(active, &ArtifactFragmentPickupComponent{startPosition: rl.NewVector3(0.1, 1, 0)})
	waiting := addTestWorldFragment(world, &ArtifactFragment{ID: 2, Weight: 1}, rl.NewVector3(0.2, 1, 0))

	system.update(game, 0)

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

	system.update(game, 0)
	if system.pickupMap.Has(waiting) {
		t.Fatal("fragment started while drone was full")
	}

	carried.DroppedOff = true
	system.update(game, 0)
	if !system.pickupMap.Has(waiting) {
		t.Fatal("fragment did not start after drop-off freed capacity")
	}
}

func TestArtifactFragmentPickupSystemCollectsScoresAndRemovesEntity(t *testing.T) {
	t.Parallel()

	world, game, system := newArtifactFragmentPickupTest(t)
	game.TotalScore = 5
	fragment := &ArtifactFragment{ID: 1, Weight: 10, Score: 17}
	entity := addTestWorldFragment(world, fragment, rl.NewVector3(0.1, 1, 0))

	unloadCount := 0
	system.unloadModel = func(rl.Model) {
		unloadCount++
	}
	system.update(game, artifactFragmentPickupHomingDuration)

	if world.Alive(entity) {
		t.Fatal("pickup entity is still alive after completion")
	}
	if !fragment.Collected {
		t.Fatal("fragment was not marked collected")
	}
	if got, want := game.TotalScore, 22; got != want {
		t.Fatalf("total score = %d, want %d", got, want)
	}
	if got, want := unloadCount, 1; got != want {
		t.Fatalf("model unload count = %d, want %d", got, want)
	}

	collectArtifactFragment(game, fragment)
	if got, want := game.TotalScore, 22; got != want {
		t.Fatalf("total score after duplicate collection = %d, want %d", got, want)
	}
}

func TestArtifactFragmentPickupSystemUnloadReleasesAllWorldFragmentModels(t *testing.T) {
	t.Parallel()

	world, _, system := newArtifactFragmentPickupTest(t)
	ready := addTestWorldFragment(world, &ArtifactFragment{ID: 1}, rl.Vector3{})
	rising := addTestWorldFragment(world, &ArtifactFragment{ID: 2}, rl.Vector3{})
	homing := addTestWorldFragment(world, &ArtifactFragment{ID: 3}, rl.Vector3{})
	ecs.NewMap[ArtifactFragmentRiseComponent](world).Add(rising, &ArtifactFragmentRiseComponent{})
	system.pickupMap.Add(homing, &ArtifactFragmentPickupComponent{})

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
