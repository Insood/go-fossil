package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 1280
	screenHeight = 800
	windowTitle  = "go-fossil"
	targetFPS    = 60

	terrainTexturePixelsPerTile = 64

	debugAxisLength = 3.0

	defaultChunkName = "default"
	droneWorldSpawnX = 1.5
	droneWorldSpawnZ = 6.5

	droneWidth             = 1.0
	droneHeight            = 0.2
	droneDepth             = 1.0
	droneCenterY           = 2.0
	droneTopSpeed          = 4.0
	droneHoverAmplitude    = 0.05
	droneHoverCyclesPerSec = 0.2

	droneGamepadIndex       = 0
	droneGamepadMoveAxisX   = 0
	droneGamepadMoveAxisZ   = 1
	droneGamepadTargetAxisX = 2
	droneGamepadTargetAxisZ = 3
	droneGamepadDeadzone    = 0.10

	cameraDistance         = 15.0
	cameraHeight           = 15.0
	cameraOrthographicSize = 10.0
	cameraFollowDeadZoneXZ = 4.0
	cameraFollowDeadZoneY  = 1.5

	droneViewFOV       = 45
	droneViewPixels    = 256
	droneViewMargin    = 16
	droneViewNearPlane = 0.01
	droneViewFarPlane  = 16.0

	artifactFragmentThumbSize    = 64
	artifactFragmentRowStep      = 70
	artifactFragmentStartX       = 16
	artifactFragmentStartY       = 16
	artifactFragmentTextGap      = 12
	artifactFragmentDisplayCount = 8
	artifactFragmentMinPixels    = 20
	totalScoreTextX              = 16
	totalScoreTextY              = 16

	generatedArtifactEdgeMarginPixels   = 96
	generatedArtifactMinCenterGapPixels = 160
	generatedArtifactMaxCount           = 3
	generatedArtifactPlacementAttempts  = 64

	laserHitMarkerRadius             = 0.015
	laserCursorBurnStepPixels        = 5.0
	artifactCutoutDetectionScanTicks = 60
	MaximumRegionSize                = 4096
	burnOverlayAlpha                 = 255
	dugOutOverlayAlpha               = 128

	lightHeight         = 10
	defaultLightOffsetX = 0
	defaultLightOffsetZ = 0
	defaultLightSize    = 10.0
	shadowDarkness      = 0.55
	shadowMapSize       = 2048
)

var (
	gamePadQuitButton1     int32   = rl.GamepadButtonRightFaceUp
	shadowNearPlane        float32 = 0.5
	shadowFarPlane         float32 = 20.0
	shadowBias             float32 = 0.0008
	shadowSlopeBias        float32 = 0.002
	shadowNormalBias       float32 = 0.02
	debugOverlayVisible            = false
	droneHoverAngularSpeed         = float32(2 * math.Pi * droneHoverCyclesPerSec)
)
