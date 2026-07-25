package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type ArtifactFragmentPickupSpawnerFactory struct {
	mapper *ecs.Map3[Position3, Renderable, ArtifactFragmentComponent]
}

func NewArtifactFragmentPickupSpawnerFactory(world *ecs.World) *ArtifactFragmentPickupSpawnerFactory {
	return &ArtifactFragmentPickupSpawnerFactory{
		mapper: ecs.NewMap3[Position3, Renderable, ArtifactFragmentComponent](world),
	}
}

func (factory *ArtifactFragmentPickupSpawnerFactory) Spawn(fragment *ArtifactFragment, position rl.Vector3) ecs.Entity {
	model := newArtifactFragmentPlaneModel(fragment)
	raisedPosition := position
	raisedPosition.Y += artifactFragmentPickupRiseHeight

	return factory.mapper.NewEntity(
		&Position3{X: position.X, Y: position.Y, Z: position.Z},
		&Renderable{
			model:          model,
			scale:          1,
			tint:           rl.White,
			castsShadow:    false,
			receivesShadow: false,
		},
		&ArtifactFragmentComponent{
			fragment:       fragment,
			startPosition:  position,
			raisedPosition: raisedPosition,
		},
	)
}

func newArtifactFragmentPlaneModel(fragment *ArtifactFragment) *rl.Model {
	width, length := artifactFragmentPlaneDimensions(fragment)
	model := rl.LoadModelFromMesh(rl.GenMeshPlane(width, length, 1, 1))
	rl.SetMaterialTexture(model.Materials, rl.MapAlbedo, fragment.Texture)
	return &model
}

func artifactFragmentPlaneDimensions(fragment *ArtifactFragment) (float32, float32) {
	bounds := fragment.Image.Bounds()
	return float32(bounds.Dx()) / float32(terrainTexturePixelsPerTile),
		float32(bounds.Dy()) / float32(terrainTexturePixelsPerTile)
}
