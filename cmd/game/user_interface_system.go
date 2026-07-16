package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type UserInterfaceSystem struct{}

func (system *UserInterfaceSystem) Initialize(game *Game) {}

func (system *UserInterfaceSystem) Update(game *Game) {
	system.drawDroneViewport(game)
}

func (system *UserInterfaceSystem) drawDroneViewport(game *Game) {
	viewport := droneViewportRectangle()

	rl.DrawRectangle(
		int32(viewport.X-2),
		int32(viewport.Y-2),
		droneViewPixels+4,
		droneViewPixels+4,
		rl.Fade(rl.Black, 0.8),
	)

	rl.DrawTexturePro(
		game.droneFramebuffer.Target.Texture,
		rl.NewRectangle(0, 0, float32(game.droneFramebuffer.Width), -float32(game.droneFramebuffer.Height)),
		viewport,
		rl.Vector2{},
		0,
		rl.White,
	)
}
