package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type ArtifactFragmentPickupSystem struct {
	pickupFilter   *ecs.Filter4[Position3, Renderable, ArtifactFragmentComponent, ArtifactFragmentPickupComponent]
	fragmentFilter *ecs.Filter3[Position3, Renderable, ArtifactFragmentComponent]
	droneFilter    *ecs.Filter2[Position3, Drone]
	pickupMap      *ecs.Map[ArtifactFragmentPickupComponent]
	movementMap    *ecs.Map[MovementAnimationComponent]
	pickupMoverMap *ecs.Map2[ArtifactFragmentPickupComponent, MovementAnimationComponent]
	world          *ecs.World
	unloadModel    func(rl.Model)
}

func (system *ArtifactFragmentPickupSystem) Initialize(game *Game) {
	system.pickupFilter = ecs.NewFilter4[Position3, Renderable, ArtifactFragmentComponent, ArtifactFragmentPickupComponent](game.world)
	system.fragmentFilter = ecs.NewFilter3[Position3, Renderable, ArtifactFragmentComponent](game.world)
	system.droneFilter = ecs.NewFilter2[Position3, Drone](game.world)
	system.pickupMap = ecs.NewMap[ArtifactFragmentPickupComponent](game.world)
	system.movementMap = ecs.NewMap[MovementAnimationComponent](game.world)
	system.pickupMoverMap = ecs.NewMap2[ArtifactFragmentPickupComponent, MovementAnimationComponent](game.world)
	system.world = game.world
	if system.unloadModel == nil {
		system.unloadModel = rl.UnloadModel
	}
}

func (system *ArtifactFragmentPickupSystem) Update(game *Game) {
	system.update(game)
}

func (system *ArtifactFragmentPickupSystem) update(game *Game) {
	droneEntity, dronePosition, ok := system.drone()
	if !ok {
		return
	}

	query := system.pickupFilter.Query()
	completed := make([]ecs.Entity, 0)
	for query.Next() {
		position, renderable, fragmentComponent, _ := query.Get()
		entity := query.Entity()
		movement := system.movementMap.Get(entity)
		if movement != nil {
			updateArtifactFragmentPickupPresentation(movement, position, renderable, game.camera)
			continue
		}

		collectArtifactFragment(fragmentComponent.fragment)
		system.unloadModel(*renderable.model)
		renderable.model = nil
		completed = append(completed, entity)
	}
	query.Close()

	for _, entity := range completed {
		game.world.RemoveEntity(entity)
	}

	system.startNearestPickup(game.artifactManager, droneEntity, dronePosition)
}

func (system *ArtifactFragmentPickupSystem) startNearestPickup(
	manager *ArtifactManager,
	droneEntity ecs.Entity,
	dronePosition rl.Vector3,
) {
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
		if system.movementMap.Has(entity) || system.pickupMap.Has(entity) {
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
		system.pickupMoverMap.Add(
			nearestEntity,
			&ArtifactFragmentPickupComponent{},
			&MovementAnimationComponent{
				startPosition:  nearestPosition,
				targetPosition: dronePosition,
				targetEntity:   droneEntity,
				duration:       artifactFragmentPickupHomingDuration,
				easing:         MovementAnimationEaseInCubic,
			},
		)
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

func (system *ArtifactFragmentPickupSystem) drone() (ecs.Entity, rl.Vector3, bool) {
	query := system.droneFilter.Query()
	defer query.Close()

	if !query.Next() {
		return ecs.Entity{}, rl.Vector3{}, false
	}

	position, _ := query.Get()
	return query.Entity(), rl.Vector3(*position), true
}

func updateArtifactFragmentPickupPresentation(
	movement *MovementAnimationComponent,
	position *Position3,
	renderable *Renderable,
	camera rl.Camera3D,
) {
	easedProgress := movementAnimationEasedProgress(movementAnimationProgress(movement), movement.easing)
	currentPosition := rl.Vector3(*position)
	renderable.scale = 1 - easedProgress

	cameraDirection := rl.Vector3Normalize(rl.Vector3Subtract(camera.Position, currentPosition))
	targetRotation := rl.QuaternionFromVector3ToVector3(rl.NewVector3(0, 1, 0), cameraDirection)
	rotation := rl.QuaternionSlerp(rl.QuaternionIdentity(), targetRotation, easedProgress)
	renderable.model.Transform = rl.QuaternionToMatrix(rotation)
}

func collectArtifactFragment(fragment *ArtifactFragment) {
	if fragment.Collected {
		return
	}

	fragment.Collected = true
}
