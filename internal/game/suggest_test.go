package game

import "testing"

func TestSuggestPlaysFreeLead(t *testing.T) {
	hand := []Card{c(Three, Spade), c(Three, Heart), c(Four, Spade)}
	plays := SuggestPlays(hand, nil)
	if len(plays) == 0 {
		t.Fatalf("expected at least one suggested play")
	}
	// Must include at least: single 3, single 4, pair of 3s.
	foundSingle3, foundSingle4, foundPair3 := false, false, false
	for _, p := range plays {
		cl, err := Classify(p)
		if err != nil {
			t.Fatalf("suggested play failed to classify: %v -> %v", p, err)
		}
		if cl.Category == Single && cl.Key == int(Three) {
			foundSingle3 = true
		}
		if cl.Category == Single && cl.Key == int(Four) {
			foundSingle4 = true
		}
		if cl.Category == Pair && cl.Key == int(Three) {
			foundPair3 = true
		}
	}
	if !foundSingle3 || !foundSingle4 || !foundPair3 {
		t.Fatalf("missing expected candidates: single3=%v single4=%v pair3=%v", foundSingle3, foundSingle4, foundPair3)
	}
}

func TestSuggestPlaysMustBeatTarget(t *testing.T) {
	hand := []Card{c(Three, Spade), c(Nine, Heart), c(King, Spade)}
	target := ClassifiedHand{Single, 1, int(Seven)}
	plays := SuggestPlays(hand, &target)
	for _, p := range plays {
		cl, _ := Classify(p)
		if cl.Category != Bomb && cl.Key <= target.Key {
			t.Fatalf("suggested play %v (key=%d) does not beat target (key=%d)", p, cl.Key, target.Key)
		}
	}
	// Three (rank < Seven) must never appear as a suggested single.
	for _, p := range plays {
		if len(p) == 1 && p[0].Rank == Three {
			t.Fatalf("suggested a Three single against a target of Seven")
		}
	}
	// Nine and King should each appear as viable singles.
	foundNine, foundKing := false, false
	for _, p := range plays {
		if len(p) == 1 && p[0].Rank == Nine {
			foundNine = true
		}
		if len(p) == 1 && p[0].Rank == King {
			foundKing = true
		}
	}
	if !foundNine || !foundKing {
		t.Fatalf("expected Nine and King singles to beat Seven: nine=%v king=%v", foundNine, foundKing)
	}
}

func TestSuggestPlaysBombAlwaysOffered(t *testing.T) {
	hand := []Card{c(Five, Spade), c(Five, Heart), c(Five, Club), c(Five, Diamond)}
	target := ClassifiedHand{Straight, 5, int(King)}
	plays := SuggestPlays(hand, &target)
	found := false
	for _, p := range plays {
		cl, _ := Classify(p)
		if cl.Category == Bomb {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected bomb to be suggested against an unbeatable-by-category straight")
	}
}

func TestSuggestPlaysSortedWeakestFirstBombsLast(t *testing.T) {
	hand := []Card{
		c(Three, Spade), c(Four, Spade),
		c(Five, Spade), c(Five, Heart), c(Five, Club), c(Five, Diamond),
	}
	plays := SuggestPlays(hand, nil)
	sawBomb := false
	for _, p := range plays {
		cl, _ := Classify(p)
		if cl.Category == Bomb {
			sawBomb = true
			continue
		}
		if sawBomb {
			t.Fatalf("found non-bomb play after a bomb in sorted order: %v", p)
		}
	}
}

func TestSuggestPlaysNoLegalMoves(t *testing.T) {
	hand := []Card{c(Three, Spade)}
	target := ClassifiedHand{Single, 1, int(Two)} // nothing beats the highest single
	plays := SuggestPlays(hand, &target)
	if len(plays) != 0 {
		t.Fatalf("expected no suggestions when nothing beats target, got %v", plays)
	}
}

func TestSuggestPlaysStraightAndConsecutivePairs(t *testing.T) {
	hand := []Card{
		c(Three, Spade), c(Three, Heart),
		c(Four, Spade), c(Four, Heart),
		c(Five, Spade), c(Five, Heart),
		c(Six, Spade), c(Seven, Heart),
	}
	plays := SuggestPlays(hand, nil)
	foundStraight, foundConsecPairs := false, false
	for _, p := range plays {
		cl, _ := Classify(p)
		if cl.Category == Straight && len(p) == 5 {
			foundStraight = true
		}
		if cl.Category == ConsecutivePairs && len(p) == 6 {
			foundConsecPairs = true
		}
	}
	if !foundStraight {
		t.Fatalf("expected a 5-card straight suggestion (3-4-5-6-7)")
	}
	if !foundConsecPairs {
		t.Fatalf("expected consecutive pairs 3-3-4-4-5-5 suggestion")
	}
}
