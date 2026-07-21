package main

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

func TestSoundSystemRegistersLoadedSounds(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	system := &SoundSystem{player: &recordingSoundPlayer{}}
	system.Initialize(&Game{
		world:  world,
		assets: testAssetManagerWithSounds("pickup", laserBurningSoundName),
	})

	for _, name := range []string{laserBurningSoundName, "pickup"} {
		state, ok := system.sounds[name]
		if !ok {
			t.Fatalf("sound %q was not registered", name)
		}
		if state.playing {
			t.Fatalf("sound %q playing = true, want false", name)
		}
	}
}

func TestSoundSystemStartsAndUpdatesBurningWhileLaserIsActive(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	laserMapper := ecs.NewMap1[Laser](world)
	laserMapper.NewEntity(&Laser{active: true})

	player := &recordingSoundPlayer{}
	system := &SoundSystem{player: player}
	system.Initialize(&Game{
		world:  world,
		assets: testAssetManagerWithSounds(laserBurningSoundName),
	})

	system.Update(&Game{})
	system.Update(&Game{})

	burningID := testSoundID(laserBurningSoundName)
	assertSoundCalls(t, "plays", player.plays, []uint32{burningID})
	assertSoundCalls(t, "updates", player.updates, []uint32{burningID, burningID})
	assertSoundCalls(t, "stops", player.stops, nil)

	if !system.sounds[laserBurningSoundName].playing {
		t.Fatal("burning sound playing = false, want true")
	}
}

func TestSoundSystemStopsBurningWhenLaserBecomesInactive(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	laserMapper := ecs.NewMap1[Laser](world)
	laserEntity := laserMapper.NewEntity(&Laser{active: true})

	player := &recordingSoundPlayer{}
	system := &SoundSystem{player: player}
	system.Initialize(&Game{
		world:  world,
		assets: testAssetManagerWithSounds(laserBurningSoundName),
	})

	system.Update(&Game{})
	laserMapper.Get(laserEntity).active = false
	system.Update(&Game{})
	system.Update(&Game{})

	burningID := testSoundID(laserBurningSoundName)
	assertSoundCalls(t, "plays", player.plays, []uint32{burningID})
	assertSoundCalls(t, "updates", player.updates, []uint32{burningID})
	assertSoundCalls(t, "stops", player.stops, []uint32{burningID})

	if system.sounds[laserBurningSoundName].playing {
		t.Fatal("burning sound playing = true, want false")
	}
}

func TestSoundSystemUsesOneBurningStreamForMultipleActiveLasers(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	laserMapper := ecs.NewMap1[Laser](world)
	laserMapper.NewEntity(&Laser{active: true})
	laserMapper.NewEntity(&Laser{active: true})

	player := &recordingSoundPlayer{}
	system := &SoundSystem{player: player}
	system.Initialize(&Game{
		world:  world,
		assets: testAssetManagerWithSounds(laserBurningSoundName),
	})

	system.Update(&Game{})

	burningID := testSoundID(laserBurningSoundName)
	assertSoundCalls(t, "plays", player.plays, []uint32{burningID})
	assertSoundCalls(t, "updates", player.updates, []uint32{burningID})
	assertSoundCalls(t, "stops", player.stops, nil)
}

func TestSoundSystemLeavesNonDrivenSoundsStopped(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	laserMapper := ecs.NewMap1[Laser](world)
	laserMapper.NewEntity(&Laser{active: true})

	player := &recordingSoundPlayer{}
	system := &SoundSystem{player: player}
	system.Initialize(&Game{
		world:  world,
		assets: testAssetManagerWithSounds(laserBurningSoundName, "pickup"),
	})

	system.Update(&Game{})

	if system.sounds["pickup"].playing {
		t.Fatal("pickup sound playing = true, want false")
	}

	pickupID := testSoundID("pickup")
	for _, id := range append(append([]uint32{}, player.plays...), append(player.updates, player.stops...)...) {
		if id == pickupID {
			t.Fatalf("non-driven pickup sound received an audio call")
		}
	}
}

type recordingSoundPlayer struct {
	plays   []uint32
	updates []uint32
	stops   []uint32
}

func (player *recordingSoundPlayer) Play(sound rl.Music) {
	player.plays = append(player.plays, sound.Stream.SampleRate)
}

func (player *recordingSoundPlayer) Update(sound rl.Music) {
	player.updates = append(player.updates, sound.Stream.SampleRate)
}

func (player *recordingSoundPlayer) Stop(sound rl.Music) {
	player.stops = append(player.stops, sound.Stream.SampleRate)
}

func testAssetManagerWithSounds(names ...string) *AssetManager {
	sounds := make(map[string]rl.Music, len(names))
	for _, name := range names {
		sounds[name] = rl.Music{
			Stream: rl.AudioStream{SampleRate: testSoundID(name)},
		}
	}

	return &AssetManager{sounds: sounds}
}

func testSoundID(name string) uint32 {
	var id uint32
	for _, char := range name {
		id += uint32(char)
	}
	return id
}

func assertSoundCalls(t *testing.T, label string, got, want []uint32) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s = %v, want %v", label, got, want)
		}
	}
}
