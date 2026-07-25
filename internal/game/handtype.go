package game

import (
	"errors"
	"sort"
)

// Category identifies a played-hand shape.
type Category int

const (
	Invalid Category = iota
	Single
	Pair
	Triple
	TriplePlusOne
	TriplePlusPair
	Straight
	ConsecutivePairs
	Airplane
	AirplanePlusSingles
	AirplanePlusPairs
	Bomb
)

// ClassifiedHand is the result of classifying a played set of cards.
type ClassifiedHand struct {
	Category Category
	Length   int // number of cards
	Key      int // comparison key: the rank (as int) that determines strength within a category
}

var ErrInvalidHand = errors.New("game: invalid hand")
var ErrIncomparable = errors.New("game: hands are not comparable (different category/length)")

// Classify determines the Category of a set of played cards. cards
// need not be pre-sorted. Returns ErrInvalidHand if the cards do not
// form any recognized shape.
func Classify(cards []Card) (ClassifiedHand, error) {
	if len(cards) == 0 {
		return ClassifiedHand{}, ErrInvalidHand
	}

	counts := map[Rank]int{}
	for _, c := range cards {
		counts[c.Rank]++
	}

	switch len(cards) {
	case 1:
		return ClassifiedHand{Single, 1, int(cards[0].Rank)}, nil
	case 2:
		if r, ok := singleGroupRank(counts, 2); ok {
			return ClassifiedHand{Pair, 2, int(r)}, nil
		}
	case 3:
		if r, ok := singleGroupRank(counts, 3); ok {
			return ClassifiedHand{Triple, 3, int(r)}, nil
		}
	case 4:
		if r, ok := singleGroupRank(counts, 4); ok {
			return ClassifiedHand{Bomb, 4, int(r)}, nil
		}
		if r, ok := tripleWithKickers(counts, 1, 1); ok {
			return ClassifiedHand{TriplePlusOne, 4, int(r)}, nil
		}
	case 5:
		if r, ok := tripleWithKickers(counts, 1, 2); ok {
			return ClassifiedHand{TriplePlusPair, 5, int(r)}, nil
		}
		if r, ok := runOfRank(counts, 5, 1); ok {
			return ClassifiedHand{Straight, 5, int(r)}, nil
		}
	default:
		// consecutive pairs: all counts == 2, len even, >=6, contiguous ranks
		if r, ok := runOfRank(counts, len(cards)/2, 2); ok && len(cards)%2 == 0 && len(cards) >= 6 {
			return ClassifiedHand{ConsecutivePairs, len(cards), int(r)}, nil
		}
		// straight: all counts == 1, len >= 5, contiguous ranks
		if r, ok := runOfRank(counts, len(cards), 1); ok && len(cards) >= 5 {
			return ClassifiedHand{Straight, len(cards), int(r)}, nil
		}
		// airplane (+ optional kickers): n>=2 contiguous triples, plus n singles or n pairs (or none)
		if ch, ok := classifyAirplane(counts, len(cards)); ok {
			return ch, nil
		}
	}

	return ClassifiedHand{}, ErrInvalidHand
}

// singleGroupRank checks that all cards belong to exactly one rank
// group of the given size (used for pair/triple/bomb: n==size).
func singleGroupRank(counts map[Rank]int, size int) (Rank, bool) {
	if len(counts) != 1 {
		return 0, false
	}
	for r, c := range counts {
		if c == size {
			return r, true
		}
	}
	return 0, false
}

// tripleWithKickers checks for exactly one triple plus kicker(s) of
// the given count/size (kickerCount groups each of kickerSize cards,
// e.g. 1 single kicker, or 2 cards forming 1 pair kicker). Kickers
// need not match each other in rank and may include Two.
func tripleWithKickers(counts map[Rank]int, tripleCountWanted, kickerCards int) (Rank, bool) {
	var tripleRank Rank
	found := false
	kickerTotal := 0
	for r, c := range counts {
		if c == 3 && !found {
			tripleRank = r
			found = true
			continue
		}
		kickerTotal += c
		// a kicker "pair" (for triple+pair) must be an actual pair (c==2),
		// not two singles of different ranks, when kickerCards==2.
		if kickerCards == 2 && c != 2 {
			return 0, false
		}
	}
	if !found || kickerTotal != kickerCards {
		return 0, false
	}
	return tripleRank, true
}

// runOfRank checks that counts describes exactly `groups` distinct
// ranks, each appearing `perGroup` times, forming a contiguous
// sequence eligible for runs (excludes Two). Returns the highest
// rank in the run.
func runOfRank(counts map[Rank]int, groups, perGroup int) (Rank, bool) {
	if len(counts) != groups {
		return 0, false
	}
	ranks := make([]int, 0, groups)
	for r, c := range counts {
		if c != perGroup {
			return 0, false
		}
		if !r.IsSequenceEligible() {
			return 0, false
		}
		ranks = append(ranks, int(r))
	}
	sort.Ints(ranks)
	for i := 1; i < len(ranks); i++ {
		if ranks[i] != ranks[i-1]+1 {
			return 0, false
		}
	}
	return Rank(ranks[len(ranks)-1]), true
}

// classifyAirplane attempts to find n>=2 contiguous triples within
// counts, with the remaining cards (if any) forming either n single
// kickers or n pair kickers.
func classifyAirplane(counts map[Rank]int, totalLen int) (ClassifiedHand, bool) {
	var tripleRanks []int
	leftoverCards := 0
	for r, c := range counts {
		switch {
		case c == 3:
			if !r.IsSequenceEligible() {
				return ClassifiedHand{}, false
			}
			tripleRanks = append(tripleRanks, int(r))
		case c > 0:
			leftoverCards += c
		}
	}
	n := len(tripleRanks)
	if n < 2 {
		return ClassifiedHand{}, false
	}
	sort.Ints(tripleRanks)
	for i := 1; i < n; i++ {
		if tripleRanks[i] != tripleRanks[i-1]+1 {
			return ClassifiedHand{}, false
		}
	}
	highest := tripleRanks[n-1]

	switch leftoverCards {
	case 0:
		if totalLen != 3*n {
			return ClassifiedHand{}, false
		}
		return ClassifiedHand{Airplane, totalLen, highest}, true
	case n:
		// each leftover rank must appear exactly once (n single kickers)
		return ClassifiedHand{AirplanePlusSingles, totalLen, highest}, true
	case 2 * n:
		// leftover must be entirely pairs: every non-triple rank present has count==2
		for r, c := range counts {
			if c == 3 {
				continue
			}
			if c != 2 {
				return ClassifiedHand{}, false
			}
			_ = r
		}
		return ClassifiedHand{AirplanePlusPairs, totalLen, highest}, true
	default:
		return ClassifiedHand{}, false
	}
}

// Compare returns -1, 0, or 1 if a is weaker than, equal to, or
// stronger than b. A Bomb beats any non-bomb regardless of length.
// Two bombs compare by Key. Otherwise a and b must share Category and
// Length or ErrIncomparable is returned (caller should treat this as
// "cannot follow, must Pass").
func Compare(a, b ClassifiedHand) (int, error) {
	if a.Category == Bomb && b.Category != Bomb {
		return 1, nil
	}
	if b.Category == Bomb && a.Category != Bomb {
		return -1, nil
	}
	if a.Category != b.Category || a.Length != b.Length {
		return 0, ErrIncomparable
	}
	switch {
	case a.Key > b.Key:
		return 1, nil
	case a.Key < b.Key:
		return -1, nil
	default:
		return 0, nil
	}
}
