package main

import (
	ecs "github.com/mlange-42/ark/ecs"
)

type GameOverDetectionSystem struct {
	filter              *ecs.Filter5[Drone, PlayerControlled, HoverMotion, Velocity3, Battery]
	playerControlledMap *ecs.Map[PlayerControlled]
	hoverMotionMap      *ecs.Map[HoverMotion]
	gameOverMap         *ecs.Map[GameOver]
	gameOverFilter      *ecs.Filter1[GameOver]
}

func (system *GameOverDetectionSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter5[Drone, PlayerControlled, HoverMotion, Velocity3, Battery](game.world)
	system.playerControlledMap = ecs.NewMap[PlayerControlled](game.world)
	system.hoverMotionMap = ecs.NewMap[HoverMotion](game.world)
	system.gameOverMap = ecs.NewMap[GameOver](game.world)
	system.gameOverFilter = ecs.NewFilter1[GameOver](game.world)
}

func (system *GameOverDetectionSystem) Update(game *Game) {
	query := system.filter.Query()
	depletedDrones := make([]ecs.Entity, 0, 1)

	for query.Next() {
		_, _, _, velocity, battery := query.Get()
		if battery.charge > 0 {
			continue
		}

		velocity.X = 0
		velocity.Y = droneGameOverFallSpeed
		velocity.Z = 0
		depletedDrones = append(depletedDrones, query.Entity())
	}
	query.Close()

	if len(depletedDrones) == 0 {
		return
	}

	for _, entity := range depletedDrones {
		system.playerControlledMap.Remove(entity)
		system.hoverMotionMap.Remove(entity)
	}
	if !gameOverActive(system.gameOverFilter) {
		system.gameOverMap.NewEntity(&GameOver{})
	}
}

func (system *GameOverDetectionSystem) Unload() {}

func gameOverActive(filter *ecs.Filter1[GameOver]) bool {
	query := filter.Query()
	defer query.Close()
	return query.Next()
}
