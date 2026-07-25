package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type ArtifactFragmentPickupSystem struct {
	pickupFilter   *ecs.Filter4[Position3, Renderable, ArtifactFragmentComponent, ArtifactFragmentPickupComponent]
	fragmentFilter *ecs.Filter3[Position3, Renderable, ArtifactFragmentComponent]
	droneFilter    *ecs.Filter2[Position3, Drone]
	riseMap        *ecs.Map[ArtifactFragmentRiseComponent]
	pickupMap      *ecs.Map[ArtifactFragmentPickupComponent]
	world          *ecs.World
	unloadModel    func(rl.Model)
}

func (system *ArtifactFragmentPickupSystem) Initialize(game *Game) {
	system.pickupFilter = ecs.NewFilter4[Position3, Renderable, ArtifactFragmentComponent, ArtifactFragmentPickupComponent](game.world)
	system.fragmentFilter = ecs.NewFilter3[Position3, Renderable, ArtifactFragmentComponent](game.world)
	system.droneFilter = ecs.NewFilter2[Position3, Drone](game.world)
	system.riseMap = ecs.NewMap[ArtifactFragmentRiseComponent](game.world)
	system.pickupMap = ecs.NewMap[ArtifactFragmentPickupComponent](game.world)
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

	system.startNearestPickup(game.artifactManager, dronePosition)

	query := system.pickupFilter.Query()
	completed := make([]ecs.Entity, 0)
	for query.Next() {
		position, renderable, fragmentComponent, pickup := query.Get()
		if !updateArtifactFragmentPickup(pickup, position, renderable, dronePosition, game.camera, dt) {
			continue
		}

		collectArtifactFragment(game, fragmentComponent.fragment)
		system.unloadModel(*renderable.model)
		renderable.model = nil
		completed = append(completed, query.Entity())
	}
	query.Close()

	for _, entity := range completed {
		game.world.RemoveEntity(entity)
	}
}

func (system *ArtifactFragmentPickupSystem) startNearestPickup(manager *ArtifactManager, dronePosition rl.Vector3) {
	availableWeight := droneMaximumCarryWeight - system.occupiedWeight(manager)
	if availableWeight <= 0 {
		return
	}

	query := system.fragmentFilter.Query()
	var nearestEntity ecs.Entity
	var nearestFragment *ArtifactFragment
	var nearestPosition rl.Vector3
	nearestDistance := float32(0)
	found := false
	for query.Next() {
		position, _, fragmentComponent := query.Get()
		entity := query.Entity()
		fragment := fragmentComponent.fragment
		if fragment == nil || fragment.Collected || fragment.Weight > availableWeight {
			continue
		}
		if system.riseMap.Has(entity) || system.pickupMap.Has(entity) {
			continue
		}

		distance := rl.Vector2Distance(xzVector(dronePosition), xzVector(rl.Vector3(*position)))
		if distance > artifactFragmentPickupProximity {
			continue
		}
		if found && (distance > nearestDistance || distance == nearestDistance && fragment.ID >= nearestFragment.ID) {
			continue
		}

		nearestEntity = entity
		nearestFragment = fragment
		nearestPosition = rl.Vector3(*position)
		nearestDistance = distance
		found = true
	}
	query.Close()

	if found {
		system.pickupMap.Add(nearestEntity, &ArtifactFragmentPickupComponent{
			startPosition: nearestPosition,
		})
	}
}

func (system *ArtifactFragmentPickupSystem) occupiedWeight(manager *ArtifactManager) int {
	weight := manager.CarriedFragmentWeight()

	query := system.pickupFilter.Query()
	for query.Next() {
		_, _, fragmentComponent, _ := query.Get()
		if fragmentComponent.fragment != nil && !fragmentComponent.fragment.Collected {
			weight += fragmentComponent.fragment.Weight
		}
	}
	query.Close()

	return weight
}

func (system *ArtifactFragmentPickupSystem) Unload() {
	if system.fragmentFilter == nil {
		return
	}

	query := system.fragmentFilter.Query()
	entities := make([]ecs.Entity, 0)
	for query.Next() {
		_, renderable, _ := query.Get()
		if renderable.model != nil {
			system.unloadModel(*renderable.model)
			renderable.model = nil
		}
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
	pickup *ArtifactFragmentPickupComponent,
	position *Position3,
	renderable *Renderable,
	dronePosition rl.Vector3,
	camera rl.Camera3D,
	dt float32,
) bool {
	pickup.elapsed = minFloat32(pickup.elapsed+dt, artifactFragmentPickupHomingDuration)
	progress := pickup.elapsed / artifactFragmentPickupHomingDuration
	easedProgress := easeInCubic(progress)
	currentPosition := rl.Vector3Lerp(pickup.startPosition, dronePosition, easedProgress)
	*position = Position3(currentPosition)
	renderable.scale = 1 - easedProgress

	cameraDirection := rl.Vector3Normalize(rl.Vector3Subtract(camera.Position, currentPosition))
	targetRotation := rl.QuaternionFromVector3ToVector3(rl.NewVector3(0, 1, 0), cameraDirection)
	rotation := rl.QuaternionSlerp(rl.QuaternionIdentity(), targetRotation, easedProgress)
	renderable.model.Transform = rl.QuaternionToMatrix(rotation)

	return pickup.elapsed >= artifactFragmentPickupHomingDuration
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
