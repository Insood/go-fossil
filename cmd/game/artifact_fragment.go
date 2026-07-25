package main

import (
	"image"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ArtifactFragment struct {
	ID         int32
	Weight     int
	Score      int
	Image      *image.RGBA
	Texture    rl.Texture2D
	Collected  bool
	DroppedOff bool
}
