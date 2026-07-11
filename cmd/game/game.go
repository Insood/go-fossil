package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type Game struct {
	assets            *AssetManager
	camera            rl.Camera3D
	shadowFramebuffer *Framebuffer
	world             *ecs.World
	systems           []System
}

func InitializeGame() *Game {
	assets := NewAssetManager()
	assets.Load()

	game := &Game{
		assets:            assets,
		camera:            newCamera(),
		shadowFramebuffer: NewFramebuffer(shadowMapSize, shadowMapSize),
		world:             ecs.NewWorld(),
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
	game.spawnSceneProps()
	game.spawnLight()
}

func (game *Game) registerSystems() {
	game.AddSystem(&DroneInputSystem{})
	game.AddSystem(&PhysicsSystem{})
	game.AddSystem(&LightSystem{})
	game.AddSystem(&RenderSystem3D{})
	game.AddSystem(&DebugRender3DSystem{})
	game.AddSystem(&DebugRenderSystem2D{})
}

func (game *Game) spawnGroundPlane() {
	groundMapper := ecs.NewMap2[Position3, Renderable](game.world)
	groundMapper.NewEntity(
		&Position3{X: gridSize / 2, Y: 0, Z: gridSize / 2},
		&Renderable{
			model:          game.assets.Model("ground"),
			scale:          1.0,
			tint:           rl.White,
			castsShadow:    false,
			receivesShadow: true,
		},
	)
}

func (game *Game) spawnDrone() {
	droneMapper := ecs.NewMap5[Position3, Velocity3, Renderable, Drone, HoverMotion](game.world)
	droneMapper.NewEntity(
		&Position3{X: gridSize / 2, Y: droneCenterY, Z: gridSize / 2},
		&Velocity3{},
		&Renderable{
			model:          game.assets.Model("drone"),
			scale:          1.0,
			tint:           rl.Gray,
			castsShadow:    true,
			receivesShadow: true,
		},
		&Drone{},
		&HoverMotion{
			baseY:        droneCenterY,
			amplitude:    droneHoverAmplitude,
			angularSpeed: droneHoverAngularSpeed,
		},
	)
}

func (game *Game) spawnSceneProps() {
	props := []struct {
		modelName string
		position  rl.Vector3
		tint      rl.Color
	}{
		{
			modelName: "prop_sphere",
			position:  rl.NewVector3(2.0, 1.5, 2.0),
			tint:      rl.SkyBlue,
		},
		{
			modelName: "prop_sphere",
			position:  rl.NewVector3(6.0, 4.2, 5.5),
			tint:      rl.Magenta,
		},
		{
			modelName: "prop_cube",
			position:  rl.NewVector3(2.5, 0.5, 6.0),
			tint:      rl.Orange,
		},
		{
			modelName: "prop_cube",
			position:  rl.NewVector3(5.8, 0.5, 1.8),
			tint:      rl.Lime,
		},
	}

	renderableMapper := ecs.NewMap2[Position3, Renderable](game.world)
	for _, prop := range props {
		position := prop.position
		renderableMapper.NewEntity(
			&Position3{X: position.X, Y: position.Y, Z: position.Z},
			&Renderable{
				model:          game.assets.Model(prop.modelName),
				scale:          1.0,
				tint:           prop.tint,
				castsShadow:    true,
				receivesShadow: true,
			},
		)
	}
}

func (game *Game) spawnLight() {
	lightMapper := ecs.NewMap1[Light](game.world)
	lightMapper.NewEntity(
		newLight(),
	)
}

func newLight() *Light {
	center := rl.NewVector3(gridSize/2, 0, gridSize/2)
	return &Light{
		origin:           rl.NewVector3(center.X+defaultLightOffsetX, lightHeight, center.Z+defaultLightOffsetZ),
		target:           center,
		up:               rl.NewVector3(0, 0, -1),
		orthographicSize: defaultLightSize,
	}
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
	game.shadowFramebuffer.Unload()
	game.assets.Unload()
}
