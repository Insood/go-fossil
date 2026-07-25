package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type MovementAnimationSystem struct {
	filter      *ecs.Filter2[Position3, MovementAnimationComponent]
	positionMap *ecs.Map[Position3]
	movementMap *ecs.Map[MovementAnimationComponent]
	world       *ecs.World
}

func (system *MovementAnimationSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter2[Position3, MovementAnimationComponent](game.world)
	system.positionMap = ecs.NewMap[Position3](game.world)
	system.movementMap = ecs.NewMap[MovementAnimationComponent](game.world)
	system.world = game.world
}

func (system *MovementAnimationSystem) Update(*Game) {
	system.update(rl.GetFrameTime())
}

func (system *MovementAnimationSystem) update(dt float32) {
	query := system.filter.Query()
	completed := make([]ecs.Entity, 0)
	for query.Next() {
		position, movement := query.Get()
		target := system.targetPosition(movement)
		if updateMovementAnimation(movement, position, target, dt) {
			completed = append(completed, query.Entity())
		}
	}
	query.Close()

	for _, entity := range completed {
		system.movementMap.Remove(entity)
	}
}

func (system *MovementAnimationSystem) targetPosition(movement *MovementAnimationComponent) rl.Vector3 {
	if movement.targetEntity.IsZero() || !system.world.Alive(movement.targetEntity) {
		return movement.targetPosition
	}

	position := system.positionMap.Get(movement.targetEntity)
	if position == nil {
		return movement.targetPosition
	}
	return rl.Vector3(*position)
}

func (system *MovementAnimationSystem) Unload() {}

func updateMovementAnimation(
	movement *MovementAnimationComponent,
	position *Position3,
	target rl.Vector3,
	dt float32,
) bool {
	if movement.duration <= 0 {
		*position = Position3(target)
		movement.elapsed = movement.duration
		return true
	}

	movement.elapsed = minFloat32(movement.elapsed+dt, movement.duration)
	progress := movementAnimationProgress(movement)
	*position = Position3(rl.Vector3Lerp(movement.startPosition, target, movementAnimationEasedProgress(progress, movement.easing)))
	return movement.elapsed >= movement.duration
}

func movementAnimationProgress(movement *MovementAnimationComponent) float32 {
	if movement == nil || movement.duration <= 0 {
		return 1
	}
	return clampFloat32(movement.elapsed/movement.duration, 0, 1)
}

func movementAnimationEasedProgress(progress float32, easing MovementAnimationEasing) float32 {
	progress = clampFloat32(progress, 0, 1)
	switch easing {
	case MovementAnimationEaseInCubic:
		return progress * progress * progress
	case MovementAnimationEaseOutCubic:
		inverse := 1 - progress
		return 1 - inverse*inverse*inverse
	default:
		return progress
	}
}
