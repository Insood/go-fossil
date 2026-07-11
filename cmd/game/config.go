package main

import "math"

const (
	screenWidth  = 1280
	screenHeight = 800
	windowTitle  = "go-fossil"
	targetFPS    = 60

	defaultLevelName            = "level01"
	terrainTexturePixelsPerTile = 64

	debugAxisLength = 3.0

	droneWidth             = 1.0
	droneHeight            = 0.2
	droneDepth             = 1.0
	droneCenterY           = 2.0
	droneTopSpeed          = 4.0
	droneHoverAmplitude    = 0.05
	droneHoverCyclesPerSec = 0.2

	droneGamepadIndex    = 0
	droneGamepadAxisX    = 0
	droneGamepadAxisZ    = 1
	droneGamepadDeadzone = 0.10

	cameraDistance         = 15.0
	cameraHeight           = 15.0
	cameraOrthographicSize = 10.0

	lightHeight         = 10
	defaultLightOffsetX = -0.75
	defaultLightOffsetZ = -0.75
	defaultLightSize    = 10.0
	shadowDarkness      = 0.55
	shadowMapSize       = 2048
)

var (
	shadowNearPlane        float32 = 0.5
	shadowFarPlane         float32 = 20.0
	shadowBias             float32 = 0.0008
	shadowSlopeBias        float32 = 0.002
	debugOverlayVisible            = false
	droneHoverAngularSpeed         = float32(2 * math.Pi * droneHoverCyclesPerSec)
)
