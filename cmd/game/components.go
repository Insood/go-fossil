package main

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Position3 rl.Vector3

type Renderable struct {
	model *rl.Model
	scale float32
	tint  color.RGBA
}

type Drone struct{}
