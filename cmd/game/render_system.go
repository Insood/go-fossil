package main

import (
	"fmt"
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type RenderSystem3D struct {
	filter      *ecs.Filter2[Position3, Renderable]
	lightFilter *ecs.Filter1[Light]
	droneFilter *ecs.Filter2[Position3, Drone]
	laserFilter *ecs.Filter3[Position3, Drone, Laser]
}

func (system *RenderSystem3D) Initialize(game *Game) {
	system.filter = ecs.NewFilter2[Position3, Renderable](game.world)
	system.lightFilter = ecs.NewFilter1[Light](game.world)
	system.droneFilter = ecs.NewFilter2[Position3, Drone](game.world)
	system.laserFilter = ecs.NewFilter3[Position3, Drone, Laser](game.world)
}

func (system *RenderSystem3D) Update(game *Game) {
	system.renderShadowPass(game)

	if rl.IsKeyPressed(rl.KeyF11) {
		system.saveDepthBufferScreenshot(game)
	}

	system.renderScenePass(game)
	system.renderDroneViewportPass(game)
	system.drawDroneViewport(game)
}

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

func (system *RenderSystem3D) drawDroneViewport(game *Game) {
	viewport := droneViewportRectangle()

	rl.DrawRectangle(
		int32(viewport.X-2),
		int32(viewport.Y-2),
		droneViewPixels+4,
		droneViewPixels+4,
		rl.Fade(rl.Black, 0.8),
	)

	rl.DrawTexturePro(
		game.droneFramebuffer.Target.Texture,
		rl.NewRectangle(0, 0, float32(game.droneFramebuffer.Width), -float32(game.droneFramebuffer.Height)),
		viewport,
		rl.Vector2{},
		0,
		rl.White,
	)
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

func (system *RenderSystem3D) saveDepthBufferScreenshot(game *Game) {
	depthRender := rl.LoadRenderTexture(game.shadowFramebuffer.Width, game.shadowFramebuffer.Height)
	if depthRender.ID == 0 {
		return
	}
	defer rl.UnloadRenderTexture(depthRender)

	depthShader := game.assets.Shader("depth_render")
	nearPlaneLoc := rl.GetShaderLocation(depthShader, "nearPlane")
	farPlaneLoc := rl.GetShaderLocation(depthShader, "farPlane")
	isOrthographicLoc := rl.GetShaderLocation(depthShader, "isOrthographic")

	nearPlane := shadowNearPlane
	farPlane := shadowFarPlane
	isOrthographic := float32(0)
	if lightCamera, ok := system.lightCamera(); ok && lightCamera.Projection == rl.CameraOrthographic {
		isOrthographic = 1
	}

	rl.SetShaderValue(depthShader, nearPlaneLoc, []float32{nearPlane}, rl.ShaderUniformFloat)
	rl.SetShaderValue(depthShader, farPlaneLoc, []float32{farPlane}, rl.ShaderUniformFloat)
	rl.SetShaderValue(depthShader, isOrthographicLoc, []float32{isOrthographic}, rl.ShaderUniformFloat)

	rl.BeginTextureMode(depthRender)
	rl.ClearBackground(rl.Black)
	rl.BeginShaderMode(depthShader)
	rl.DrawTexturePro(
		game.shadowFramebuffer.Target.Depth,
		rl.NewRectangle(0, 0, float32(game.shadowFramebuffer.Width), float32(game.shadowFramebuffer.Height)),
		rl.NewRectangle(0, 0, float32(game.shadowFramebuffer.Width), float32(game.shadowFramebuffer.Height)),
		rl.Vector2{},
		0,
		rl.White,
	)
	rl.EndShaderMode()
	rl.EndTextureMode()

	image := rl.LoadImageFromTexture(depthRender.Texture)
	if image == nil {
		return
	}

	defer rl.UnloadImage(image)

	rl.ImageFlipVertical(image)

	fileName := fmt.Sprintf("shadow-depth-%s.png", time.Now().Format("20060102-150405"))
	rl.ExportImage(*image, fileName)
}

func (system *RenderSystem3D) configureShadowReceiverShader(game *Game, lightCamera rl.Camera3D, darkness float32) {
	shadowShader := game.assets.Shader("shadow_receiver")
	lightViewProjectionLoc := rl.GetShaderLocation(shadowShader, "lightViewProjection")
	lightDirectionLoc := rl.GetShaderLocation(shadowShader, "lightDirection")
	shadowBiasLoc := rl.GetShaderLocation(shadowShader, "shadowBias")
	shadowSlopeBiasLoc := rl.GetShaderLocation(shadowShader, "shadowSlopeBias")
	shadowNormalBiasLoc := rl.GetShaderLocation(shadowShader, "shadowNormalBias")
	shadowDarknessLoc := rl.GetShaderLocation(shadowShader, "shadowDarkness")

	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		_, renderable := query.Get()
		if !renderable.receivesShadow {
			continue
		}

		rl.SetMaterialTexture(renderable.model.Materials, rl.MapHeight, game.shadowFramebuffer.Target.Depth)
	}

	for _, chunk := range game.chunkManager.Chunks() {
		rl.SetMaterialTexture(chunk.Model.Materials, rl.MapHeight, game.shadowFramebuffer.Target.Depth)
	}

	lightDirection := rl.Vector3Normalize(rl.Vector3Subtract(lightCamera.Target, lightCamera.Position))
	rl.SetShaderValueMatrix(shadowShader, lightViewProjectionLoc, lightViewProjectionMatrix(lightCamera, game.shadowFramebuffer))
	rl.SetShaderValue(shadowShader, lightDirectionLoc, []float32{lightDirection.X, lightDirection.Y, lightDirection.Z}, rl.ShaderUniformVec3)
	rl.SetShaderValue(shadowShader, shadowBiasLoc, []float32{shadowBias}, rl.ShaderUniformFloat)
	rl.SetShaderValue(shadowShader, shadowSlopeBiasLoc, []float32{shadowSlopeBias}, rl.ShaderUniformFloat)
	rl.SetShaderValue(shadowShader, shadowNormalBiasLoc, []float32{shadowNormalBias}, rl.ShaderUniformFloat)
	rl.SetShaderValue(shadowShader, shadowDarknessLoc, []float32{darkness}, rl.ShaderUniformFloat)
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
