package main

import (
	"image"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	ecs "github.com/mlange-42/ark/ecs"
)

func TestArtifactFragmentSpawnerQueuesPopSound(t *testing.T) {
	t.Parallel()

	world := ecs.NewWorld()
	factory := NewArtifactFragmentSpawnerFactory(world)
	factory.newModel = func(*ArtifactFragment) *rl.Model {
		return &rl.Model{}
	}
	fragment := &ArtifactFragment{
		ID:    1,
		Image: image.NewRGBA(image.Rect(0, 0, 1, 1)),
	}

	entity := factory.Spawn(fragment, rl.NewVector3(1, 2, 3))

	if !world.Alive(entity) {
		t.Fatal("spawned artifact fragment entity is not alive")
	}
	requestFilter := ecs.NewFilter1[SoundPlaybackRequest](world)
	requestQuery := requestFilter.Query()
	if !requestQuery.Next() {
		requestQuery.Close()
		t.Fatal("artifact fragment creation did not queue a sound")
	}
	if got, want := requestQuery.Get().Name, artifactFragmentCreatedSoundName; got != want {
		requestQuery.Close()
		t.Fatalf("artifact fragment sound request = %q, want %q", got, want)
	}
	if requestQuery.Next() {
		requestQuery.Close()
		t.Fatal("artifact fragment creation queued more than one sound")
	}
	requestQuery.Close()
}
