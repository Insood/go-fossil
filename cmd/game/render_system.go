package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type RenderSystem3D struct {
	filter            *ecs.Filter2[Position3, Renderable]
	lightFilter       *ecs.Filter1[Light]
	droneFilter       *ecs.Filter2[Position3, Drone]
	laserFilter       *ecs.Filter3[Position3, Drone, Laser]
	skipDroneViewport bool
}

func (system *RenderSystem3D) Initialize(game *Game) {
	system.filter = ecs.NewFilter2[Position3, Renderable](game.world)
	system.lightFilter = ecs.NewFilter1[Light](game.world)
	system.droneFilter = ecs.NewFilter2[Position3, Drone](game.world)
	system.laserFilter = ecs.NewFilter3[Position3, Drone, Laser](game.world)
}

func (system *RenderSystem3D) Update(game *Game) {
	system.renderShadowPass(game)

	system.renderScenePass(game)
	if !system.skipDroneViewport {
		system.renderDroneViewportPass(game)
	}
}

func (system *RenderSystem3D) Unload() {}

func (system *RenderSystem3D) renderShadowPass(game *Game) {
	lightCamera, ok := system.lightCamera()
	if !ok {
		return
	}

	system.withClipPlanes(shadowNearPlane, shadowFarPlane, func() {
		rl.BeginTextureMode(game.shadowFramebuffer.Target)
		rl.ClearBackground(rl.White)

		rl.BeginMode3D(lightCamera)
		system.renderModels(true)
		rl.EndMode3D()

		rl.EndTextureMode()
	})
}

func (system *RenderSystem3D) renderScenePass(game *Game) {
	lightCamera, ok := system.lightCamera()
	if ok {
		system.configureShadowReceiverShader(game, lightCamera, shadowDarkness)
	}

	rl.BeginMode3D(game.camera)
	system.renderTerrainChunks(game)
	system.renderModels(false)
	system.renderLasers()
	rl.EndMode3D()
}

func (system *RenderSystem3D) renderDroneViewportPass(game *Game) {
	droneCamera, ok := system.droneCamera()
	if !ok {
		return
	}

	lightCamera, hasLight := system.lightCamera()
	if hasLight {
		system.configureShadowReceiverShader(game, lightCamera, 0)
	}

	system.withClipPlanes(droneViewNearPlane, droneViewFarPlane, func() {
		rl.BeginTextureMode(game.droneFramebuffer.Target)
		rl.ClearBackground(rl.SkyBlue)

		rl.BeginMode3D(droneCamera)
		system.renderTerrainChunks(game)
		system.renderModels(false)
		system.renderLasers()
		rl.EndMode3D()

		rl.EndTextureMode()
	})
}

func (system *RenderSystem3D) lightCamera() (rl.Camera3D, bool) {
	query := system.lightFilter.Query()
	defer query.Close()

	if !query.Next() {
		return rl.Camera3D{}, false
	}

	light := query.Get()
	return light.camera, true
}

func (system *RenderSystem3D) droneCamera() (rl.Camera3D, bool) {
	query := system.droneFilter.Query()
	defer query.Close()

	if !query.Next() {
		return rl.Camera3D{}, false
	}

	position, _ := query.Get()
	return droneCameraForPosition(rl.Vector3(*position)), true
}

func (system *RenderSystem3D) renderModels(onlyShadowCasters bool) {
	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		position, renderable := query.Get()
		if onlyShadowCasters && !renderable.castsShadow {
			continue
		}

		rl.DrawModel(*renderable.model, rl.Vector3(*position), renderable.scale, renderable.tint)
	}
}

func (system *RenderSystem3D) renderTerrainChunks(game *Game) {
	for _, chunk := range game.chunkManager.Chunks() {
		rl.DrawModel(*chunk.Model, chunk.Center(), 1.0, rl.White)
	}
}

func (system *RenderSystem3D) renderLasers() {
	query := system.laserFilter.Query()
	defer query.Close()

	for query.Next() {
		position, _, laser := query.Get()
		if !laser.active {
			continue
		}

		start := droneLaserEmitterPosition(rl.Vector3(*position))
		rl.DrawLine3D(start, laser.target, rl.Red)
		rl.DrawSphere(laser.target, laserHitMarkerRadius, rl.Red)
	}
}

