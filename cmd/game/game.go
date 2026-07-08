package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type Game struct {
	assets  *AssetManager
	camera  rl.Camera3D
	world   *ecs.World
	systems []System
}

func InitializeGame() *Game {
	assets := NewAssetManager()
	assets.Load()

	game := &Game{
		assets: assets,
		camera: newCamera(),
		world:  ecs.NewWorld(),
	}

	game.spawnInitialEntities()
	game.registerSystems()
	game.InitializeSystems()

	rl.SetTargetFPS(targetFPS)
	return game
}

func newCamera() rl.Camera3D {
	cameraTarget := rl.NewVector3(gridSize/2, 0, gridSize/2)
	return rl.NewCamera3D(
		rl.NewVector3(cameraDistance, cameraHeight, cameraDistance),
		cameraTarget,
		rl.NewVector3(0, 1, 0),
		cameraOrthographicSize,
		rl.CameraOrthographic,
	)
}

func (game *Game) spawnInitialEntities() {
	game.spawnGroundPlane()
	game.spawnDrone()
}

func (game *Game) registerSystems() {
	game.AddSystem(&RenderSystem3D{})
	game.AddSystem(&DebugRender3DSystem{})
}

func (game *Game) spawnGroundPlane() {
	groundMapper := ecs.NewMap2[Position3, Renderable](game.world)
	groundMapper.NewEntity(
		&Position3{X: gridSize / 2, Y: 0, Z: gridSize / 2},
		&Renderable{
			model: game.assets.Model("ground"),
			scale: 1.0,
			tint:  rl.Beige,
		},
	)
}

func (game *Game) spawnDrone() {
	droneMapper := ecs.NewMap3[Position3, Renderable, Drone](game.world)
	droneMapper.NewEntity(
		&Position3{X: gridSize / 2, Y: droneCenterY, Z: gridSize / 2},
		&Renderable{
			model: game.assets.Model("drone"),
			scale: 1.0,
			tint:  rl.Gray,
		},
		&Drone{},
	)
}

func (game *Game) AddSystem(system System) {
	game.systems = append(game.systems, system)
}

func (game *Game) InitializeSystems() {
	for _, system := range game.systems {
		system.Initialize(game)
	}
}

func (game *Game) UpdateSystems() {
	for _, system := range game.systems {
		system.Update(game)
	}
}

func (game *Game) UnloadAssets() {
	game.assets.Unload()
}
