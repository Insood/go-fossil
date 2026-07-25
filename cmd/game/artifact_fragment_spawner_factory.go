package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type ArtifactFragmentSpawnerFactory struct {
	mapper *ecs.Map4[Position3, Renderable, ArtifactFragmentComponent, MovementAnimationComponent]
}

func NewArtifactFragmentSpawnerFactory(world *ecs.World) *ArtifactFragmentSpawnerFactory {
	return &ArtifactFragmentSpawnerFactory{
		mapper: ecs.NewMap4[Position3, Renderable, ArtifactFragmentComponent, MovementAnimationComponent](world),
	}
}

func (factory *ArtifactFragmentSpawnerFactory) Spawn(fragment *ArtifactFragment, position rl.Vector3) ecs.Entity {
	model := newArtifactFragmentPlaneModel(fragment)

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
			fragment: fragment,
		},
		artifactFragmentRiseMovement(position),
	)
}

func artifactFragmentRiseMovement(position rl.Vector3) *MovementAnimationComponent {
	raisedPosition := position
	raisedPosition.Y += artifactFragmentPickupRiseHeight
	return &MovementAnimationComponent{
		startPosition:  position,
		targetPosition: raisedPosition,
		duration:       artifactFragmentPickupRiseDuration,
		easing:         MovementAnimationEaseOutCubic,
	}
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
