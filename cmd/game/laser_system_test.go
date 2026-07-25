package main

import (
	"image"
	"math/rand"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"

	"go-fossil/internal/terrain"
)

func TestLaserSystemConsumesTargetsAndPresentsLastTarget(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	manager, chunk := newLaserTestChunkManager(world)
	first := rl.NewVector3(0, 0, 0)
	last := rl.NewVector3(1, 0, 1)
	mapper := ecs.NewMap3[Laser, DroneFireTargets, Battery](world)
	entity := mapper.NewEntity(
		&Laser{},
		&DroneFireTargets{targets: []rl.Vector3{first, last}},
		&Battery{charge: droneBatteryCharge},
	)
	game := &Game{world: world, chunkManager: manager}
	system := &LaserSystem{}
	system.Initialize(game)

	system.Update(game)

	laser := ecs.NewMap[Laser](world).Get(entity)
	if !laser.active {
		t.Fatal("laser is inactive after consuming targets")
	}
	if laser.target != last {
		t.Fatalf("laser target = %+v, want %+v", laser.target, last)
	}
	targets := ecs.NewMap[DroneFireTargets](world).Get(entity)
	if len(targets.targets) != 0 {
		t.Fatalf("remaining target count = %d, want 0", len(targets.targets))
	}
	battery := ecs.NewMap[Battery](world).Get(entity)
	if battery.charge != droneBatteryCharge-laserBatteryDrainPerBurn {
		t.Fatalf("battery charge = %v, want %v", battery.charge, droneBatteryCharge-laserBatteryDrainPerBurn)
	}
	if chunk.BurnOverlayImage.RGBAAt(0, 0).A != 255 || chunk.BurnOverlayImage.RGBAAt(1, 1).A != 255 {
		t.Fatal("laser did not burn every supplied target")
	}
}

func TestLaserSystemClearsTargetsWithoutFiringWhenBatteryIsEmpty(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	manager, chunk := newLaserTestChunkManager(world)
	mapper := ecs.NewMap3[Laser, DroneFireTargets, Battery](world)
	entity := mapper.NewEntity(
		&Laser{active: true},
		&DroneFireTargets{targets: []rl.Vector3{rl.NewVector3(0, 0, 0)}},
		&Battery{},
	)
	game := &Game{world: world, chunkManager: manager}
	system := &LaserSystem{}
	system.Initialize(game)

	system.Update(game)

	if ecs.NewMap[Laser](world).Get(entity).active {
		t.Fatal("laser remained active with an empty battery")
	}
	if targets := ecs.NewMap[DroneFireTargets](world).Get(entity); len(targets.targets) != 0 {
		t.Fatalf("remaining target count = %d, want 0", len(targets.targets))
	}
	if got := chunk.BurnOverlayImage.RGBAAt(0, 0).A; got != 0 {
		t.Fatalf("burn alpha = %d, want 0", got)
	}
}

func TestLaserSystemStaysInactiveWithoutTargets(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	manager, _ := newLaserTestChunkManager(world)
	mapper := ecs.NewMap3[Laser, DroneFireTargets, Battery](world)
	entity := mapper.NewEntity(
		&Laser{active: true},
		&DroneFireTargets{},
		&Battery{charge: droneBatteryCharge},
	)
	game := &Game{world: world, chunkManager: manager}
	system := &LaserSystem{}
	system.Initialize(game)

	system.Update(game)

	if ecs.NewMap[Laser](world).Get(entity).active {
		t.Fatal("laser remained active without targets")
	}
	if got := ecs.NewMap[Battery](world).Get(entity).charge; got != droneBatteryCharge {
		t.Fatalf("battery charge = %v, want %v", got, droneBatteryCharge)
	}
}

func TestLaserSystemApplyBurnTargetsMarksChunkDamaged(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	manager := &ChunkManager{
		world:           world,
		terrainChunkMap: ecs.NewMap1[TerrainChunkComponent](world),
		chunks:          make(map[ChunkCoords]*TerrainChunk),
	}
	chunk := &TerrainChunk{
		Coords:  ChunkCoords{X: 0, Z: 0},
		OriginX: 0,
		OriginZ: 0,
		Data: terrain.ChunkData{
			Width:  4,
			Height: 4,
		},
		BurnOverlayImage: image.NewRGBA(image.Rect(0, 0, 4, 4)),
	}

	manager.registerTerrainChunkEntity(chunk)
	manager.chunks[chunk.Coords] = chunk

	system := &LaserSystem{}
	game := &Game{world: world, chunkManager: manager}
	system.Initialize(game)

	system.applyBurnTargets(game, []rl.Vector3{rl.NewVector3(0, 0, 0)})

	if got := chunk.BurnOverlayImage.RGBAAt(0, 0); got.A != 255 {
		t.Fatalf("burn overlay at (0,0) = %#v, want alpha 255", got)
	}
	if !system.damageMap.Has(chunk.Entity) {
		t.Fatal("chunk entity is missing TerrainChunkDamaged")
	}
}

