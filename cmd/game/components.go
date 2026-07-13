package main

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
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

type Drone struct{}
