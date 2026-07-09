package main

import (
	"fmt"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type RenderSystem3D struct {
	filter      *ecs.Filter2[Position3, Renderable]
	lightFilter *ecs.Filter1[Light]
}

func (system *RenderSystem3D) Initialize(game *Game) {
	system.filter = ecs.NewFilter2[Position3, Renderable](game.world)
	system.lightFilter = ecs.NewFilter1[Light](game.world)
}

func (system *RenderSystem3D) Update(game *Game) {
	system.renderShadowPass(game)

	if rl.IsKeyPressed(rl.KeyF11) {
		system.saveDepthBufferScreenshot(game)
	}

	system.renderScenePass(game)
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
	rl.BeginMode3D(game.camera)
	system.renderModels(false)
	rl.EndMode3D()
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

	nearPlane := float32(shadowNearPlane)
	farPlane := float32(shadowFarPlane)
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

func (system *RenderSystem3D) withClipPlanes(nearPlane, farPlane float64, draw func()) {
	previousNear := rl.GetCullDistanceNear()
	previousFar := rl.GetCullDistanceFar()

	rl.SetClipPlanes(nearPlane, farPlane)
	defer rl.SetClipPlanes(previousNear, previousFar)

	draw()
}
