package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type SplashRenderSystem struct {
	filter *ecs.Filter1[SplashText]
}

func (system *SplashRenderSystem) Initialize(screen *SplashScreen) {
	system.filter = ecs.NewFilter1[SplashText](screen.world)
}

func (system *SplashRenderSystem) Update(screen *SplashScreen) {
	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		text := query.Get()
		textWidth := rl.MeasureText(text.Text, text.FontSize)
		textX := (screenWidth - textWidth) / 2
		rl.DrawText(text.Text, textX, text.Y, text.FontSize, rl.White)
	}
}

func (system *SplashRenderSystem) Unload() {}
