package main

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type ParticleSystem struct {
	filter *ecs.Filter2[Particle, Renderable]
}

func (system *ParticleSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter2[Particle, Renderable](game.world)
}

func (system *ParticleSystem) Update(game *Game) {
	dt := rl.GetFrameTime()
	query := system.filter.Query()
	expired := make([]ecs.Entity, 0)

	for query.Next() {
		particle, renderable := query.Get()
		if updateParticle(particle, renderable, dt) {
			expired = append(expired, query.Entity())
		}
	}

	query.Close()
	for _, entity := range expired {
		game.world.RemoveEntity(entity)
	}
}

func (system *ParticleSystem) Unload() {}

func updateParticle(particle *Particle, renderable *Renderable, dt float32) bool {
	particle.age += dt
	progress := particle.age / particle.lifespan
	if progress > 1 {
		progress = 1
	}

	renderable.tint = lerpColor(particle.startTint, particle.endTint, progress)
	return particle.age >= particle.lifespan
}

func lerpColor(start, end color.RGBA, t float32) color.RGBA {
	return color.RGBA{
		R: lerpUint8(start.R, end.R, t),
		G: lerpUint8(start.G, end.G, t),
		B: lerpUint8(start.B, end.B, t),
		A: lerpUint8(start.A, end.A, t),
	}
}

func lerpUint8(start, end uint8, t float32) uint8 {
	return uint8(float32(start) + (float32(end)-float32(start))*t)
}
