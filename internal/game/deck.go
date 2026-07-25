package game

import "math/rand"

// NewDeck builds the 48-card deck used by 二人关牌.
//
// Composition (confirmed with the user):
//   - Ranks Three..King (11 ranks): all 4 suits kept  -> 11*4 = 44 cards
//   - Ace: only 3 copies kept (one suit's ace removed) -> 3 cards
//   - Two: only 1 copy kept                            -> 1 card
//   - No jokers.
//
// Total: 44 + 3 + 1 = 48 cards, which splits evenly into two 24-card
// hands with no kitty/bottom cards.
func NewDeck() []Card {
	deck := make([]Card, 0, 48)
	suits := []Suit{Spade, Heart, Club, Diamond}

	for r := Three; r <= King; r++ {
		for _, s := range suits {
			deck = append(deck, Card{Suit: s, Rank: r})
		}
	}

	// Ace: keep 3 of 4 suits (drop Diamond arbitrarily).
	for _, s := range []Suit{Spade, Heart, Club} {
		deck = append(deck, Card{Suit: s, Rank: Ace})
	}

	// Two: keep exactly 1 copy. Spade is chosen so this card is
	// visually distinct; the specific suit carries no rules meaning
	// since Two never participates in sequences.
	deck = append(deck, Card{Suit: Spade, Rank: Two})

	return deck
}

// Shuffle randomizes deck order in place using the provided source.
// Pass a seeded rand.Rand in tests for determinism.
func Shuffle(deck []Card, rng *rand.Rand) {
	rng.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

// Deal splits a 48-card deck into two 24-card hands, dealt alternately.
func Deal(deck []Card) (hand0, hand1 []Card) {
	if len(deck) != 48 {
		panic("game: Deal requires exactly 48 cards")
	}
	hand0 = make([]Card, 0, 24)
	hand1 = make([]Card, 0, 24)
	for i, c := range deck {
		if i%2 == 0 {
			hand0 = append(hand0, c)
		} else {
			hand1 = append(hand1, c)
		}
	}
	SortCards(hand0)
	SortCards(hand1)
	return hand0, hand1
}
