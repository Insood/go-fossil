package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
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
	scene          *Game
	systems        []SplashSystem
	FrameTime      float32
	StartRequested bool
}

func InitializeSplashScreen(assets *AssetManager) *SplashScreen {
	world := ecs.NewWorld()
	artifactManager := NewArtifactManager()
	chunkManager := NewChunkManager(world, assets, artifactManager)
	generator := NewChunkGenerator()

	var centerChunk *TerrainChunk
	for _, coords := range splashChunkCoords() {
		chunk := chunkManager.LoadGeneratedChunk(coords, generator)
		if coords == (ChunkCoords{X: 0, Z: 0}) {
			centerChunk = chunk
		}
	}

	scene := &Game{
		assets:            assets,
		camera:            newCamera(centerChunk),
		artifactManager:   artifactManager,
		chunkManager:      chunkManager,
		shadowFramebuffer: NewFramebuffer(shadowMapSize, shadowMapSize),
		world:             world,
		Running:           true,
	}
	screen := &SplashScreen{
		world: world,
		scene: scene,
	}

	screen.spawnDrone(centerChunk)
	scene.spawnLight()
	screen.spawnText()
	screen.registerSceneSystems()
	screen.AddSystem(&SplashInputSystem{})
	screen.AddSystem(&SplashRenderSystem{})
	scene.InitializeSystems()
	screen.InitializeSystems()

	return screen
}

func splashChunkCoords() []ChunkCoords {
	return []ChunkCoords{
		{X: 0, Z: 0},
		{X: 0, Z: -1},
		{X: 1, Z: 0},
		{X: 0, Z: 1},
		{X: -1, Z: 0},
		{X: -1, Z: -1},
		{X: 1, Z: -1},
		{X: 1, Z: 1},
		{X: -1, Z: 1},
	}
}

func (screen *SplashScreen) spawnDrone(centerChunk *TerrainChunk) {
	center := centerChunk.Center()
	position := Position3{
		X: center.X,
		Y: screen.scene.chunkManager.SampleHeight(center.X, center.Z) + droneCenterY,
		Z: center.Z,
	}
	model := Must(screen.scene.assets.LookupModel("drone"))
	spawnSplashDrone(screen.world, model, position)
}

func spawnSplashDrone(world *ecs.World, model *rl.Model, position Position3) ecs.Entity {
	mapper := ecs.NewMap5[Position3, Velocity3, Renderable, Drone, HoverMotion](world)
	return mapper.NewEntity(
		&position,
		&Velocity3{},
		&Renderable{
			model:          model,
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
	)
}

func (screen *SplashScreen) registerSceneSystems() {
	screen.scene.AddSystem(&SplashScreenDroneControlSystem{})
	screen.scene.AddSystem(&PhysicsSystem{})
	screen.scene.AddSystem(&DroneHeightSystem{})
	screen.scene.AddSystem(&LightSystem{})
	screen.scene.AddSystem(&RenderSystem3D{skipDroneViewport: true})
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
	screen.scene.FrameTime = screen.FrameTime
	screen.scene.UpdateSystems()
	for _, system := range screen.systems {
		system.Update(screen)
	}
}

func (screen *SplashScreen) Unload() {
	for i := len(screen.systems) - 1; i >= 0; i-- {
		screen.systems[i].Unload()
	}
	if screen.scene != nil {
		screen.scene.Unload()
	}
	screen.systems = nil
	screen.scene = nil
	screen.world = nil
}
