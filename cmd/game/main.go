package main

import rl "github.com/gen2brain/raylib-go/raylib"

func main() {
	rl.InitWindow(screenWidth, screenHeight, windowTitle)
	defer rl.CloseWindow()

	cameraTarget := rl.NewVector3(gridSize/2, 0, gridSize/2)
	camera := rl.NewCamera3D(
		rl.NewVector3(cameraDistance, cameraHeight, cameraDistance),
		cameraTarget,
		rl.NewVector3(0, 1, 0),
		cameraOrthographicSize,
		rl.CameraOrthographic,
	)

	groundMesh := rl.GenMeshPlane(gridSize, gridSize, gridSubdivisions, gridSubdivisions)
	groundModel := rl.LoadModelFromMesh(groundMesh)
	defer rl.UnloadModel(groundModel)

	gridShader := rl.LoadShader("cmd/game/assets/shaders/grid.vs", "cmd/game/assets/shaders/grid.fs")
	defer rl.UnloadShader(gridShader)

	gridCellsLoc := rl.GetShaderLocation(gridShader, "gridCells")
	lineWidthLoc := rl.GetShaderLocation(gridShader, "lineWidth")
	rl.SetShaderValue(gridShader, gridCellsLoc, []float32{gridSubdivisions}, rl.ShaderUniformFloat)
	rl.SetShaderValue(gridShader, lineWidthLoc, []float32{gridLineWidth}, rl.ShaderUniformFloat)
	groundModel.Materials.Shader = gridShader

	groundPosition := rl.NewVector3(gridSize/2, 0, gridSize/2)
	origin := rl.NewVector3(0, 0, 0)

	rl.SetTargetFPS(targetFPS)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		rl.ClearBackground(rl.SkyBlue)

		rl.BeginMode3D(camera)
		rl.DrawModel(groundModel, groundPosition, 1, rl.Beige)
		rl.DrawLine3D(origin, rl.NewVector3(axisLength, 0, 0), rl.Red)
		rl.DrawLine3D(origin, rl.NewVector3(0, axisLength, 0), rl.Green)
		rl.DrawLine3D(origin, rl.NewVector3(0, 0, axisLength), rl.Blue)
		rl.EndMode3D()

		rl.EndDrawing()
	}
}
