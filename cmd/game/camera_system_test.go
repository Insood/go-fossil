package main

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

func TestCameraSystemSlidesWithinDeadZone(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	droneMapper := ecs.NewMap2[Position3, Drone](world)
	droneEntity := droneMapper.NewEntity(
		&Position3{X: 4, Y: 2, Z: 4},
		&Drone{},
	)

	game := &Game{
		world: world,
		camera: rl.NewCamera3D(
			rl.NewVector3(15, 15, 15),
			rl.NewVector3(4, 0, 4),
			rl.NewVector3(0, 1, 0),
			cameraOrthographicSize,
			rl.CameraOrthographic,
		),
	}

	system := &CameraSystem{}
	system.Initialize(game)

	system.Update(game)
	assertCameraState(t, game.camera, 4, 0.5, 4, 15, 15.5, 15)

	position, _ := droneMapper.Get(droneEntity)
	position.Z = 8
	system.Update(game)
	assertCameraState(t, game.camera, 4, 0.5, 4, 15, 15.5, 15)

	position.Z = 9
	system.Update(game)
	assertCameraState(t, game.camera, 4, 0.5, 5, 15, 15.5, 16)

	position.Y = 4
	system.Update(game)
	assertCameraState(t, game.camera, 4, 2.5, 5, 15, 17.5, 16)

	position.Z = 1
	system.Update(game)
	assertCameraState(t, game.camera, 4, 2.5, 5, 15, 17.5, 16)

	position.Z = 0
	system.Update(game)
	assertCameraState(t, game.camera, 4, 2.5, 4, 15, 17.5, 15)
}

func assertCameraState(t *testing.T, camera rl.Camera3D, wantTargetX, wantTargetY, wantTargetZ, wantPosX, wantPosY, wantPosZ float32) {
	t.Helper()

	if camera.Target.X != wantTargetX || camera.Target.Y != wantTargetY || camera.Target.Z != wantTargetZ {
		t.Fatalf("camera target = (%.2f, %.2f, %.2f), want (%.2f, %.2f, %.2f)", camera.Target.X, camera.Target.Y, camera.Target.Z, wantTargetX, wantTargetY, wantTargetZ)
	}

	if camera.Position.X != wantPosX || camera.Position.Y != wantPosY || camera.Position.Z != wantPosZ {
		t.Fatalf("camera position = (%.2f, %.2f, %.2f), want (%.2f, %.2f, %.2f)", camera.Position.X, camera.Position.Y, camera.Position.Z, wantPosX, wantPosY, wantPosZ)
	}
}
