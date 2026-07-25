package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type ArtifactFragmentDropOffSystem struct {
	filter            *ecs.Filter3[Position3, Renderable, ArtifactFragmentDropOffComponent]
	droneFilter       *ecs.Filter2[Position3, Drone]
	chargingPadFilter *ecs.Filter2[Position3, ChargingPad]
	mapper            *ecs.Map4[Position3, Renderable, ArtifactFragmentDropOffComponent, MovementAnimationComponent]
	movementMap       *ecs.Map[MovementAnimationComponent]
	world             *ecs.World
	timeSinceEject    float32
	newModel          func(*ArtifactFragment) *rl.Model
	unloadModel       func(rl.Model)
}

func (system *ArtifactFragmentDropOffSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter3[Position3, Renderable, ArtifactFragmentDropOffComponent](game.world)
	system.droneFilter = ecs.NewFilter2[Position3, Drone](game.world)
	system.chargingPadFilter = ecs.NewFilter2[Position3, ChargingPad](game.world)
	system.mapper = ecs.NewMap4[Position3, Renderable, ArtifactFragmentDropOffComponent, MovementAnimationComponent](game.world)
	system.movementMap = ecs.NewMap[MovementAnimationComponent](game.world)
	system.world = game.world
	system.timeSinceEject = artifactFragmentDropOffDelay
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
	system.completeFlights(game)
	system.timeSinceEject += dt
	if system.timeSinceEject < artifactFragmentDropOffDelay {
		return
	}

	fragments := sortedArtifactFragments(game.artifactManager)
	if len(fragments) == 0 {
		return
	}

	dronePosition, ok := system.dronePosition()
	if !ok {
		return
	}

	padPosition, ok := nearestChargingPadPosition(system.chargingPadFilter, dronePosition)
	if !ok {
		return
	}
	padTarget := padPosition
	padTarget.Y += artifactFragmentDropOffTargetLift

	if !droneWithinXZDistance(dronePosition, padPosition, artifactFragmentDropOffProximity) {
		return
	}

	system.ejectNext(fragments[0], dronePosition, padTarget)
}

func (system *ArtifactFragmentDropOffSystem) completeFlights(game *Game) {
	query := system.filter.Query()
	completed := make([]ecs.Entity, 0)
	for query.Next() {
		_, renderable, dropOff := query.Get()
		if system.movementMap.Get(query.Entity()) != nil {
			continue
		}

		game.TotalScore += dropOff.fragment.Score
		system.unloadModel(*renderable.model)
		renderable.model = nil
		completed = append(completed, query.Entity())
	}
	query.Close()

	for _, entity := range completed {
		game.world.RemoveEntity(entity)
	}
}

func (system *ArtifactFragmentDropOffSystem) ejectNext(
	fragment *ArtifactFragment,
	dronePosition,
	padTarget rl.Vector3,
) {
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
			fragment: fragment,
		},
		&MovementAnimationComponent{
			startPosition:  dronePosition,
			targetPosition: padTarget,
			duration:       artifactFragmentDropOffDuration(dronePosition, padTarget),
			easing:         MovementAnimationLinear,
		},
	)
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

func artifactFragmentDropOffDuration(start, target rl.Vector3) float32 {
	if artifactFragmentDropOffSpeed <= 0 {
		return 0
	}
	return rl.Vector3Distance(start, target) / artifactFragmentDropOffSpeed
}

func droneWithinXZDistance(dronePosition, targetPosition rl.Vector3, proximity float32) bool {
	return rl.Vector2Distance(xzVector(dronePosition), xzVector(targetPosition)) <= proximity
}
