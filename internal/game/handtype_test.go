package game

import "testing"

func c(rank Rank, suit Suit) Card { return Card{Suit: suit, Rank: rank} }

func TestClassify(t *testing.T) {
	tests := []struct {
		name     string
		cards    []Card
		wantCat  Category
		wantKey  int
		wantErr  bool
	}{
		{"single", []Card{c(Five, Spade)}, Single, int(Five), false},
		{"single two allowed", []Card{c(Two, Spade)}, Single, int(Two), false},
		{"pair", []Card{c(Seven, Spade), c(Seven, Heart)}, Pair, int(Seven), false},
		{"pair mismatched ranks invalid", []Card{c(Seven, Spade), c(Eight, Heart)}, Invalid, 0, true},
		{"triple", []Card{c(Nine, Spade), c(Nine, Heart), c(Nine, Club)}, Triple, int(Nine), false},
		{"bomb", []Card{c(Jack, Spade), c(Jack, Heart), c(Jack, Club), c(Jack, Diamond)}, Bomb, int(Jack), false},
		{"triple+1", []Card{c(Nine, Spade), c(Nine, Heart), c(Nine, Club), c(King, Spade)}, TriplePlusOne, int(Nine), false},
		{"triple+1 with two kicker ok", []Card{c(Nine, Spade), c(Nine, Heart), c(Nine, Club), c(Two, Spade)}, TriplePlusOne, int(Nine), false},
		{"4 of a kind not triple+1 (bomb wins)", []Card{c(Nine, Spade), c(Nine, Heart), c(Nine, Club), c(Nine, Diamond)}, Bomb, int(Nine), false},
		{"triple+pair", []Card{c(Nine, Spade), c(Nine, Heart), c(Nine, Club), c(King, Spade), c(King, Heart)}, TriplePlusPair, int(Nine), false},
		{"triple+2 singles invalid (must be pair)", []Card{c(Nine, Spade), c(Nine, Heart), c(Nine, Club), c(King, Spade), c(Queen, Heart)}, Invalid, 0, true},
		{"straight min5", []Card{c(Three, Spade), c(Four, Spade), c(Five, Spade), c(Six, Spade), c(Seven, Spade)}, Straight, int(Seven), false},
		{"straight 4 too short", []Card{c(Three, Spade), c(Four, Spade), c(Five, Spade), c(Six, Spade)}, Invalid, 0, true},
		{"straight cannot include two", []Card{c(King, Spade), c(Ace, Spade), c(Two, Heart), c(Three, Club), c(Four, Diamond)}, Invalid, 0, true},
		{"straight non-contiguous invalid", []Card{c(Three, Spade), c(Four, Spade), c(Six, Spade), c(Seven, Spade), c(Eight, Spade)}, Invalid, 0, true},
		{"consecutive pairs min3 (6 cards)", []Card{
			c(Three, Spade), c(Three, Heart), c(Four, Spade), c(Four, Heart), c(Five, Spade), c(Five, Heart),
		}, ConsecutivePairs, int(Five), false},
		{"consecutive pairs 2 pairs too short", []Card{
			c(Three, Spade), c(Three, Heart), c(Four, Spade), c(Four, Heart),
		}, Invalid, 0, true},
		{"consecutive pairs cannot include two", []Card{
			c(Ace, Spade), c(Ace, Heart), c(Two, Spade), c(Two, Heart),
		}, Invalid, 0, true},
		{"airplane min2 triples", []Card{
			c(Three, Spade), c(Three, Heart), c(Three, Club),
			c(Four, Spade), c(Four, Heart), c(Four, Club),
		}, Airplane, int(Four), false},
		{"airplane single triple invalid", []Card{
			c(Three, Spade), c(Three, Heart), c(Three, Club),
		}, Triple, int(Three), false}, // len 3 dispatches to Triple case, correctly not Airplane
		{"airplane with two-triple invalid (2 breaks sequence)", []Card{
			c(Two, Spade), c(Two, Heart), c(Two, Club),
			c(Three, Spade), c(Three, Heart), c(Three, Club),
		}, Invalid, 0, true},
		{"airplane plus singles", []Card{
			c(Three, Spade), c(Three, Heart), c(Three, Club),
			c(Four, Spade), c(Four, Heart), c(Four, Club),
			c(Nine, Spade), c(King, Heart),
		}, AirplanePlusSingles, int(Four), false},
		{"airplane plus pairs", []Card{
			c(Three, Spade), c(Three, Heart), c(Three, Club),
			c(Four, Spade), c(Four, Heart), c(Four, Club),
			c(Nine, Spade), c(Nine, Heart), c(King, Club), c(King, Diamond),
		}, AirplanePlusPairs, int(Four), false},
		{"empty invalid", []Card{}, Invalid, 0, true},
		{"random junk invalid", []Card{c(Three, Spade), c(Five, Heart), c(Nine, Club), c(King, Diamond), c(Two, Spade), c(Four, Heart)}, Invalid, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Classify(tt.cards)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Category != tt.wantCat {
				t.Errorf("category = %v, want %v", got.Category, tt.wantCat)
			}
			if got.Key != tt.wantKey {
				t.Errorf("key = %v, want %v", got.Key, tt.wantKey)
			}
			if got.Length != len(tt.cards) {
				t.Errorf("length = %v, want %v", got.Length, len(tt.cards))
			}
		})
	}
}

func TestCompare(t *testing.T) {
	single5 := ClassifiedHand{Single, 1, int(Five)}
	single9 := ClassifiedHand{Single, 1, int(Nine)}
	pair5 := ClassifiedHand{Pair, 2, int(Five)}
	bombJ := ClassifiedHand{Bomb, 4, int(Jack)}
	bombK := ClassifiedHand{Bomb, 4, int(King)}

	cases := []struct {
		name    string
		a, b    ClassifiedHand
		want    int
		wantErr bool
	}{
		{"bigger single wins", single9, single5, 1, false},
		{"smaller single loses", single5, single9, -1, false},
		{"equal", single5, single5, 0, false},
		{"bomb beats single", bombJ, single9, 1, false},
		{"single loses to bomb", single9, bombJ, -1, false},
		{"bigger bomb wins", bombK, bombJ, 1, false},
		{"mismatched category incomparable", single5, pair5, 0, true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Compare(tt.a, tt.b)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("compare = %v, want %v", got, tt.want)
			}
		})
	}
}

func FuzzClassify(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4})
	f.Fuzz(func(t *testing.T, data []byte) {
		cards := make([]Card, 0, len(data))
		for _, b := range data {
			cards = append(cards, Card{Suit: Suit(b % 4), Rank: Rank(b % 13)})
		}
		_, _ = Classify(cards) // must not panic
	})
}
