package game

import "sort"

// SuggestPlays enumerates candidate card combinations from hand that
// are legal to play right now: if target is nil (free lead), any
// classifiable shape; otherwise only shapes that would beat target
// (same rules as PlayCards/HasBeatingPlay). Results are canonical --
// for shapes with interchangeable kickers (triple+1, airplane+singles,
// etc.) only one representative combination per distinct rank-run is
// returned, using the lowest available kicker cards, since kicker
// choice doesn't affect legality or strength. Results are sorted
// weakest-to-strongest (a common "play small first" heuristic), with
// bombs always sorted last since they're usually saved for real need.
func SuggestPlays(hand []Card, target *ClassifiedHand) [][]Card {
	counts := map[Rank]int{}
	byRank := map[Rank][]Card{}
	for _, c := range hand {
		counts[c.Rank]++
		byRank[c.Rank] = append(byRank[c.Rank], c)
	}

	var candidates [][]Card

	add := func(cards []Card) {
		classified, err := Classify(cards)
		if err != nil {
			return
		}
		if target != nil {
			cmp, err := Compare(classified, *target)
			if err != nil || cmp <= 0 {
				return
			}
		}
		candidates = append(candidates, cards)
	}

	// Singles, pairs, triples, bombs: one per rank.
	for r := Three; ; r++ {
		if n := counts[r]; n > 0 {
			cards := byRank[r]
			add([]Card{cards[0]})
			if n >= 2 {
				add(append([]Card{}, cards[:2]...))
			}
			if n >= 3 {
				add(append([]Card{}, cards[:3]...))
			}
			if n >= 4 {
				add(append([]Card{}, cards[:4]...))
			}
		}
		if r == Two {
			break
		}
	}

	// Triple+1 / Triple+2: for each triple rank, pair with the lowest
	// available kicker(s) from other ranks.
	for r := Three; ; r++ {
		if counts[r] >= 3 {
			triple := append([]Card{}, byRank[r][:3]...)
			if kicker := lowestKickerSingles(byRank, r, 1); kicker != nil {
				add(append(append([]Card{}, triple...), kicker...))
			}
			if kicker := lowestKickerPair(byRank, counts, r); kicker != nil {
				add(append(append([]Card{}, triple...), kicker...))
			}
		}
		if r == Two {
			break
		}
	}

	// Straights: every contiguous window of eligible ranks present,
	// length 5..maxlen.
	addRuns(counts, byRank, 1, 5, add)
	// Consecutive pairs: windows of ranks with count>=2, length 3..max.
	addRuns(counts, byRank, 2, 3, add)
	// Airplane (+ optional kickers): windows of ranks with count>=3, length 2..max.
	addAirplaneRuns(counts, byRank, add)

	sort.SliceStable(candidates, func(i, j int) bool {
		ci, _ := Classify(candidates[i])
		cj, _ := Classify(candidates[j])
		if (ci.Category == Bomb) != (cj.Category == Bomb) {
			return ci.Category != Bomb // non-bombs first
		}
		if len(candidates[i]) != len(candidates[j]) {
			return len(candidates[i]) < len(candidates[j])
		}
		return ci.Key < cj.Key
	})

	return candidates
}

func lowestKickerSingles(byRank map[Rank][]Card, exclude Rank, n int) []Card {
	var out []Card
	for r := Three; ; r++ {
		if r != exclude && len(byRank[r]) > 0 {
			out = append(out, byRank[r][0])
			if len(out) == n {
				return out
			}
		}
		if r == Two {
			break
		}
	}
	return nil
}

func lowestKickerPair(byRank map[Rank][]Card, counts map[Rank]int, exclude Rank) []Card {
	for r := Three; ; r++ {
		if r != exclude && counts[r] >= 2 {
			return append([]Card{}, byRank[r][:2]...)
		}
		if r == Two {
			break
		}
	}
	return nil
}

// addRuns finds every contiguous run of sequence-eligible ranks (each
// with count>=perGroup) of length >= minLen, and for each run length
// emits the perGroup*length lowest cards from those ranks.
func addRuns(counts map[Rank]int, byRank map[Rank][]Card, perGroup, minLen int, add func([]Card)) {
	eligible := make([]bool, int(Two)+1)
	for r := Three; r < Two; r++ {
		eligible[r] = counts[r] >= perGroup
	}

	start := -1
	flush := func(end int) {
		runLen := end - start
		if start == -1 || runLen < minLen {
			return
		}
		for length := minLen; length <= runLen; length++ {
			for winStart := start; winStart+length <= end; winStart++ {
				var cards []Card
				for r := winStart; r < winStart+length; r++ {
					cards = append(cards, byRank[Rank(r)][:perGroup]...)
				}
				add(cards)
			}
		}
	}

	for r := int(Three); r < int(Two); r++ {
		if eligible[r] {
			if start == -1 {
				start = r
			}
		} else {
			flush(r)
			start = -1
		}
	}
	flush(int(Two))
}

// addAirplaneRuns finds contiguous runs of ranks with count>=3
// (length >= 2) and emits the plain airplane plus, where available,
// one variant with n single kickers and one with n pair kickers.
func addAirplaneRuns(counts map[Rank]int, byRank map[Rank][]Card, add func([]Card)) {
	eligible := make([]bool, int(Two)+1)
	for r := Three; r < Two; r++ {
		eligible[r] = counts[r] >= 3
	}

	start := -1
	flush := func(end int) {
		runLen := end - start
		if start == -1 || runLen < 2 {
			return
		}
		for length := 2; length <= runLen; length++ {
			for winStart := start; winStart+length <= end; winStart++ {
				tripleRanks := map[Rank]bool{}
				var plain []Card
				for r := winStart; r < winStart+length; r++ {
					tripleRanks[Rank(r)] = true
					plain = append(plain, byRank[Rank(r)][:3]...)
				}
				add(append([]Card{}, plain...))

				if kickers := lowestKickersExcluding(byRank, tripleRanks, 1, length); kickers != nil {
					add(append(append([]Card{}, plain...), kickers...))
				}
				if kickers := lowestPairKickersExcluding(byRank, counts, tripleRanks, length); kickers != nil {
					add(append(append([]Card{}, plain...), kickers...))
				}
			}
		}
	}

	for r := int(Three); r < int(Two); r++ {
		if eligible[r] {
			if start == -1 {
				start = r
			}
		} else {
			flush(r)
			start = -1
		}
	}
	flush(int(Two))
}

func lowestKickersExcluding(byRank map[Rank][]Card, exclude map[Rank]bool, perRank, n int) []Card {
	var out []Card
	for r := Three; ; r++ {
		if !exclude[r] && len(byRank[r]) >= perRank {
			out = append(out, byRank[r][:perRank]...)
			if len(out) == n*perRank {
				return out
			}
		}
		if r == Two {
			break
		}
	}
	return nil
}

func lowestPairKickersExcluding(byRank map[Rank][]Card, counts map[Rank]int, exclude map[Rank]bool, n int) []Card {
	var out []Card
	for r := Three; ; r++ {
		if !exclude[r] && counts[r] >= 2 {
			out = append(out, byRank[r][:2]...)
			if len(out) == n*2 {
				return out
			}
		}
		if r == Two {
			break
		}
	}
	return nil
}
