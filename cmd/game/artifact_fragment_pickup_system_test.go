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

func TestUpdateArtifactFragmentPickupRisesThenTracksDrone(t *testing.T) {
	t.Parallel()

	start := rl.NewVector3(1, 2, 3)
	raised := rl.NewVector3(1, 2+artifactFragmentPickupRiseHeight, 3)
	pickup := &ArtifactFragmentComponent{
		startPosition:  start,
		raisedPosition: raised,
	}
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

	if complete := updateArtifactFragmentPickup(
		pickup,
		&position,
		renderable,
		dronePosition,
		camera,
		artifactFragmentPickupRiseDuration,
	); complete {
		t.Fatal("pickup completed at rise boundary")
	}
	assertVector3Close(t, rl.Vector3(position), raised)
	if renderable.scale != 1 {
		t.Fatalf("scale at rise boundary = %v, want 1", renderable.scale)
	}

	if complete := updateArtifactFragmentPickup(
		pickup,
		&position,
		renderable,
		dronePosition,
		camera,
		artifactFragmentPickupHomingDuration/2,
	); complete {
		t.Fatal("pickup completed midway through homing")
	}
	wantMidpoint := rl.Vector3Lerp(raised, dronePosition, easeInCubic(0.5))
	assertVector3Close(t, rl.Vector3(position), wantMidpoint)
	if renderable.scale >= 1 || renderable.scale <= 0 {
		t.Fatalf("scale midway through homing = %v, want between 0 and 1", renderable.scale)
	}
	if model.Transform == rl.MatrixIdentity() {
		t.Fatal("model did not tilt during homing")
	}

	movedDronePosition := rl.NewVector3(13, 10, 11)
	updateArtifactFragmentPickup(pickup, &position, renderable, movedDronePosition, camera, 0)
	wantTrackedPosition := rl.Vector3Lerp(raised, movedDronePosition, easeInCubic(0.5))
	assertVector3Close(t, rl.Vector3(position), wantTrackedPosition)

	if complete := updateArtifactFragmentPickup(
		pickup,
		&position,
		renderable,
		movedDronePosition,
		camera,
		artifactFragmentPickupHomingDuration/2+0.001,
	); !complete {
		t.Fatal("pickup did not complete after one second")
	}
	assertVector3Close(t, rl.Vector3(position), movedDronePosition)
	if renderable.scale != 0 {
		t.Fatalf("scale at completion = %v, want 0", renderable.scale)
	}
}

func TestArtifactFragmentPickupSystemCollectsScoresAndRemovesEntity(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	droneMapper := ecs.NewMap2[Position3, Drone](world)
	droneMapper.NewEntity(&Position3{X: 5, Y: 4, Z: 3}, &Drone{})

	model := &rl.Model{}
	fragment := &ArtifactFragment{ID: 1, Score: 17}
	pickupMapper := ecs.NewMap3[Position3, Renderable, ArtifactFragmentComponent](world)
	entity := pickupMapper.NewEntity(
		&Position3{X: 1, Y: 1, Z: 1},
		&Renderable{model: model, scale: 1},
		&ArtifactFragmentComponent{
			fragment:       fragment,
			startPosition:  rl.NewVector3(1, 1, 1),
			raisedPosition: rl.NewVector3(1, 1+artifactFragmentPickupRiseHeight, 1),
		},
	)

	unloadCount := 0
	system := &ArtifactFragmentPickupSystem{
		unloadModel: func(rl.Model) {
			unloadCount++
		},
	}
	game := &Game{
		world: world,
		camera: rl.NewCamera3D(
			rl.NewVector3(10, 10, 10),
			rl.Vector3{},
			rl.NewVector3(0, 1, 0),
			45,
			rl.CameraPerspective,
		),
		TotalScore: 5,
	}
	system.Initialize(game)
	system.update(game, artifactFragmentPickupRiseDuration+artifactFragmentPickupHomingDuration)

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

func TestArtifactFragmentPickupSystemUnloadReleasesActiveModels(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	model := &rl.Model{}
	mapper := ecs.NewMap3[Position3, Renderable, ArtifactFragmentComponent](world)
	entity := mapper.NewEntity(
		&Position3{},
		&Renderable{model: model, scale: 1},
		&ArtifactFragmentComponent{fragment: &ArtifactFragment{ID: 1}},
	)

	unloadCount := 0
	system := &ArtifactFragmentPickupSystem{
		unloadModel: func(rl.Model) {
			unloadCount++
		},
	}
	game := &Game{world: world}
	system.Initialize(game)
	system.Unload()

	if world.Alive(entity) {
		t.Fatal("pickup entity is still alive after system unload")
	}
	if got, want := unloadCount, 1; got != want {
		t.Fatalf("model unload count = %d, want %d", got, want)
	}
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
