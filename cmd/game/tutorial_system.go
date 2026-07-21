package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type tutorialStep int

const (
	tutorialStepComplete tutorialStep = iota
	tutorialStepMoveDrone
)

type TutorialState struct {
	currentStep tutorialStep
}

func NewTutorialState() *TutorialState {
	return &TutorialState{currentStep: tutorialStepMoveDrone}
}

type TutorialSystem struct {
	droneFilter             *ecs.Filter2[Position3, Drone]
	state                   TutorialState
	initialDronePosition    rl.Vector3
	hasInitialDronePosition bool
}

func (system *TutorialSystem) Initialize(game *Game) {
	system.droneFilter = ecs.NewFilter2[Position3, Drone](game.world)
	system.state = *NewTutorialState()
	system.captureInitialDronePosition()
}

func (system *TutorialSystem) Update(game *Game) {
	system.updateCurrentStep()
	system.drawCurrentStep(game)
}

func (system *TutorialSystem) updateCurrentStep() {
	if system.state.currentStep != tutorialStepMoveDrone {
		return
	}
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

	system.state.currentStep = tutorialStepComplete
}

func (system *TutorialSystem) drawCurrentStep(game *Game) {
	if system.state.currentStep != tutorialStepMoveDrone {
		return
	}
	if game.assets == nil {
		return
	}

	texture, ok := game.assets.LookupTexture(tutorialMoveDroneTextureName)
	if !ok || texture.ID == 0 {
		return
	}

	system.drawMoveDroneStep(texture)
}

func (system *TutorialSystem) drawMoveDroneStep(texture rl.Texture2D) {
	const text = "Move using"

	textWidth := rl.MeasureText(text, tutorialPromptFontSize)
	textX := (screenWidth - textWidth) / 2
	imageX := (screenWidth - texture.Width) / 2
	imageY := tutorialPromptTopY + tutorialPromptFontSize + tutorialPromptImageGap

	rl.DrawText(text, textX, tutorialPromptTopY, tutorialPromptFontSize, rl.White)
	rl.DrawTexture(texture, imageX, imageY, rl.White)
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
