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
	if !system.sounds[laserBurningSoundName].stream.Looping {
		t.Fatal("burning sound looping = false, want true")
	}
	if system.sounds["pickup"].stream.Looping {
		t.Fatal("non-burning sound looping = true, want false")
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

func TestSoundSystemPlaysQueuedRequestsInOrder(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	requestMapper := ecs.NewMap1[SoundPlaybackRequest](world)
	requestMapper.NewEntity(&SoundPlaybackRequest{Name: scoreSoundName})
	requestMapper.NewEntity(&SoundPlaybackRequest{Name: scoreSoundName})

	player := &recordingSoundPlayer{}
	system := &SoundSystem{player: player}
	game := &Game{
		world:  world,
		assets: testAssetManagerWithSounds(scoreSoundName),
	}
	system.Initialize(game)

	system.Update(game)

	scoreID := testSoundID(scoreSoundName)
	assertSoundCalls(t, "plays after first update", player.plays, []uint32{scoreID})
	if got, want := system.sounds[scoreSoundName].queued, 1; got != want {
		t.Fatalf("queued score sounds = %d, want %d", got, want)
	}
	if got := soundPlaybackRequestCount(world); got != 0 {
		t.Fatalf("sound playback request count = %d, want 0", got)
	}

	player.playing[scoreID] = false
	system.Update(game)

	assertSoundCalls(t, "plays after second update", player.plays, []uint32{scoreID, scoreID})
	if got, want := system.sounds[scoreSoundName].queued, 0; got != want {
		t.Fatalf("queued score sounds = %d, want %d", got, want)
	}
}

type recordingSoundPlayer struct {
	plays   []uint32
	updates []uint32
	stops   []uint32
	playing map[uint32]bool
}

func (player *recordingSoundPlayer) Play(sound rl.Music) {
	id := sound.Stream.SampleRate
	player.plays = append(player.plays, id)
	if player.playing == nil {
		player.playing = make(map[uint32]bool)
	}
	player.playing[id] = true
}

func (player *recordingSoundPlayer) Update(sound rl.Music) {
	player.updates = append(player.updates, sound.Stream.SampleRate)
}

func (player *recordingSoundPlayer) Stop(sound rl.Music) {
	id := sound.Stream.SampleRate
	player.stops = append(player.stops, id)
	if player.playing != nil {
		player.playing[id] = false
	}
}

func (player *recordingSoundPlayer) IsPlaying(sound rl.Music) bool {
	return player.playing[sound.Stream.SampleRate]
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

func soundPlaybackRequestCount(world *ecs.World) int {
	filter := ecs.NewFilter1[SoundPlaybackRequest](world)
	query := filter.Query()
	defer query.Close()

	count := 0
	for query.Next() {
		count++
	}
	return count
}
