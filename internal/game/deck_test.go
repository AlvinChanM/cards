package game

import "testing"

func TestDeckComposition(t *testing.T) {
	deck := NewDeck()
	if len(deck) != 48 {
		t.Fatalf("deck size = %d, want 48", len(deck))
	}
	counts := map[Rank]int{}
	for _, c := range deck {
		counts[c.Rank]++
	}
	for r := Three; r <= King; r++ {
		if counts[r] != 4 {
			t.Errorf("rank %v count = %d, want 4", r, counts[r])
		}
	}
	if counts[Ace] != 3 {
		t.Errorf("Ace count = %d, want 3", counts[Ace])
	}
	if counts[Two] != 1 {
		t.Errorf("Two count = %d, want 1", counts[Two])
	}
}

func TestDealSplitsEvenly(t *testing.T) {
	deck := NewDeck()
	h0, h1 := Deal(deck)
	if len(h0) != 24 || len(h1) != 24 {
		t.Fatalf("hand sizes = %d/%d, want 24/24", len(h0), len(h1))
	}
}
