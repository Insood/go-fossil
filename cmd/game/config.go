package main

const (
	screenWidth            = 1280
	screenHeight           = 800
	windowTitle            = "go-fossil"
	targetFPS              = 60
	gridSize               = 8.0
	gridSubdivisions       = 8
	gridLineWidth          = 0.03
	axisLength             = 3.0
	droneWidth             = 1.0
	droneHeight            = 0.2
	droneDepth             = 1.0
	droneCenterY           = 2.0
	droneTopSpeed          = 1.0
	cameraDistance         = 15.0
	cameraHeight           = 15.0
	cameraOrthographicSize = 10.0
	lightHeight            = 12.0
	shadowDarkness         = 0.55
	shadowMapSize          = 1024
)

var (
	lightOrthographicSize float32 = 6.0
	shadowNearPlane       float32 = 0.5
	shadowFarPlane        float32 = 20.0
	shadowBias            float32 = 0.0003
	shadowSlopeBias       float32 = 0.00001
	debugOverlayVisible           = false
)
