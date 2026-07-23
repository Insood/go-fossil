package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func droneViewportRectangle() rl.Rectangle {
	panelSize := float32(droneViewPixels)
	margin := float32(droneViewMargin)
	return rl.NewRectangle(
		float32(screenWidth)-panelSize-margin,
		float32(screenHeight)-panelSize-margin,
		panelSize,
		panelSize,
	)
}

func droneCameraForPosition(dronePosition rl.Vector3) rl.Camera3D {
	cameraPosition := rl.NewVector3(dronePosition.X, dronePosition.Y-droneHeight/2, dronePosition.Z)
	cameraTarget := rl.NewVector3(cameraPosition.X, cameraPosition.Y-1, cameraPosition.Z)

	return rl.NewCamera3D(
		cameraPosition,
		cameraTarget,
		rl.NewVector3(0, 0, -1),
		droneViewFOV,
		rl.CameraPerspective,
	)
}

func droneViewportCursorFromMouse(mouse rl.Vector2, viewport rl.Rectangle) rl.Vector2 {
	return droneViewportCursorFromPixel(clampVector2ToRectangle(mouse, viewport), viewport)
}

func droneViewportCursorFromGamepad(viewport rl.Rectangle, axisX, axisZ float32) rl.Vector2 {
	return rl.NewVector2(
		clampFloat32(axisX, -1, 1),
		clampFloat32(axisZ, -1, 1),
	)
}

func droneViewportCursorFromPixel(pixel rl.Vector2, viewport rl.Rectangle) rl.Vector2 {
	return rl.NewVector2(
		((pixel.X-viewport.X)/viewport.Width)*2-1,
		((pixel.Y-viewport.Y)/viewport.Height)*2-1,
	)
}

func droneViewportCursorPixel(cursor rl.Vector2, viewport rl.Rectangle) rl.Vector2 {
	return rl.NewVector2(
		viewport.X+(clampFloat32(cursor.X, -1, 1)+1)*0.5*viewport.Width,
		viewport.Y+(clampFloat32(cursor.Y, -1, 1)+1)*0.5*viewport.Height,
	)
}

func droneViewportCursorLocalPixel(cursor rl.Vector2, viewport rl.Rectangle) rl.Vector2 {
	pixel := droneViewportCursorPixel(cursor, viewport)
	return rl.NewVector2(pixel.X-viewport.X, pixel.Y-viewport.Y)
}

func droneViewportWorldTarget(
	cursor rl.Vector2,
	dronePosition rl.Vector3,
	sampleHeight func(worldX, worldZ float32) float32,
	hasChunk func(worldX, worldZ float32) bool,
) (rl.Vector3, bool) {
	viewport := droneViewportRectangle()
	if cursor.X < -1 || cursor.X > 1 || cursor.Y < -1 || cursor.Y > 1 {
		return rl.Vector3{}, false
	}

	localCursor := droneViewportCursorLocalPixel(cursor, viewport)
	camera := droneCameraForPosition(dronePosition)
	ray := rl.GetScreenToWorldRayEx(localCursor, camera, int32(viewport.Width), int32(viewport.Height))
	target, ok := terrainRayTarget(ray, sampleHeight, hasChunk)
	if !ok {
		return rl.Vector3{}, false
	}

	return target, true
}

func droneLaserEmitterPosition(dronePosition rl.Vector3) rl.Vector3 {
	return rl.NewVector3(
		dronePosition.X-droneWidth/2,
		dronePosition.Y+droneHeight/2,
		dronePosition.Z-droneDepth/2,
	)
}

func terrainRayTarget(ray rl.Ray, sampleHeight func(worldX, worldZ float32) float32, hasChunk func(worldX, worldZ float32) bool) (rl.Vector3, bool) {
	const (
		maxDistance = droneViewFarPlane
		stepSize    = float32(0.25)
		refineSteps = 8
	)

	start := ray.Position
	if terrainDistanceToSurface(start, sampleHeight, hasChunk) <= 0 {
		return start, true
	}

	prevT := float32(0)
	for t := stepSize; t <= maxDistance; t += stepSize {
		point := terrainRayPoint(ray, t)
		distance := terrainDistanceToSurface(point, sampleHeight, hasChunk)
		if distance <= 0 {
			low := prevT
			high := t
			for i := 0; i < refineSteps; i++ {
				mid := (low + high) / 2
				midPoint := terrainRayPoint(ray, mid)
				if terrainDistanceToSurface(midPoint, sampleHeight, hasChunk) > 0 {
					low = mid
				} else {
					high = mid
				}
			}

			return terrainRayPoint(ray, high), true
		}

		prevT = t
	}

	return rl.Vector3{}, false
}

func terrainRayPoint(ray rl.Ray, distance float32) rl.Vector3 {
	return rl.NewVector3(
		ray.Position.X+ray.Direction.X*distance,
		ray.Position.Y+ray.Direction.Y*distance,
		ray.Position.Z+ray.Direction.Z*distance,
	)
}

func terrainDistanceToSurface(point rl.Vector3, sampleHeight func(worldX, worldZ float32) float32, hasChunk func(worldX, worldZ float32) bool) float32 {
	if !hasChunk(point.X, point.Z) {
		return 1
	}

	return point.Y - sampleHeight(point.X, point.Z)
}

func clampVector2ToRectangle(point rl.Vector2, rec rl.Rectangle) rl.Vector2 {
	return rl.NewVector2(
		clampFloat32(point.X, rec.X, rec.X+rec.Width),
		clampFloat32(point.Y, rec.Y, rec.Y+rec.Height),
	)
}
