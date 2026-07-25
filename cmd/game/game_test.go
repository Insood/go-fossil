package main

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

func TestSpawnDroneAddsPlayerControlled(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	chunk := newTestTerrainChunk(ChunkCoords{X: 0, Z: 0}, 0, 0, 0)
	assets := NewAssetManager()
	assets.models["drone"] = &rl.Model{}
	game := &Game{
		assets:       assets,
		chunkManager: &ChunkManager{chunks: map[ChunkCoords]*TerrainChunk{chunk.Coords: chunk}},
		world:        world,
	}

	game.spawnDrone()

	filter := ecs.NewFilter2[Drone, PlayerControlled](world)
	query := filter.Query()
	defer query.Close()
	if !query.Next() {
		t.Fatal("spawned gameplay drone is missing PlayerControlled")
	}
}

func TestGameplaySystemOrderIncludesGameOverAndTerrainCollision(t *testing.T) {
	t.Parallel()

	game := &Game{}
	game.registerSystems()

	if _, ok := game.systems[3].(*GameOverDetectionSystem); !ok {
		t.Fatalf("system 4 = %T, want *GameOverDetectionSystem", game.systems[3])
	}
	if _, ok := game.systems[7].(*TerrainCollisionDetectionSystem); !ok {
		t.Fatalf("system 8 = %T, want *TerrainCollisionDetectionSystem", game.systems[7])
	}
}

func TestGameTickAdvancesOncePerUpdate(t *testing.T) {
	t.Parallel()

	game := &Game{
		systems: []System{
			&stubSystem{},
			&stubSystem{},
		},
	}

	if got, want := game.Tick, 0; got != want {
		t.Fatalf("initial tick = %d, want %d", got, want)
	}

	game.UpdateSystems()
	if got, want := game.Tick, 1; got != want {
		t.Fatalf("tick after first update = %d, want %d", got, want)
	}

	game.UpdateSystems()
	if got, want := game.Tick, 2; got != want {
		t.Fatalf("tick after second update = %d, want %d", got, want)
	}
}

func TestGameUnloadsSystemsInReverseOrder(t *testing.T) {
	t.Parallel()

	unloadOrder := make([]int, 0, 3)
	game := &Game{}
	game.systems = []System{
		&stubSystem{id: 1, unloadOrder: &unloadOrder},
		&stubSystem{id: 2, unloadOrder: &unloadOrder},
		&stubSystem{id: 3, unloadOrder: &unloadOrder},
	}

	game.UnloadSystems()

	want := []int{3, 2, 1}
	for i := range want {
		if got := unloadOrder[i]; got != want[i] {
			t.Fatalf("unload order[%d] = %d, want %d", i, got, want[i])
		}
	}
}

func TestGameUnloadToleratesMissingOptionalResources(t *testing.T) {
	t.Parallel()

	unloaded := make([]int, 0, 1)
	game := &Game{
		systems: []System{
			&stubSystem{id: 1, unloadOrder: &unloaded},
		},
	}

	game.Unload()

	if len(unloaded) != 1 || unloaded[0] != 1 {
		t.Fatalf("unloaded systems = %v, want [1]", unloaded)
	}
}

type stubSystem struct {
	id          int
	unloadOrder *[]int
}

func (system *stubSystem) Initialize(*Game) {}

func (system *stubSystem) Update(*Game) {}

func (system *stubSystem) Unload() {
	if system.unloadOrder != nil {
		*system.unloadOrder = append(*system.unloadOrder, system.id)
	}
}
