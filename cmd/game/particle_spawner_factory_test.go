package main

import (
	"math"
	"math/rand"
	"testing"
)

func TestLaserStrikeParticleVelocityStaysInsideUpwardCone(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 1000; i++ {
		velocity := laserStrikeParticleVelocity(rng)
		speed := float32(math.Sqrt(float64(velocity.X*velocity.X + velocity.Y*velocity.Y + velocity.Z*velocity.Z)))
		angle := float32(math.Acos(float64(velocity.Y / speed)))

		if velocity.Y <= 0 {
			t.Fatalf("velocity.Y = %.4f, want positive", velocity.Y)
		}
		if angle > laserStrikeParticleMaxConeAngle+0.0001 {
			t.Fatalf("angle from Y axis = %.4f, want <= %.4f", angle, laserStrikeParticleMaxConeAngle)
		}
		if speed < laserStrikeParticleSpeedMin || speed > laserStrikeParticleSpeedMax {
			t.Fatalf("speed = %.4f, want between %.4f and %.4f", speed, laserStrikeParticleSpeedMin, laserStrikeParticleSpeedMax)
		}
	}
}
