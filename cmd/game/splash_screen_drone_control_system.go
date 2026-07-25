package main

import (
	"math"
	"math/rand"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

type SplashScreenDroneControlSystem struct {
	filter                    *ecs.Filter3[Position3, Velocity3, Drone]
	rng                       *rand.Rand
	center                    rl.Vector3
	maximumDistanceFromCenter float32
	elapsed                   float32
}

func (system *SplashScreenDroneControlSystem) Initialize(game *Game) {
	system.filter = ecs.NewFilter3[Position3, Velocity3, Drone](game.world)
	if system.rng == nil {
		system.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	centerChunk := game.chunkManager.Chunk(ChunkCoords{X: 0, Z: 0})
	system.center = rl.Vector3Add(centerChunk.Center(), rl.Vector3{X: 2, Y: 0, Z: 2})
	system.maximumDistanceFromCenter = splashDroneMaximumMoveDistance
	system.elapsed = splashDroneDirectionDuration
}

func (system *SplashScreenDroneControlSystem) Update(game *Game) {
	var chooseDirection bool
	system.elapsed, chooseDirection = advanceSplashDroneDirectionTimer(system.elapsed, game.FrameTime)

	query := system.filter.Query()
	defer query.Close()

	for query.Next() {
		position, velocity, _ := query.Get()
		if chooseDirection {
			*velocity = randomSplashDroneVelocity(system.rng)
		}
		*velocity = clampDroneVelocityToTerrainBounds(
			*position,
			*velocity,
			game.FrameTime,
			game.chunkManager,
		)
		*velocity = clampSplashDroneVelocityToCenter(
			*position,
			*velocity,
			game.FrameTime,
			system.center,
			system.maximumDistanceFromCenter,
		)
	}
}

func (system *SplashScreenDroneControlSystem) Unload() {}

func advanceSplashDroneDirectionTimer(elapsed, dt float32) (float32, bool) {
	elapsed += max(dt, 0)
	if elapsed < splashDroneDirectionDuration {
		return elapsed, false
	}

	return float32(math.Mod(
		float64(elapsed),
		float64(splashDroneDirectionDuration),
	)), true
}

func randomSplashDroneVelocity(rng *rand.Rand) Velocity3 {
	angle := rng.Float64() * 2 * math.Pi
	direction := rl.NewVector3(float32(math.Cos(angle)), 0, float32(math.Sin(angle)))
	velocity := rl.Vector3Scale(direction, droneTopSpeed)
	return Velocity3(velocity)
}

func clampSplashDroneVelocityToCenter(
	position Position3,
	velocity Velocity3,
	dt float32,
	center rl.Vector3,
	maximumDistance float32,
) Velocity3 {
	if dt <= 0 || maximumDistance < 0 {
		return velocity
	}

	nextX := position.X + velocity.X*dt
	if float32(math.Abs(float64(nextX-center.X))) > maximumDistance {
		velocity.X = 0
	}

	nextZ := position.Z + velocity.Z*dt
	if float32(math.Abs(float64(nextZ-center.Z))) > maximumDistance {
		velocity.Z = 0
	}

	return velocity
}
