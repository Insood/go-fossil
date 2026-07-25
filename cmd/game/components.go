package main

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type Position3 rl.Vector3
type Velocity3 rl.Vector3

type Renderable struct {
	model *rl.Model
	scale float32
	tint  color.RGBA

	castsShadow    bool
	receivesShadow bool
}

type HoverMotion struct {
	amplitude    float32
	angularSpeed float32
}

type Light struct {
	origin           rl.Vector3
	target           rl.Vector3
	up               rl.Vector3
	orthographicSize float32
	camera           rl.Camera3D
}

type Laser struct {
	active bool
	target rl.Vector3
}

type Particle struct {
	age       float32
	lifespan  float32
	startTint color.RGBA
	endTint   color.RGBA
}

type ArtifactFragmentComponent struct {
	fragment *ArtifactFragment
}

type MovementAnimationEasing uint8

const (
	MovementAnimationLinear MovementAnimationEasing = iota
	MovementAnimationEaseInCubic
	MovementAnimationEaseOutCubic
)

type MovementAnimationComponent struct {
	startPosition  rl.Vector3
	targetPosition rl.Vector3
	targetEntity   ecs.Entity
	duration       float32
	elapsed        float32
	easing         MovementAnimationEasing
}

type ArtifactFragmentPickupComponent struct{}

type ArtifactFragmentDropOffComponent struct {
	fragment *ArtifactFragment
}

type TerrainChunkComponent struct {
	Chunk *TerrainChunk
}

type TerrainChunkDamaged struct{}

type TutorialMarker struct{}

type ChargingPad struct{}

type Drone struct{}

type PlayerControlled struct{}

type GameOver struct{}

type Battery struct {
	charge float32
}

type BatteryRecharge struct {
	Charge float32
}

type PlayerFireInput struct {
	cursor     rl.Vector2
	firing     bool
	lastCursor rl.Vector2
	lastFiring bool
}

type DroneFireTargets struct {
	targets []rl.Vector3
}
