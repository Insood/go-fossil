package main

import "testing"

func TestSortedArtifactFragmentsOrdersByID(t *testing.T) {
	t.Parallel()

	manager := NewArtifactManager()
	manager.fragments[3] = &ArtifactFragment{ID: 3}
	manager.fragments[1] = &ArtifactFragment{ID: 1}
	manager.fragments[2] = &ArtifactFragment{ID: 2}

	fragments := sortedArtifactFragments(manager)
	if len(fragments) != 3 {
		t.Fatalf("fragment count = %d, want 3", len(fragments))
	}
	if got, want := fragments[0].ID, int32(1); got != want {
		t.Fatalf("first fragment id = %d, want %d", got, want)
	}
	if got, want := fragments[1].ID, int32(2); got != want {
		t.Fatalf("second fragment id = %d, want %d", got, want)
	}
	if got, want := fragments[2].ID, int32(3); got != want {
		t.Fatalf("third fragment id = %d, want %d", got, want)
	}
}
