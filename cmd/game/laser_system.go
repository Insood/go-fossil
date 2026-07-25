package main

import (
	"fmt"
	"math/rand"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type LaserSystem struct {
	filter          *ecs.Filter3[Laser, DroneFireTargets, Battery]
	damageMap       *ecs.Map[TerrainChunkDamaged]
	particleSpawner *ParticleSpawnerFactory
}

func (system *LaserSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter3[Laser, DroneFireTargets, Battery](game.world)
	system.damageMap = ecs.NewMap[TerrainChunkDamaged](game.world)
	if game.assets != nil {
		model, ok := game.assets.LookupModel("particle_cube")
		if ok {
			system.particleSpawner = NewParticleSpawnerFactory(game.world, model, rand.New(rand.NewSource(time.Now().UnixNano())))
		}
	}
}

func (system *LaserSystem) Update(game *Game) {
	query := system.filter.Query()
	burnTargets := make([]rl.Vector3, 0, 1)

	for query.Next() {
		laser, fireTargets, battery := query.Get()
		laser.active = false
		if len(fireTargets.targets) == 0 || battery.charge <= 0 {
			fireTargets.targets = fireTargets.targets[:0]
			continue
		}

		laser.active = true
		laser.target = fireTargets.targets[len(fireTargets.targets)-1]
		drainLaserBattery(battery)
		burnTargets = append(burnTargets, fireTargets.targets...)
		fireTargets.targets = fireTargets.targets[:0]
	}

	query.Close()
	system.applyBurnTargets(game, burnTargets)
}

func (system *LaserSystem) Unload() {}

func drainLaserBattery(battery *Battery) {
	battery.charge -= laserBatteryDrainPerBurn
	if battery.charge < 0 {
		battery.charge = 0
	}
}

func (system *LaserSystem) applyBurnTargets(game *Game, targets []rl.Vector3) {
	if len(targets) == 0 {
		return
	}

	for _, target := range targets {
		chunk, ok := game.chunkManager.ChunkForWorldPosition(target.X, target.Z)
		if !ok {
			panic(fmt.Sprintf("laser target %v is not inside a loaded chunk", target))
		}
		if !game.chunkManager.BurnAtWorldPosition(target.X, target.Z) {
			panic(fmt.Sprintf("failed to burn terrain at %v", target))
		}
		if system.particleSpawner != nil {
			system.particleSpawner.SpawnLaserStrikeParticles(target)
		}
		if system.damageMap.Has(chunk.Entity) {
			continue
		}

		system.damageMap.Add(chunk.Entity, &TerrainChunkDamaged{})
	}
}
