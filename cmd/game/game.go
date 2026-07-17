package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type Game struct {
	assets            *AssetManager
	camera            rl.Camera3D
	artifactManager   *ArtifactManager
	chunkManager      *ChunkManager
	droneFramebuffer  *Framebuffer
	shadowFramebuffer *Framebuffer
	world             *ecs.World
	systems           []System
	Tick              int
	Running           bool
}

func InitializeGame() *Game {
	world := ecs.NewWorld()
	assets := NewAssetManager()
	assets.Load()
	artifactManager := NewArtifactManager()
	chunkManager := NewChunkManager(world, assets, artifactManager)
	defaultChunkCoords := ChunkCoords{X: 0, Z: 0}
	northChunkCoords := ChunkCoords{X: 0, Z: -1}
	chunkGenerator := NewChunkGenerator()

	terrainChunk := chunkManager.LoadDiskChunk(defaultChunkCoords, defaultChunkName)
	chunkManager.LoadGeneratedChunk(northChunkCoords, chunkGenerator)

	game := &Game{
		assets:            assets,
		camera:            newCamera(terrainChunk),
		artifactManager:   artifactManager,
		chunkManager:      chunkManager,
		droneFramebuffer:  NewFramebuffer(droneViewPixels, droneViewPixels),
		shadowFramebuffer: NewFramebuffer(shadowMapSize, shadowMapSize),
		world:             world,
		Running:           true,
	}

	game.spawnInitialEntities()
	game.registerSystems()
	game.InitializeSystems()

	rl.SetTargetFPS(targetFPS)
	return game
}

func newCamera(chunk *TerrainChunk) rl.Camera3D {
	cameraTarget := chunk.Center()
	return rl.NewCamera3D(
		rl.NewVector3(cameraDistance, cameraHeight, cameraDistance),
		cameraTarget,
		rl.NewVector3(0, 1, 0),
		cameraOrthographicSize,
		rl.CameraOrthographic,
	)
}

func (game *Game) spawnInitialEntities() {
	game.spawnDrone()
	// game.spawnSceneProps()
	game.spawnLight()
}

func (game *Game) registerSystems() {
	game.AddSystem(&InputSystem{})
	game.AddSystem(&DroneInputSystem{})
	game.AddSystem(&PhysicsSystem{})
	game.AddSystem(&DroneHeightSystem{})
	game.AddSystem(&CameraSystem{})
	game.AddSystem(&LightSystem{})
	game.AddSystem(&LaserSystem{})
	game.AddSystem(&ArtifactCutoutDetectionSystem{})
	game.AddSystem(&RenderSystem3D{})
	game.AddSystem(&UserInterfaceSystem{})
	game.AddSystem(&DebugRender3DSystem{})
	game.AddSystem(&DebugRenderSystem2D{})
}

func (game *Game) spawnDrone() {
	baseY := game.chunkManager.SampleHeight(droneWorldSpawnX, droneWorldSpawnZ) + droneCenterY
	droneMapper := ecs.NewMap6[Position3, Velocity3, Renderable, Drone, HoverMotion, Laser](game.world)
	droneMapper.NewEntity(
		&Position3{X: droneWorldSpawnX, Y: baseY, Z: droneWorldSpawnZ},
		&Velocity3{},
		&Renderable{
			model:          Must(game.assets.LookupModel("drone")),
			scale:          1.0,
			tint:           rl.White,
			castsShadow:    true,
			receivesShadow: false,
		},
		&Drone{},
		&HoverMotion{
			amplitude:    droneHoverAmplitude,
			angularSpeed: droneHoverAngularSpeed,
		},
		&Laser{},
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
				model:          Must(game.assets.LookupModel(prop.modelName)),
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
		game.newLight(),
	)
}

func (game *Game) newLight() *Light {
	center := game.chunkManager.Chunk(ChunkCoords{X: 0, Z: 0}).Center()
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
	game.Tick++
}

func (game *Game) UnloadAssets() {
	game.chunkManager.Unload()
	game.artifactManager.Unload()
	game.droneFramebuffer.Unload()
	game.shadowFramebuffer.Unload()
	game.assets.Unload()
}
