package main

import (
	"image/color"
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type ParticleSpawnerFactory struct {
	mapper *ecs.Map4[Position3, Velocity3, Renderable, Particle]
	model  *rl.Model
	rng    *rand.Rand
}

func NewParticleSpawnerFactory(world *ecs.World, model *rl.Model, rng *rand.Rand) *ParticleSpawnerFactory {
	return &ParticleSpawnerFactory{
		mapper: ecs.NewMap4[Position3, Velocity3, Renderable, Particle](world),
		model:  model,
		rng:    rng,
	}
}

func (factory *ParticleSpawnerFactory) SpawnLaserStrikeParticles(position rl.Vector3) {
	for i := 0; i < laserStrikeParticleCount; i++ {
		spawnPosition := position
		spawnPosition.Y += laserStrikeParticleSpawnLift

		factory.mapper.NewEntity(
			&Position3{X: spawnPosition.X, Y: spawnPosition.Y, Z: spawnPosition.Z},
			laserStrikeParticleVelocity(factory.rng),
			&Renderable{
				model:          factory.model,
				scale:          laserStrikeParticleScale,
				tint:           laserStrikeParticleStartTint(),
				castsShadow:    true,
				receivesShadow: false,
			},
			&Particle{
				lifespan:  randomRange(factory.rng, laserStrikeParticleLifespanMin, laserStrikeParticleLifespanMax),
				startTint: laserStrikeParticleStartTint(),
				endTint:   laserStrikeParticleEndTint(),
			},
		)
	}
}

func laserStrikeParticleVelocity(rng *rand.Rand) *Velocity3 {
	speed := randomRange(rng, laserStrikeParticleSpeedMin, laserStrikeParticleSpeedMax)
	coneAngle := randomRange(rng, 0, laserStrikeParticleMaxConeAngle)
	azimuth := randomRange(rng, 0, 2*math.Pi)
	horizontalSpeed := float32(math.Sin(float64(coneAngle))) * speed

	return &Velocity3{
		X: float32(math.Cos(float64(azimuth))) * horizontalSpeed,
		Y: float32(math.Cos(float64(coneAngle))) * speed,
		Z: float32(math.Sin(float64(azimuth))) * horizontalSpeed,
	}
}

func laserStrikeParticleStartTint() color.RGBA {
	return color.RGBA{R: 255, G: 185, B: 70, A: 230}
}

func laserStrikeParticleEndTint() color.RGBA {
	return color.RGBA{R: 255, G: 70, B: 20, A: 0}
}

func randomRange(rng *rand.Rand, min, max float32) float32 {
	return min + rng.Float32()*(max-min)
}
