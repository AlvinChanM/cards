// Package game implements the core rules engine for 二人关牌（跑得快），
// a two-player Chinese shedding card game. This package has no
// networking dependencies and can be exercised entirely via unit tests.
package game

import "fmt"

// Suit identifies one of the four standard suits.
type Suit uint8

const (
	Spade Suit = iota
	Heart
	Club
	Diamond
)

func (s Suit) String() string {
	switch s {
	case Spade:
		return "S"
	case Heart:
		return "H"
	case Club:
		return "C"
	case Diamond:
		return "D"
	default:
		return "?"
	}
}

// Rank identifies a card's point value. Values are chosen so that
// plain integer comparison matches game ordering: 3 is lowest, 2 is
// highest. Sequences (straights/consecutive pairs/airplanes) may
// never include Two.
type Rank uint8

const (
	Three Rank = iota
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace
	Two
)

var rankNames = map[Rank]string{
	Three: "3", Four: "4", Five: "5", Six: "6", Seven: "7",
	Eight: "8", Nine: "9", Ten: "10", Jack: "J", Queen: "Q",
	King: "K", Ace: "A", Two: "2",
}

func (r Rank) String() string {
	if n, ok := rankNames[r]; ok {
		return n
	}
	return "?"
}

// IsSequenceEligible reports whether this rank may participate in a
// straight, consecutive-pairs, or airplane run. Two is excluded per
// standard 跑得快 rules.
func (r Rank) IsSequenceEligible() bool {
	return r != Two
}

// Card is a single playing card.
type Card struct {
	Suit Suit
	Rank Rank
}

func (c Card) String() string {
	return fmt.Sprintf("%s%s", c.Rank, c.Suit)
}

// SortCards sorts a slice of cards ascending by rank (suit is not
// game-relevant for ordering but is kept stable for determinism).
func SortCards(cards []Card) {
	// simple insertion sort; hands are small (<=24) so this is plenty fast
	for i := 1; i < len(cards); i++ {
		for j := i; j > 0 && cards[j-1].Rank > cards[j].Rank; j-- {
			cards[j-1], cards[j] = cards[j], cards[j-1]
		}
	}
}