func TestLaserSystemApplyBurnTargetsIgnoresRepeatedDamageTags(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	manager := &ChunkManager{
		world:           world,
		terrainChunkMap: ecs.NewMap1[TerrainChunkComponent](world),
		chunks:          make(map[ChunkCoords]*TerrainChunk),
	}
	chunk := &TerrainChunk{
		Coords:  ChunkCoords{X: 0, Z: 0},
		OriginX: 0,
		OriginZ: 0,
		Data: terrain.ChunkData{
			Width:  4,
			Height: 4,
		},
		BurnOverlayImage: image.NewRGBA(image.Rect(0, 0, 4, 4)),
	}

	manager.registerTerrainChunkEntity(chunk)
	manager.chunks[chunk.Coords] = chunk

	system := &LaserSystem{}
	game := &Game{world: world, chunkManager: manager}
	system.Initialize(game)

	target := rl.NewVector3(0, 0, 0)
	system.applyBurnTargets(game, []rl.Vector3{target, target})

	if got := chunk.BurnOverlayImage.RGBAAt(0, 0); got.A != 255 {
		t.Fatalf("burn overlay at (0,0) = %#v, want alpha 255", got)
	}
	if !system.damageMap.Has(chunk.Entity) {
		t.Fatal("chunk entity is missing TerrainChunkDamaged")
	}
}

func TestLaserSystemApplyBurnTargetsSkipsEmptyTargets(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	manager := &ChunkManager{
		world:           world,
		terrainChunkMap: ecs.NewMap1[TerrainChunkComponent](world),
		chunks:          make(map[ChunkCoords]*TerrainChunk),
	}
	chunk := &TerrainChunk{
		Coords:  ChunkCoords{X: 0, Z: 0},
		OriginX: 0,
		OriginZ: 0,
		Data: terrain.ChunkData{
			Width:  4,
			Height: 4,
		},
		BurnOverlayImage: image.NewRGBA(image.Rect(0, 0, 4, 4)),
	}

	manager.registerTerrainChunkEntity(chunk)
	manager.chunks[chunk.Coords] = chunk

	system := &LaserSystem{}
	game := &Game{world: world, chunkManager: manager}
	system.Initialize(game)
	system.applyBurnTargets(game, nil)

	if got := chunk.BurnOverlayImage.RGBAAt(0, 0); got.A != 0 {
		t.Fatalf("burn overlay at (0,0) alpha = %d, want 0", got.A)
	}
	if system.damageMap.Has(chunk.Entity) {
		t.Fatal("chunk entity should not be marked damaged")
	}
}

func TestLaserSystemApplyBurnTargetsSpawnsParticlesAfterSuccessfulBurn(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	manager := &ChunkManager{
		world:           world,
		terrainChunkMap: ecs.NewMap1[TerrainChunkComponent](world),
		chunks:          make(map[ChunkCoords]*TerrainChunk),
	}
	chunk := &TerrainChunk{
		Coords:  ChunkCoords{X: 0, Z: 0},
		OriginX: 0,
		OriginZ: 0,
		Data: terrain.ChunkData{
			Width:  4,
			Height: 4,
		},
		BurnOverlayImage: image.NewRGBA(image.Rect(0, 0, 4, 4)),
	}

	manager.registerTerrainChunkEntity(chunk)
	manager.chunks[chunk.Coords] = chunk

	model := &rl.Model{}
	system := &LaserSystem{}
	game := &Game{world: world, chunkManager: manager}
	system.Initialize(game)
	system.particleSpawner = NewParticleSpawnerFactory(world, model, rand.New(rand.NewSource(1)))

	system.applyBurnTargets(game, []rl.Vector3{rl.NewVector3(0, 0, 0)})

	particleFilter := ecs.NewFilter1[Particle](world)
	query := particleFilter.Query()
	defer query.Close()

	count := 0
	for query.Next() {
		count++
	}

	if count != laserStrikeParticleCount {
		t.Fatalf("particle count = %d, want %d", count, laserStrikeParticleCount)
	}
}

func TestDrainLaserBatteryUsesConfiguredAmount(t *testing.T) {
	t.Parallel()

	battery := &Battery{charge: droneBatteryCharge}

	drainLaserBattery(battery)

	if battery.charge != droneBatteryCharge-laserBatteryDrainPerBurn {
		t.Fatalf("battery charge = %.2f, want %.2f", battery.charge, droneBatteryCharge-laserBatteryDrainPerBurn)
	}
}

func TestDrainLaserBatteryClampsAtZero(t *testing.T) {
	t.Parallel()

	battery := &Battery{charge: laserBatteryDrainPerBurn / 2}

	drainLaserBattery(battery)

	if battery.charge != 0 {
		t.Fatalf("battery charge = %.2f, want 0", battery.charge)
	}
}

func newLaserTestChunkManager(world *ecs.World) (*ChunkManager, *TerrainChunk) {
	manager := &ChunkManager{
		world:           world,
		terrainChunkMap: ecs.NewMap1[TerrainChunkComponent](world),
		chunks:          make(map[ChunkCoords]*TerrainChunk),
	}
	chunk := &TerrainChunk{
		Coords: ChunkCoords{X: 0, Z: 0},
		Data: terrain.ChunkData{
			Width:  4,
			Height: 4,
		},
		BurnOverlayImage: image.NewRGBA(image.Rect(0, 0, 4, 4)),
	}
	manager.registerTerrainChunkEntity(chunk)
	manager.chunks[chunk.Coords] = chunk
	return manager, chunk
}
