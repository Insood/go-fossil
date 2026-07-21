package main

import (
	"image/color"
	"testing"

	ecs "github.com/mlange-42/ark/ecs"
)

func TestUpdateParticleFadesTintAlpha(t *testing.T) {
	t.Parallel()

	particle := &Particle{
		lifespan:  1,
		startTint: color.RGBA{R: 255, G: 200, B: 100, A: 200},
		endTint:   color.RGBA{R: 255, G: 60, B: 20, A: 0},
	}
	renderable := &Renderable{tint: particle.startTint}

	if expired := updateParticle(particle, renderable, 0.5); expired {
		t.Fatal("particle expired before its lifespan")
	}
	if got, want := renderable.tint.A, uint8(100); got != want {
		t.Fatalf("renderable tint alpha = %d, want %d", got, want)
	}
}

func TestUpdateParticleExpiresAtLifespan(t *testing.T) {
	t.Parallel()

	particle := &Particle{
		lifespan:  1,
		startTint: color.RGBA{A: 200},
		endTint:   color.RGBA{A: 0},
	}
	renderable := &Renderable{tint: particle.startTint}

	if expired := updateParticle(particle, renderable, 1); !expired {
		t.Fatal("particle should expire at its lifespan")
	}
	if got, want := renderable.tint.A, uint8(0); got != want {
		t.Fatalf("renderable tint alpha = %d, want %d", got, want)
	}
}

func TestParticleSystemRemovesExpiredParticles(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	mapper := ecs.NewMap2[Particle, Renderable](world)
	expiredEntity := mapper.NewEntity(
		&Particle{age: 1, lifespan: 1, startTint: color.RGBA{A: 255}, endTint: color.RGBA{A: 0}},
		&Renderable{},
	)
	liveEntity := mapper.NewEntity(
		&Particle{lifespan: 1, startTint: color.RGBA{A: 255}, endTint: color.RGBA{A: 0}},
		&Renderable{},
	)

	system := &ParticleSystem{}
	game := &Game{world: world}
	system.Initialize(game)
	system.Update(game)

	if world.Alive(expiredEntity) {
		t.Fatal("expired particle entity is still alive")
	}
	if !world.Alive(liveEntity) {
		t.Fatal("unexpired particle entity was removed")
	}
}
