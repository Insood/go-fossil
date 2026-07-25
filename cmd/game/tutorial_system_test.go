package main

import (
	"image"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

func TestTutorialSystemStartsAtMoveDroneStep(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	mapper := ecs.NewMap2[Position3, Drone](world)
	mapper.NewEntity(
		&Position3{X: 1, Y: 2, Z: 3},
		&Drone{},
	)

	game := &Game{world: world}
	system := &TutorialSystem{}
	system.Initialize(game)

	if got, want := system.state.currentStep, tutorialStepMoveDrone; got != want {
		t.Fatalf("current step = %d, want %d", got, want)
	}
}

func TestTutorialStepOneIgnoresDroneYChanges(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	mapper := ecs.NewMap2[Position3, Drone](world)
	entity := mapper.NewEntity(
		&Position3{X: 1, Y: 2, Z: 3},
		&Drone{},
	)

	game := &Game{world: world}
	system := &TutorialSystem{}
	system.Initialize(game)

	position, _ := mapper.Get(entity)
	position.Y += 1
	system.updateCurrentStep(game)

	if got, want := system.state.currentStep, tutorialStepMoveDrone; got != want {
		t.Fatalf("current step = %d, want %d", got, want)
	}
}

func TestTutorialStepOneStartsFindArtifactStepWhenDroneMovesOnXZ(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	mapper := ecs.NewMap2[Position3, Drone](world)
	entity := mapper.NewEntity(
		&Position3{X: 1, Y: 2, Z: 3},
		&Drone{},
	)

	game := &Game{world: world}
	system := &TutorialSystem{}
	system.Initialize(game)

	position, _ := mapper.Get(entity)
	position.X += 0.001
	system.updateCurrentStep(game)

	if got, want := system.state.currentStep, tutorialStepFindArtifact; got != want {
		t.Fatalf("current step = %d, want %d", got, want)
	}
}

func TestTutorialStepTwoSpawnsMarkersAtArtifactCenters(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	droneMapper := ecs.NewMap2[Position3, Drone](world)
	droneEntity := droneMapper.NewEntity(
		&Position3{X: 1, Y: 2, Z: 3},
		&Drone{},
	)

	artifactManager := NewArtifactManager()
	chunk := newTestTerrainChunk(ChunkCoords{X: 1, Z: 0}, 8, 0, 2)
	artifactManager.RegisterChunkArtifact(chunk, "phone", 10, 20, 128, 192, image.Rectangle{})

	model := &rl.Model{}
	game := &Game{
		world:           world,
		artifactManager: artifactManager,
		assets: &AssetManager{
			models: map[string]*rl.Model{
				tutorialArtifactMarkerModelName: model,
			},
		},
	}
	system := &TutorialSystem{}
	system.Initialize(game)

	position, _ := droneMapper.Get(droneEntity)
	position.X += 1
	system.updateCurrentStep(game)

	markers := tutorialMarkers(system)
	if len(markers) != 1 {
		t.Fatalf("marker count = %d, want 1", len(markers))
	}

	marker := markers[0]
	assertVector3(t, marker.position, 10, 2+tutorialArtifactMarkerLift, 3)
	if marker.renderable.model != model {
		t.Fatal("marker renderable does not use tutorial cone model")
	}
	if marker.renderable.tint != rl.Red {
		t.Fatalf("marker tint = %#v, want red", marker.renderable.tint)
	}
}

func TestTutorialStepTwoStartsMoveLaserStepAndRemovesMarkersWhenDroneReachesMarker(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	droneMapper := ecs.NewMap2[Position3, Drone](world)
	droneEntity := droneMapper.NewEntity(
		&Position3{X: 1, Y: 2, Z: 3},
		&Drone{},
	)

	system := &TutorialSystem{}
	game := &Game{world: world}
	system.Initialize(game)
	system.state.currentStep = tutorialStepFindArtifact
	system.markersSpawned = true
	markerEntity := system.markerMapper.NewEntity(
		&Position3{X: 2, Y: 1, Z: 3},
		&Renderable{},
		&TutorialMarker{},
	)

	position, _ := droneMapper.Get(droneEntity)
	position.X = 1.6
	system.updateCurrentStep(game)

	if got, want := system.state.currentStep, tutorialStepMoveLaser; got != want {
		t.Fatalf("current step = %d, want %d", got, want)
	}
	if world.Alive(markerEntity) {
		t.Fatal("tutorial marker entity is still alive")
	}
}

func TestTutorialStepThreeWaitsForCursorMovement(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	playerFireInputMapper := ecs.NewMap1[PlayerFireInput](world)
	playerFireInputEntity := playerFireInputMapper.NewEntity(&PlayerFireInput{
		cursor: rl.NewVector2(0, 0),
	})

	system := &TutorialSystem{}
	game := &Game{world: world}
	system.Initialize(game)
	system.state.currentStep = tutorialStepMoveLaser

	system.updateCurrentStep(game)
	control := playerFireInputMapper.Get(playerFireInputEntity)
	control.cursor = rl.NewVector2(0, 0)
	system.updateCurrentStep(game)

	control.cursor = rl.NewVector2(tutorialLaserMoveThresholdNormalized-0.01, 0)
	system.updateCurrentStep(game)
	if got, want := system.state.currentStep, tutorialStepMoveLaser; got != want {
		t.Fatalf("current step = %d, want %d before enough laser movement", got, want)
	}
}

func TestTutorialStepThreeStartsFireLaserStepAfterCursorMovesThreshold(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	playerFireInputMapper := ecs.NewMap1[PlayerFireInput](world)
	playerFireInputEntity := playerFireInputMapper.NewEntity(&PlayerFireInput{
		cursor: rl.NewVector2(0, 0),
	})

	system := &TutorialSystem{}
	game := &Game{world: world}
	system.Initialize(game)
	system.state.currentStep = tutorialStepMoveLaser

	system.updateCurrentStep(game)

	control := playerFireInputMapper.Get(playerFireInputEntity)
	control.cursor = rl.NewVector2(tutorialLaserMoveThresholdNormalized, 0)
	system.updateCurrentStep(game)

	if got, want := system.state.currentStep, tutorialStepFireLaser; got != want {
		t.Fatalf("current step = %d, want %d", got, want)
	}
}

func TestTutorialStepFourStartsCutOutFossilStepWhenLaserIsActive(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	laserMapper := ecs.NewMap1[Laser](world)
	laserEntity := laserMapper.NewEntity(&Laser{})

	system := &TutorialSystem{}
	game := &Game{world: world}
	system.Initialize(game)
	system.state.currentStep = tutorialStepFireLaser

	system.updateCurrentStep(game)
	if got, want := system.state.currentStep, tutorialStepFireLaser; got != want {
		t.Fatalf("current step = %d, want %d before laser is active", got, want)
	}

	laserMapper.Get(laserEntity).active = true
	system.updateCurrentStep(game)
	if got, want := system.state.currentStep, tutorialStepCutOutFossil; got != want {
		t.Fatalf("current step = %d, want %d after laser is active", got, want)
	}
}

func TestTutorialStepFiveWaitsForCollectedArtifactFragmentThenStartsReturnHome(t *testing.T) {
	t.Parallel()

	system := &TutorialSystem{}
	game := &Game{
		world:           ecs.NewWorld(),
		artifactManager: NewArtifactManager(),
	}
	system.Initialize(game)
	system.state.currentStep = tutorialStepCutOutFossil

	system.updateCurrentStep(game)
	if got, want := system.state.currentStep, tutorialStepCutOutFossil; got != want {
		t.Fatalf("current step = %d, want %d before fragment exists", got, want)
	}

	game.artifactManager.fragments[1] = &ArtifactFragment{ID: 1}
	system.updateCurrentStep(game)
	if got, want := system.state.currentStep, tutorialStepCutOutFossil; got != want {
		t.Fatalf("current step = %d, want %d while fragment pickup is in flight", got, want)
	}

	game.artifactManager.fragments[1].Collected = true
	system.updateCurrentStep(game)
	if got, want := system.state.currentStep, tutorialStepReturnHome; got != want {
		t.Fatalf("current step = %d, want %d after fragment is collected", got, want)
	}
}

func TestTutorialStepSixMarksChargingPadAndAdvancesAtProximity(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	droneMapper := ecs.NewMap2[Position3, Drone](world)
	droneEntity := droneMapper.NewEntity(
		&Position3{X: 1, Y: 2, Z: 3},
		&Drone{},
	)
	chargingPadMapper := ecs.NewMap2[Position3, ChargingPad](world)
	chargingPadMapper.NewEntity(
		&Position3{X: 9.5, Y: 1, Z: -1.5},
		&ChargingPad{},
	)

	markerModel := &rl.Model{}
	game := &Game{
		world: world,
		assets: &AssetManager{
			models: map[string]*rl.Model{
				tutorialArtifactMarkerModelName: markerModel,
			},
		},
	}
	system := &TutorialSystem{}
	system.Initialize(game)
	system.state.currentStep = tutorialStepReturnHome
	system.stepElapsed = 3

	system.updateCurrentStep(game)

	markers := tutorialMarkers(system)
	if got, want := len(markers), 1; got != want {
		t.Fatalf("marker count = %d, want %d", got, want)
	}
	assertVector3(t, markers[0].position, 9.5, 1+tutorialArtifactMarkerLift, -1.5)
	if markers[0].renderable.model != markerModel {
		t.Fatal("marker renderable does not use tutorial cone model")
	}

	dronePosition, _ := droneMapper.Get(droneEntity)
	dronePosition.X = markers[0].position.X + tutorialArtifactMarkerProximity + 0.01
	dronePosition.Z = markers[0].position.Z
	system.updateCurrentStep(game)
	if got, want := system.state.currentStep, tutorialStepReturnHome; got != want {
		t.Fatalf("current step = %d, want %d outside charging pad proximity", got, want)
	}

	dronePosition.X = markers[0].position.X + tutorialArtifactMarkerProximity
	system.updateCurrentStep(game)
	if got, want := system.state.currentStep, tutorialStepCollectMore; got != want {
		t.Fatalf("current step = %d, want %d at charging pad proximity", got, want)
	}
	if system.stepElapsed != 0 {
		t.Fatalf("step elapsed = %v, want 0 after entering collect-more step", system.stepElapsed)
	}
	if got := len(tutorialMarkers(system)); got != 0 {
		t.Fatalf("marker count after return-home completion = %d, want 0", got)
	}
}

func TestTutorialStepSevenCompletesAfterFiveSeconds(t *testing.T) {
	t.Parallel()

	system := &TutorialSystem{
		state: TutorialState{currentStep: tutorialStepCollectMore},
	}

	system.updateTimedStep(tutorialCollectMoreDuration - 0.01)
	if got, want := system.state.currentStep, tutorialStepCollectMore; got != want {
		t.Fatalf("current step = %d, want %d before five seconds", got, want)
	}

	system.updateTimedStep(0.01)
	if got, want := system.state.currentStep, tutorialStepComplete; got != want {
		t.Fatalf("current step = %d, want %d after five seconds", got, want)
	}
}

func TestAnimationFrameAtStartsAndLoopsFromFirstFrame(t *testing.T) {
	t.Parallel()

	animation := &Animation{
		Frames: []AnimationFrame{
			{Texture: rl.Texture2D{ID: 1}, Duration: 0.1},
			{Texture: rl.Texture2D{ID: 2}, Duration: 0.2},
		},
	}

	frame, ok := animationFrameAt(animation, 0)
	if !ok {
		t.Fatal("animationFrameAt() ok = false, want true")
	}
	if got, want := frame.Texture.ID, uint32(1); got != want {
		t.Fatalf("frame at start = %d, want %d", got, want)
	}

	frame, ok = animationFrameAt(animation, 0.3)
	if !ok {
		t.Fatal("animationFrameAt() ok = false, want true")
	}
	if got, want := frame.Texture.ID, uint32(1); got != want {
		t.Fatalf("frame after loop = %d, want %d", got, want)
	}
}

type tutorialMarkerSnapshot struct {
	position   rl.Vector3
	renderable *Renderable
}

func tutorialMarkers(system *TutorialSystem) []tutorialMarkerSnapshot {
	query := system.markerFilter.Query()
	defer query.Close()

	markers := make([]tutorialMarkerSnapshot, 0)
	for query.Next() {
		position, _ := query.Get()
		_, renderable, _ := system.markerMapper.Get(query.Entity())
		markers = append(markers, tutorialMarkerSnapshot{
			position:   rl.Vector3(*position),
			renderable: renderable,
		})
	}

	return markers
}
