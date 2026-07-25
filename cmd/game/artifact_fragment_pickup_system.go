package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type ArtifactFragmentPickupSystem struct {
	filter      *ecs.Filter3[Position3, Renderable, ArtifactFragmentComponent]
	droneFilter *ecs.Filter2[Position3, Drone]
	world       *ecs.World
	unloadModel func(rl.Model)
}

func (system *ArtifactFragmentPickupSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter3[Position3, Renderable, ArtifactFragmentComponent](game.world)
	system.droneFilter = ecs.NewFilter2[Position3, Drone](game.world)
	system.world = game.world
	if system.unloadModel == nil {
		system.unloadModel = rl.UnloadModel
	}
}

func (system *ArtifactFragmentPickupSystem) Update(game *Game) {
	system.update(game, rl.GetFrameTime())
}

func (system *ArtifactFragmentPickupSystem) update(game *Game, dt float32) {
	dronePosition, ok := system.dronePosition()
	if !ok {
		return
	}

	query := system.filter.Query()
	completed := make([]ecs.Entity, 0)
	for query.Next() {
		position, renderable, pickup := query.Get()
		if !updateArtifactFragmentPickup(pickup, position, renderable, dronePosition, game.camera, dt) {
			continue
		}

		collectArtifactFragment(game, pickup.fragment)
		system.unloadModel(*renderable.model)
		renderable.model = nil
		completed = append(completed, query.Entity())
	}
	query.Close()

	for _, entity := range completed {
		game.world.RemoveEntity(entity)
	}
}

func (system *ArtifactFragmentPickupSystem) Unload() {
	if system.filter == nil {
		return
	}

	query := system.filter.Query()
	entities := make([]ecs.Entity, 0)
	for query.Next() {
		_, renderable, _ := query.Get()
		system.unloadModel(*renderable.model)
		renderable.model = nil
		entities = append(entities, query.Entity())
	}
	query.Close()

	for _, entity := range entities {
		system.world.RemoveEntity(entity)
	}
}

func (system *ArtifactFragmentPickupSystem) dronePosition() (rl.Vector3, bool) {
	query := system.droneFilter.Query()
	defer query.Close()

	if !query.Next() {
		return rl.Vector3{}, false
	}

	position, _ := query.Get()
	return rl.Vector3(*position), true
}

func updateArtifactFragmentPickup(
	pickup *ArtifactFragmentComponent,
	position *Position3,
	renderable *Renderable,
	dronePosition rl.Vector3,
	camera rl.Camera3D,
	dt float32,
) bool {
	pickup.elapsed = minFloat32(
		pickup.elapsed+dt,
		artifactFragmentPickupRiseDuration+artifactFragmentPickupHomingDuration,
	)

	if pickup.elapsed <= artifactFragmentPickupRiseDuration {
		progress := pickup.elapsed / artifactFragmentPickupRiseDuration
		easedProgress := easeOutCubic(progress)
		*position = Position3(rl.Vector3Lerp(pickup.startPosition, pickup.raisedPosition, easedProgress))
		renderable.scale = 1
		renderable.model.Transform = rl.MatrixIdentity()
		return false
	}

	progress := (pickup.elapsed - artifactFragmentPickupRiseDuration) / artifactFragmentPickupHomingDuration
	easedProgress := easeInCubic(progress)
	currentPosition := rl.Vector3Lerp(pickup.raisedPosition, dronePosition, easedProgress)
	*position = Position3(currentPosition)
	renderable.scale = 1 - easedProgress

	cameraDirection := rl.Vector3Normalize(rl.Vector3Subtract(camera.Position, currentPosition))
	targetRotation := rl.QuaternionFromVector3ToVector3(rl.NewVector3(0, 1, 0), cameraDirection)
	rotation := rl.QuaternionSlerp(rl.QuaternionIdentity(), targetRotation, easedProgress)
	renderable.model.Transform = rl.QuaternionToMatrix(rotation)

	return pickup.elapsed >= artifactFragmentPickupRiseDuration+artifactFragmentPickupHomingDuration
}

func collectArtifactFragment(game *Game, fragment *ArtifactFragment) {
	if fragment.Collected {
		return
	}

	fragment.Collected = true
	game.TotalScore += fragment.Score
}

func easeOutCubic(value float32) float32 {
	inverse := 1 - clampFloat32(value, 0, 1)
	return 1 - inverse*inverse*inverse
}

func easeInCubic(value float32) float32 {
	value = clampFloat32(value, 0, 1)
	return value * value * value
}
