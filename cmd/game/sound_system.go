package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

const (
	laserBurningSoundName = "burning"
	scoreSoundName        = "score"
)

type SoundSystem struct {
	laserFilter   *ecs.Filter1[Laser]
	requestFilter *ecs.Filter1[SoundPlaybackRequest]
	player        soundPlayer
	sounds        map[string]soundPlaybackState
}

type soundPlaybackState struct {
	stream  rl.Music
	playing bool
	queued  int
}

type soundPlayer interface {
	Play(rl.Music)
	Update(rl.Music)
	Stop(rl.Music)
	IsPlaying(rl.Music) bool
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

func (player raylibSoundPlayer) IsPlaying(sound rl.Music) bool {
	return rl.IsMusicStreamPlaying(sound)
}

func (system *SoundSystem) Initialize(game *Game) {
	system.laserFilter = ecs.NewFilter1[Laser](game.world)
	system.requestFilter = ecs.NewFilter1[SoundPlaybackRequest](game.world)
	if system.player == nil {
		system.player = raylibSoundPlayer{}
	}
	system.sounds = make(map[string]soundPlaybackState)

	if game.assets == nil {
		return
	}

	for _, name := range game.assets.SoundNames() {
		stream := Must(game.assets.LookupSound(name))
		stream.Looping = name == laserBurningSoundName
		system.sounds[name] = soundPlaybackState{
			stream: stream,
		}
	}
}

func (system *SoundSystem) Update(game *Game) {
	system.consumePlaybackRequests(game)
	system.setPlaying(laserBurningSoundName, system.anyLaserActive())
	system.updateQueuedSounds()
}

func (system *SoundSystem) Unload() {}

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

func (system *SoundSystem) consumePlaybackRequests(game *Game) {
	query := system.requestFilter.Query()
	requests := make([]ecs.Entity, 0)
	for query.Next() {
		request := query.Get()
		if sound, ok := system.sounds[request.Name]; ok {
			sound.queued++
			system.sounds[request.Name] = sound
		}
		requests = append(requests, query.Entity())
	}
	query.Close()

	for _, entity := range requests {
		game.world.RemoveEntity(entity)
	}
}

func (system *SoundSystem) updateQueuedSounds() {
	for name, sound := range system.sounds {
		if name == laserBurningSoundName || (!sound.playing && sound.queued == 0) {
			continue
		}

		if sound.playing {
			system.player.Update(sound.stream)
			if system.player.IsPlaying(sound.stream) {
				continue
			}
			sound.playing = false
		}

		if sound.queued > 0 {
			system.player.Play(sound.stream)
			system.player.Update(sound.stream)
			sound.playing = true
			sound.queued--
		}
		system.sounds[name] = sound
	}
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
