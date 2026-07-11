package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type DroneHeightSystem struct {
	filter *ecs.Filter3[Position3, HoverMotion, Drone]
}

func (system *DroneHeightSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter3[Position3, HoverMotion, Drone](game.world)
}

func (system *DroneHeightSystem) Update(game *Game) {
	surface := game.assets.Terrain(defaultLevelName).Surface
	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		position, hover, _ := query.Get()
		hoverOffset := hover.amplitude * float32(math.Sin(float64(float32(rl.GetTime())*hover.angularSpeed)))
		position.Y = surface.SampleHeight(position.X, position.Z) + droneCenterY + hoverOffset
	}
}
