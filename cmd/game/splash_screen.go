package main

import (
	ecs "github.com/mlange-42/ark/ecs"
)

type SplashText struct {
	Text     string
	FontSize int32
	Y        int32
}

type SplashSystem interface {
	Initialize(*SplashScreen)
	Update(*SplashScreen)
	Unload()
}

type SplashScreen struct {
	world          *ecs.World
	systems        []SplashSystem
	StartRequested bool
}

func InitializeSplashScreen() *SplashScreen {
	world := ecs.NewWorld()
	screen := &SplashScreen{world: world}

	screen.spawnText()
	screen.AddSystem(&SplashInputSystem{})
	screen.AddSystem(&SplashRenderSystem{})
	screen.InitializeSystems()

	return screen
}

func (screen *SplashScreen) spawnText() {
	mapper := ecs.NewMap1[SplashText](screen.world)
	mapper.NewEntity(&SplashText{
		Text:     splashTitleText,
		FontSize: splashTitleFontSize,
		Y:        splashTitleY,
	})
	mapper.NewEntity(&SplashText{
		Text:     splashStartPromptText,
		FontSize: splashStartPromptFontSize,
		Y:        splashStartPromptY,
	})
}

func (screen *SplashScreen) AddSystem(system SplashSystem) {
	screen.systems = append(screen.systems, system)
}

func (screen *SplashScreen) InitializeSystems() {
	for _, system := range screen.systems {
		system.Initialize(screen)
	}
}

func (screen *SplashScreen) UpdateSystems() {
	for _, system := range screen.systems {
		system.Update(screen)
	}
}

func (screen *SplashScreen) Unload() {
	for i := len(screen.systems) - 1; i >= 0; i-- {
		screen.systems[i].Unload()
	}
	screen.systems = nil
	screen.world = nil
}
