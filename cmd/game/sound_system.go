package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

const laserBurningSoundName = "burning"

type SoundSystem struct {
	laserFilter *ecs.Filter1[Laser]
	player      soundPlayer
	sounds      map[string]soundPlaybackState
}

type soundPlaybackState struct {
	stream  rl.Music
	playing bool
}

type soundPlayer interface {
	Play(rl.Music)
	Update(rl.Music)
	Stop(rl.Music)
}

type raylibSoundPlayer struct{}

func (player raylibSoundPlayer) Play(sound rl.Music) {
	rl.PlayMusicStream(sound)
}

func (player raylibSoundPlayer) Update(sound rl.Music) {
	rl.UpdateMusicStream(sound)
}

func (player raylibSoundPlayer) Stop(sound rl.Music) {
	rl.StopMusicStream(sound)
}

func (system *SoundSystem) Initialize(game *Game) {
	system.laserFilter = ecs.NewFilter1[Laser](game.world)
	if system.player == nil {
		system.player = raylibSoundPlayer{}
	}
	system.sounds = make(map[string]soundPlaybackState)

	if game.assets == nil {
		return
	}

	for _, name := range game.assets.SoundNames() {
		system.sounds[name] = soundPlaybackState{
			stream: Must(game.assets.LookupSound(name)),
		}
	}
}

func (system *SoundSystem) Update(game *Game) {
	system.setPlaying(laserBurningSoundName, system.anyLaserActive())
}

func (system *SoundSystem) anyLaserActive() bool {
	query := system.laserFilter.Query()
	defer query.Close()

	for query.Next() {
		laser := query.Get()
		if laser.active {
			return true
		}
	}

	return false
}

func (system *SoundSystem) setPlaying(name string, shouldPlay bool) {
	sound, ok := system.sounds[name]
	if !ok {
		return
	}

	if shouldPlay {
		if !sound.playing {
			system.player.Play(sound.stream)
			sound.playing = true
		}
		system.player.Update(sound.stream)
		system.sounds[name] = sound
		return
	}

	if sound.playing {
		system.player.Stop(sound.stream)
		sound.playing = false
		system.sounds[name] = sound
	}
}
