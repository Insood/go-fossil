package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type DebugRender3DSystem struct {
	droneFilter *ecs.Filter2[Position3, Drone]
	lightFilter *ecs.Filter1[Light]
}

func (system *DebugRender3DSystem) Initialize(game *Game) {
	system.droneFilter = ecs.NewFilter2[Position3, Drone](game.world)
	system.lightFilter = ecs.NewFilter1[Light](game.world)
}

func (system *DebugRender3DSystem) Update(game *Game) {
	if !debugOverlayVisible {
		return
	}

	rl.BeginMode3D(game.camera)
	system.drawCoordinateSystemAt(rl.NewVector3(0, 0, 0))
	system.drawDroneGroundRay()
	system.drawLightGuide()
	rl.EndMode3D()

	system.drawArtifactLabels(game)
}

func (system *DebugRender3DSystem) Unload() {}

func (system *DebugRender3DSystem) drawCoordinateSystemAt(origin rl.Vector3) {
	rl.DrawLine3D(origin, rl.Vector3Add(origin, rl.NewVector3(debugAxisLength, 0, 0)), rl.Red)
	rl.DrawLine3D(origin, rl.Vector3Add(origin, rl.NewVector3(0, debugAxisLength, 0)), rl.Green)
	rl.DrawLine3D(origin, rl.Vector3Add(origin, rl.NewVector3(0, 0, debugAxisLength)), rl.Blue)
}

func (system *DebugRender3DSystem) drawDroneGroundRay() {
	query := system.droneFilter.Query()
	defer query.Close()

	if !query.Next() {
		return
	}

	position, _ := query.Get()
	start := rl.Vector3(*position)
	end := rl.NewVector3(start.X, 0, start.Z)

	rl.DrawLine3D(start, end, rl.Yellow)
	rl.DrawSphere(end, 0.08, rl.Orange)
}

func (system *DebugRender3DSystem) drawLightGuide() {
	query := system.lightFilter.Query()
	defer query.Close()

	if !query.Next() {
		return
	}

	light := query.Get()
	rl.DrawLine3D(light.origin, light.target, rl.SkyBlue)
	rl.DrawSphere(light.origin, 0.1, rl.Blue)
}

func (system *DebugRender3DSystem) drawArtifactLabels(game *Game) {
	const (
		fontSize = 16
		spacing  = 1
	)

	for _, artifact := range game.artifactManager.artifacts {
		if artifact == nil || artifact.Chunk == nil {
			continue
		}

		worldX := artifact.Chunk.OriginX + artifact.CenterX/float32(terrainTexturePixelsPerTile)
		worldZ := artifact.Chunk.OriginZ + artifact.CenterZ/float32(terrainTexturePixelsPerTile)
		worldY := artifact.Chunk.HeightAtWorldPosition(worldX, worldZ) + 0.25
		screenPos := rl.GetWorldToScreen(rl.NewVector3(worldX, worldY, worldZ), game.camera)
		label := fmt.Sprintf("%d", artifact.ID)

		textSize := rl.MeasureTextEx(rl.GetFontDefault(), label, fontSize, spacing)
		labelPos := rl.NewVector2(screenPos.X-textSize.X/2, screenPos.Y-textSize.Y)

		rl.DrawTextEx(rl.GetFontDefault(), label, rl.NewVector2(labelPos.X+1, labelPos.Y+1), fontSize, spacing, rl.Black)
		rl.DrawTextEx(rl.GetFontDefault(), label, labelPos, fontSize, spacing, rl.Yellow)
	}
}