func (system *RenderSystem3D) configureShadowReceiverShader(game *Game, lightCamera rl.Camera3D, darkness float32) {
	shadowShader := Must(game.assets.LookupShader("terrain_shader"))
	modelShadowShader := Must(game.assets.LookupShader("model_shadow_receiver"))
	slopeShadeStrengthLoc := rl.GetShaderLocation(shadowShader, "slopeShadeStrength")
	terrainCutoutDivotDepthLoc := rl.GetShaderLocation(shadowShader, "terrainCutoutDivotDepth")
	terrainCutoutOverlayAlphaLoc := rl.GetShaderLocation(shadowShader, "terrainCutoutOverlayAlpha")

	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		_, renderable := query.Get()
		if !renderable.receivesShadow {
			continue
		}

		materials := renderable.model.GetMaterials()
		for i := range materials {
			rl.SetMaterialTexture(&materials[i], rl.MapNormal, game.shadowFramebuffer.Target.Depth)
		}
	}

	for _, chunk := range game.chunkManager.Chunks() {
		rl.SetMaterialTexture(chunk.Model.Materials, rl.MapHeight, game.shadowFramebuffer.Target.Depth)
	}

	lightDirection := rl.Vector3Normalize(rl.Vector3Subtract(lightCamera.Target, lightCamera.Position))
	lightViewProjection := lightViewProjectionMatrix(lightCamera, game.shadowFramebuffer)
	configureShadowShaderValues(shadowShader, lightViewProjection, lightDirection, darkness)
	configureShadowShaderValues(modelShadowShader, lightViewProjection, lightDirection, darkness)
	rl.SetShaderValue(shadowShader, slopeShadeStrengthLoc, []float32{slopeShadeStrength}, rl.ShaderUniformFloat)
	rl.SetShaderValue(shadowShader, terrainCutoutDivotDepthLoc, []float32{terrainCutoutDivotDepth}, rl.ShaderUniformFloat)
	rl.SetShaderValue(shadowShader, terrainCutoutOverlayAlphaLoc, []float32{float32(dugOutOverlayAlpha) / 255}, rl.ShaderUniformFloat)
}

func configureShadowShaderValues(shader rl.Shader, lightViewProjection rl.Matrix, lightDirection rl.Vector3, darkness float32) {
	lightViewProjectionLoc := rl.GetShaderLocation(shader, "lightViewProjection")
	lightDirectionLoc := rl.GetShaderLocation(shader, "lightDirection")
	shadowBiasLoc := rl.GetShaderLocation(shader, "shadowBias")
	shadowSlopeBiasLoc := rl.GetShaderLocation(shader, "shadowSlopeBias")
	shadowNormalBiasLoc := rl.GetShaderLocation(shader, "shadowNormalBias")
	shadowDarknessLoc := rl.GetShaderLocation(shader, "shadowDarkness")

	rl.SetShaderValueMatrix(shader, lightViewProjectionLoc, lightViewProjection)
	rl.SetShaderValue(shader, lightDirectionLoc, []float32{lightDirection.X, lightDirection.Y, lightDirection.Z}, rl.ShaderUniformVec3)
	rl.SetShaderValue(shader, shadowBiasLoc, []float32{shadowBias}, rl.ShaderUniformFloat)
	rl.SetShaderValue(shader, shadowSlopeBiasLoc, []float32{shadowSlopeBias}, rl.ShaderUniformFloat)
	rl.SetShaderValue(shader, shadowNormalBiasLoc, []float32{shadowNormalBias}, rl.ShaderUniformFloat)
	rl.SetShaderValue(shader, shadowDarknessLoc, []float32{darkness}, rl.ShaderUniformFloat)
}

func (system *RenderSystem3D) withClipPlanes(nearPlane, farPlane float32, draw func()) {
	previousNear := rl.GetCullDistanceNear()
	previousFar := rl.GetCullDistanceFar()

	rl.SetClipPlanes(float64(nearPlane), float64(farPlane))
	defer rl.SetClipPlanes(previousNear, previousFar)

	draw()
}

func lightViewProjectionMatrix(camera rl.Camera3D, framebuffer *Framebuffer) rl.Matrix {
	aspectRatio := float32(framebuffer.Width) / float32(framebuffer.Height)
	view := rl.GetCameraMatrix(camera)

	if camera.Projection == rl.CameraOrthographic {
		top := camera.Fovy / 2
		right := top * aspectRatio
		projection := rl.MatrixOrtho(-right, right, -top, top, shadowNearPlane, shadowFarPlane)
		return rl.MatrixMultiply(view, projection)
	}

	projection := rl.MatrixPerspective(camera.Fovy*float32(math.Pi/180), aspectRatio, shadowNearPlane, shadowFarPlane)
	return rl.MatrixMultiply(view, projection)
}
