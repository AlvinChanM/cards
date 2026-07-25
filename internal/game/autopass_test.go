package game

import "testing"

func TestHasBeatingPlaySingle(t *testing.T) {
	target := ClassifiedHand{Single, 1, int(Seven)}
	hand := []Card{c(Three, Spade), c(Nine, Heart)}
	if !HasBeatingPlay(hand, target) {
		t.Fatalf("expected Nine to beat Seven")
	}
	hand2 := []Card{c(Three, Spade), c(Five, Heart)}
	if HasBeatingPlay(hand2, target) {
		t.Fatalf("expected no card to beat Seven")
	}
}

func TestHasBeatingPlayBombAlwaysWorks(t *testing.T) {
	target := ClassifiedHand{Straight, 5, int(Nine)}
	hand := []Card{c(Three, Spade), c(Three, Heart), c(Three, Club), c(Three, Diamond)}
	if !HasBeatingPlay(hand, target) {
		t.Fatalf("expected bomb to beat straight")
	}
}

func TestHasBeatingPlayNoBombNoMatchingCategory(t *testing.T) {
	target := ClassifiedHand{Straight, 5, int(Nine)}
	hand := []Card{c(King, Spade), c(King, Heart), c(Ace, Club)}
	if HasBeatingPlay(hand, target) {
		t.Fatalf("expected no beating play (no straight, no bomb)")
	}
}

func TestHasBeatingPlayStraightHigherRun(t *testing.T) {
	target := ClassifiedHand{Straight, 5, int(Seven)} // 3-4-5-6-7
	hand := []Card{
		c(Six, Spade), c(Seven, Heart), c(Eight, Club), c(Nine, Diamond), c(Ten, Spade),
	}
	if !HasBeatingPlay(hand, target) {
		t.Fatalf("expected 6-7-8-9-10 to beat 3-4-5-6-7")
	}
}

// TestAutoPassTriggersWhenNoBeatingPlay drives PlayCards end-to-end and
// verifies the opponent is auto-passed (turn returns to the leader)
// when they hold nothing that can beat the lead.
func TestAutoPassTriggersWhenNoBeatingPlay(t *testing.T) {
	g := NewGame()
	g.Phase = Playing
	g.Players[0].Hand = []Card{c(Nine, Spade), c(King, Heart)}
	g.Players[1].Hand = []Card{c(Three, Spade), c(Four, Heart)} // nothing beats a Nine
	g.CurrentTurn = 0
	g.LeadPlayer = 0

	if err := g.PlayCards(0, []Card{c(Nine, Spade)}); err != nil {
		t.Fatalf("p0 play: %v", err)
	}

	if g.LastAutoPass != 1 {
		t.Fatalf("expected p1 to be auto-passed, LastAutoPass=%d", g.LastAutoPass)
	}
	if g.CurrentTurn != 0 || g.LeadPlayer != 0 || g.LastPlay != nil {
		t.Fatalf("expected lead to return to p0 immediately: turn=%d lead=%d lastPlay=%+v", g.CurrentTurn, g.LeadPlayer, g.LastPlay)
	}
}

func TestAutoPassDoesNotTriggerWhenBeatingPlayExists(t *testing.T) {
	g := NewGame()
	g.Phase = Playing
	g.Players[0].Hand = []Card{c(Nine, Spade), c(King, Heart)}
	g.Players[1].Hand = []Card{c(Ace, Spade), c(Four, Heart)} // Ace beats Nine
	g.CurrentTurn = 0
	g.LeadPlayer = 0

	if err := g.PlayCards(0, []Card{c(Nine, Spade)}); err != nil {
		t.Fatalf("p0 play: %v", err)
	}

	if g.LastAutoPass != -1 {
		t.Fatalf("expected no auto-pass, got LastAutoPass=%d", g.LastAutoPass)
	}
	if g.CurrentTurn != 1 {
		t.Fatalf("expected turn to remain with p1 to make their own choice, got %d", g.CurrentTurn)
	}
}
