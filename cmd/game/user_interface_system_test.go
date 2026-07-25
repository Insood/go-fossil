package main

import "testing"

func TestSortedArtifactFragmentsOrdersByID(t *testing.T) {
	t.Parallel()

	manager := NewArtifactManager()
	manager.fragments[3] = &ArtifactFragment{ID: 3, Collected: true}
	manager.fragments[1] = &ArtifactFragment{ID: 1, Collected: true}
	manager.fragments[2] = &ArtifactFragment{ID: 2, Collected: true}
	manager.fragments[4] = &ArtifactFragment{ID: 4}

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

func TestRecentArtifactFragmentsReturnsNewestFirstWithLimit(t *testing.T) {
	t.Parallel()

	manager := NewArtifactManager()
	for id := int32(1); id <= 10; id++ {
		manager.fragments[id] = &ArtifactFragment{ID: id, Collected: true}
	}
	manager.fragments[11] = &ArtifactFragment{ID: 11}

	fragments := recentArtifactFragments(manager, 8)
	if len(fragments) != 8 {
		t.Fatalf("fragment count = %d, want 8", len(fragments))
	}

	want := []int32{10, 9, 8, 7, 6, 5, 4, 3}
	for i, fragment := range fragments {
		if got, wantID := fragment.ID, want[i]; got != wantID {
			t.Fatalf("fragment %d id = %d, want %d", i, got, wantID)
		}
	}
}
