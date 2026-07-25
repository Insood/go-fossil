package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type ArtifactFragmentDropOffSystem struct {
	filter         *ecs.Filter3[Position3, Renderable, ArtifactFragmentDropOffComponent]
	droneFilter    *ecs.Filter2[Position3, Drone]
	mapper         *ecs.Map3[Position3, Renderable, ArtifactFragmentDropOffComponent]
	world          *ecs.World
	queue          []*ArtifactFragment
	queued         map[int32]struct{}
	timeSinceEject float32
	hasEjected     bool
	newModel       func(*ArtifactFragment) *rl.Model
	unloadModel    func(rl.Model)
}

func (system *ArtifactFragmentDropOffSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter3[Position3, Renderable, ArtifactFragmentDropOffComponent](game.world)
	system.droneFilter = ecs.NewFilter2[Position3, Drone](game.world)
	system.mapper = ecs.NewMap3[Position3, Renderable, ArtifactFragmentDropOffComponent](game.world)
	system.world = game.world
	system.queued = make(map[int32]struct{})
	if system.newModel == nil {
		system.newModel = newArtifactFragmentPlaneModel
	}
	if system.unloadModel == nil {
		system.unloadModel = rl.UnloadModel
	}
}

func (system *ArtifactFragmentDropOffSystem) Update(game *Game) {
	system.update(game, rl.GetFrameTime())
}

func (system *ArtifactFragmentDropOffSystem) update(game *Game, dt float32) {
	system.updateFlights(game, dt)

	dronePosition, ok := system.dronePosition()
	if !ok {
		return
	}

	padPosition, ok := chargingPadPosition(game)
	if !ok {
		return
	}
	padTarget := padPosition
	padTarget.Y += artifactFragmentDropOffTargetLift

	if len(system.queue) == 0 && droneWithinXZDistance(dronePosition, padPosition, artifactFragmentDropOffProximity) {
		system.enqueueCollectedFragments(game.artifactManager)
	}

	if !system.hasEjected {
		if len(system.queue) > 0 {
			system.ejectNext(dronePosition, padTarget)
		}
		return
	}

	system.timeSinceEject += dt
	if len(system.queue) > 0 && system.timeSinceEject >= artifactFragmentDropOffDelay {
		system.timeSinceEject -= artifactFragmentDropOffDelay
		system.ejectNext(dronePosition, padTarget)
	}
}

func (system *ArtifactFragmentDropOffSystem) updateFlights(game *Game, dt float32) {
	query := system.filter.Query()
	completed := make([]ecs.Entity, 0)
	for query.Next() {
		position, renderable, dropOff := query.Get()
		if !updateArtifactFragmentDropOff(position, dropOff.targetPosition, dt) {
			continue
		}

		system.unloadModel(*renderable.model)
		renderable.model = nil
		completed = append(completed, query.Entity())
	}
	query.Close()

	for _, entity := range completed {
		game.world.RemoveEntity(entity)
	}
}

func (system *ArtifactFragmentDropOffSystem) enqueueCollectedFragments(manager *ArtifactManager) {
	for _, fragment := range sortedArtifactFragments(manager) {
		if _, ok := system.queued[fragment.ID]; ok {
			continue
		}
		system.queue = append(system.queue, fragment)
		system.queued[fragment.ID] = struct{}{}
	}
}

func (system *ArtifactFragmentDropOffSystem) ejectNext(dronePosition, padTarget rl.Vector3) {
	fragment := system.queue[0]
	system.queue = system.queue[1:]
	delete(system.queued, fragment.ID)
	fragment.DroppedOff = true

	system.mapper.NewEntity(
		&Position3{X: dronePosition.X, Y: dronePosition.Y, Z: dronePosition.Z},
		&Renderable{
			model:          system.newModel(fragment),
			scale:          1,
			tint:           rl.White,
			castsShadow:    false,
			receivesShadow: false,
		},
		&ArtifactFragmentDropOffComponent{
			fragment:       fragment,
			targetPosition: padTarget,
		},
	)
	system.hasEjected = true
	system.timeSinceEject = 0
}

func (system *ArtifactFragmentDropOffSystem) dronePosition() (rl.Vector3, bool) {
	query := system.droneFilter.Query()
	defer query.Close()
	if !query.Next() {
		return rl.Vector3{}, false
	}

	position, _ := query.Get()
	return rl.Vector3(*position), true
}

func (system *ArtifactFragmentDropOffSystem) Unload() {
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

func updateArtifactFragmentDropOff(position *Position3, target rl.Vector3, dt float32) bool {
	current := rl.Vector3(*position)
	distance := rl.Vector3Distance(current, target)
	maxDistance := artifactFragmentDropOffSpeed * dt
	if distance <= maxDistance {
		*position = Position3(target)
		return true
	}
	if distance == 0 || maxDistance <= 0 {
		return distance == 0
	}

	direction := rl.Vector3Normalize(rl.Vector3Subtract(target, current))
	*position = Position3(rl.Vector3Add(current, rl.Vector3Scale(direction, maxDistance)))
	return false
}

func chargingPadPosition(game *Game) (rl.Vector3, bool) {
	if game == nil || game.chunkManager == nil {
		return rl.Vector3{}, false
	}

	for _, chunk := range game.chunkManager.Chunks() {
		for _, placement := range chunk.Data.Models {
			if placement.Name == chargingPadModelName {
				return chunkModelPlacementPosition(chunk, placement), true
			}
		}
	}

	return rl.Vector3{}, false
}

func droneWithinXZDistance(dronePosition, targetPosition rl.Vector3, proximity float32) bool {
	return rl.Vector2Distance(xzVector(dronePosition), xzVector(targetPosition)) <= proximity
}
