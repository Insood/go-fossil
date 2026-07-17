package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type UserInterfaceSystem struct {
	filter *ecs.Filter1[DroneFireControl]
}

func (system *UserInterfaceSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter1[DroneFireControl](game.world)
}

func (system *UserInterfaceSystem) Update(game *Game) {
	system.drawDroneViewport(game)
	system.drawDroneReticle()
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

func (system *UserInterfaceSystem) drawDroneReticle() {
	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		control := query.Get()
		center := control.cursor
		halfSize := int32(5)

		rl.DrawLine(
			int32(center.X)-halfSize,
			int32(center.Y),
			int32(center.X)+halfSize,
			int32(center.Y),
			rl.Red,
		)
		rl.DrawLine(
			int32(center.X),
			int32(center.Y)-halfSize,
			int32(center.X),
			int32(center.Y)+halfSize,
			rl.Red,
		)
	}
}
