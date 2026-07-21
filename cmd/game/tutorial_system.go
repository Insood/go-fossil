package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type tutorialStep int

const (
	tutorialStepComplete tutorialStep = iota
	tutorialStepMoveDrone
	tutorialStepFindArtifact
)

type TutorialState struct {
	currentStep tutorialStep
}

func NewTutorialState() *TutorialState {
	return &TutorialState{currentStep: tutorialStepMoveDrone}
}

type TutorialSystem struct {
	droneFilter             *ecs.Filter2[Position3, Drone]
	markerMapper            *ecs.Map3[Position3, Renderable, TutorialMarker]
	markerFilter            *ecs.Filter2[Position3, TutorialMarker]
	state                   TutorialState
	initialDronePosition    rl.Vector3
	hasInitialDronePosition bool
	markersSpawned          bool
}

func (system *TutorialSystem) Initialize(game *Game) {
	system.droneFilter = ecs.NewFilter2[Position3, Drone](game.world)
	system.markerMapper = ecs.NewMap3[Position3, Renderable, TutorialMarker](game.world)
	system.markerFilter = ecs.NewFilter2[Position3, TutorialMarker](game.world)
	system.state = *NewTutorialState()
	system.captureInitialDronePosition()
}

func (system *TutorialSystem) Update(game *Game) {
	system.updateCurrentStep(game)
	system.drawCurrentStep(game)
}

func (system *TutorialSystem) updateCurrentStep(game *Game) {
	switch system.state.currentStep {
	case tutorialStepMoveDrone:
		system.updateMoveDroneStep(game)
	case tutorialStepFindArtifact:
		system.updateFindArtifactStep(game)
	}
}

func (system *TutorialSystem) updateMoveDroneStep(game *Game) {
	if !system.hasInitialDronePosition {
		system.captureInitialDronePosition()
	}

	position, ok := system.dronePosition()
	if !ok {
		return
	}
	if position.X == system.initialDronePosition.X && position.Z == system.initialDronePosition.Z {
		return
	}

	system.state.currentStep = tutorialStepFindArtifact
	system.spawnArtifactMarkers(game)
}

func (system *TutorialSystem) updateFindArtifactStep(game *Game) {
	if !system.markersSpawned {
		system.spawnArtifactMarkers(game)
	}
	if !system.droneIsNearMarker() {
		return
	}

	system.removeArtifactMarkers(game)
	system.state.currentStep = tutorialStepComplete
}

func (system *TutorialSystem) drawCurrentStep(game *Game) {
	switch system.state.currentStep {
	case tutorialStepMoveDrone:
		system.drawMoveDronePrompt(game)
	case tutorialStepFindArtifact:
		system.drawFindArtifactPrompt()
	}
}

func (system *TutorialSystem) drawMoveDronePrompt(game *Game) {
	if game.assets == nil {
		return
	}

	texture, ok := game.assets.LookupTexture(tutorialMoveDroneTextureName)
	if !ok || texture.ID == 0 {
		return
	}

	system.drawCenteredPromptText("Move using")
	system.drawCenteredTextureBelowPrompt(texture)
}

func (system *TutorialSystem) drawFindArtifactPrompt() {
	system.drawCenteredPromptText("Move here")
}

func (system *TutorialSystem) drawCenteredPromptText(text string) {
	textWidth := rl.MeasureText(text, tutorialPromptFontSize)
	textX := (screenWidth - textWidth) / 2
	rl.DrawText(text, textX, tutorialPromptTopY, tutorialPromptFontSize, rl.White)
}

func (system *TutorialSystem) drawCenteredTextureBelowPrompt(texture rl.Texture2D) {
	imageX := (screenWidth - texture.Width) / 2
	imageY := tutorialPromptTopY + tutorialPromptFontSize + tutorialPromptImageGap
	rl.DrawTexture(texture, imageX, imageY, rl.White)
}

func (system *TutorialSystem) spawnArtifactMarkers(game *Game) {
	system.markersSpawned = true
	if game.assets == nil {
		return
	}

	model, ok := game.assets.LookupModel(tutorialArtifactMarkerModelName)
	if !ok || model == nil {
		return
	}

	for _, artifact := range game.artifactManager.Artifacts() {
		position, ok := tutorialMarkerPosition(artifact)
		if !ok {
			continue
		}

		system.markerMapper.NewEntity(
			&Position3{X: position.X, Y: position.Y, Z: position.Z},
			&Renderable{
				model:          model,
				scale:          1,
				tint:           rl.Red,
				castsShadow:    true,
				receivesShadow: false,
			},
			&TutorialMarker{},
		)
	}
}

func tutorialMarkerPosition(artifact *Artifact) (rl.Vector3, bool) {
	if artifact == nil || artifact.Chunk == nil {
		return rl.Vector3{}, false
	}

	worldX := artifact.Chunk.OriginX + artifact.CenterX/float32(terrainTexturePixelsPerTile)
	worldZ := artifact.Chunk.OriginZ + artifact.CenterZ/float32(terrainTexturePixelsPerTile)
	worldY := artifact.Chunk.HeightAtWorldPosition(worldX, worldZ) + tutorialArtifactMarkerLift
	return rl.NewVector3(worldX, worldY, worldZ), true
}

func (system *TutorialSystem) droneIsNearMarker() bool {
	dronePosition, ok := system.dronePosition()
	if !ok {
		return false
	}

	query := system.markerFilter.Query()
	defer query.Close()

	for query.Next() {
		markerPosition, _ := query.Get()
		if rl.Vector2Distance(xzVector(rl.Vector3(*markerPosition)), xzVector(dronePosition)) <= tutorialArtifactMarkerProximity {
			return true
		}
	}

	return false
}

func xzVector(position rl.Vector3) rl.Vector2 {
	return rl.NewVector2(position.X, position.Z)
}

func (system *TutorialSystem) removeArtifactMarkers(game *Game) {
	query := system.markerFilter.Query()
	defer query.Close()

	entities := make([]ecs.Entity, 0)
	for query.Next() {
		entities = append(entities, query.Entity())
	}

	for _, entity := range entities {
		game.world.RemoveEntity(entity)
	}
}

func (system *TutorialSystem) captureInitialDronePosition() {
	position, ok := system.dronePosition()
	if !ok {
		return
	}

	system.initialDronePosition = position
	system.hasInitialDronePosition = true
}

func (system *TutorialSystem) dronePosition() (rl.Vector3, bool) {
	query := system.droneFilter.Query()
	defer query.Close()

	if !query.Next() {
		return rl.Vector3{}, false
	}

	position, _ := query.Get()
	return rl.Vector3(*position), true
}
