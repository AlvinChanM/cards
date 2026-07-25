package game

// HasBeatingPlay reports whether hand contains any combination of
// cards that can legally beat target (same rules as Compare/PlayCards
// would accept). Used to auto-pass a player who has no legal response,
// rather than making them click Pass manually.
func HasBeatingPlay(hand []Card, target ClassifiedHand) bool {
	counts := map[Rank]int{}
	for _, c := range hand {
		counts[c.Rank]++
	}
	total := len(hand)

	// A bomb beats anything non-bomb; check first regardless of target
	// category (skip if target itself is a bomb -- handled below).
	if target.Category != Bomb {
		for r, n := range counts {
			if n == 4 {
				return true
			}
			_ = r
		}
	}

	switch target.Category {
	case Bomb:
		for r, n := range counts {
			if n == 4 && int(r) > target.Key {
				return true
			}
		}
		return false

	case Single:
		for r, n := range counts {
			if n >= 1 && int(r) > target.Key {
				return true
			}
		}
		return false

	case Pair:
		for r, n := range counts {
			if n >= 2 && int(r) > target.Key {
				return true
			}
		}
		return false

	case Triple:
		for r, n := range counts {
			if n >= 3 && int(r) > target.Key {
				return true
			}
		}
		return false

	case TriplePlusOne:
		for r, n := range counts {
			if n >= 3 && int(r) > target.Key && total-3 >= 1 {
				return true
			}
		}
		return false

	case TriplePlusPair:
		for r, n := range counts {
			if n < 3 || int(r) <= target.Key {
				continue
			}
			for s, m := range counts {
				if s != r && m >= 2 {
					return true
				}
			}
		}
		return false

	case Straight:
		return hasConsecutiveRun(counts, target.Length, 1, target.Key)

	case ConsecutivePairs:
		return hasConsecutiveRun(counts, target.Length/2, 2, target.Key)

	case Airplane:
		return hasConsecutiveTripleRun(counts, target.Length/3, target.Key, 0)

	case AirplanePlusSingles:
		n := target.Length / 4
		return hasConsecutiveTripleRun(counts, n, target.Key, n)

	case AirplanePlusPairs:
		n := target.Length / 5
		return hasConsecutiveTripleRunWithPairKickers(counts, n, target.Key)

	default:
		return false
	}
}

// hasConsecutiveRun checks for `groups` consecutive sequence-eligible
// ranks, each with count >= perGroup, whose highest rank exceeds
// minKey.
func hasConsecutiveRun(counts map[Rank]int, groups, perGroup, minKey int) bool {
	if groups <= 0 {
		return false
	}
	for start := int(Three); start+groups-1 <= int(King)+1; start++ {
		ok := true
		for i := 0; i < groups; i++ {
			r := Rank(start + i)
			if !r.IsSequenceEligible() || counts[r] < perGroup {
				ok = false
				break
			}
		}
		if ok && start+groups-1 > minKey {
			return true
		}
	}
	return false
}

// hasConsecutiveTripleRun checks for `n` consecutive ranks with
// count>=3 (highest > minKey), plus `kickers` additional leftover
// cards of any rank (kickers==0 for plain Airplane).
func hasConsecutiveTripleRun(counts map[Rank]int, n, minKey, kickers int) bool {
	if n <= 0 {
		return false
	}
	total := 0
	for _, c := range counts {
		total += c
	}
	for start := int(Three); start+n-1 <= int(King)+1; start++ {
		ok := true
		for i := 0; i < n; i++ {
			r := Rank(start + i)
			if !r.IsSequenceEligible() || counts[r] < 3 {
				ok = false
				break
			}
		}
		if ok && start+n-1 > minKey && total-3*n >= kickers {
			return true
		}
	}
	return false
}

// hasConsecutiveTripleRunWithPairKickers is like hasConsecutiveTripleRun
// but requires the leftover kickers to come from n distinct ranks each
// contributing a pair (approximation: n other ranks with count>=2).
func hasConsecutiveTripleRunWithPairKickers(counts map[Rank]int, n, minKey int) bool {
	if n <= 0 {
		return false
	}
	for start := int(Three); start+n-1 <= int(King)+1; start++ {
		tripleRanks := map[Rank]bool{}
		ok := true
		for i := 0; i < n; i++ {
			r := Rank(start + i)
			if !r.IsSequenceEligible() || counts[r] < 3 {
				ok = false
				break
			}
			tripleRanks[r] = true
		}
		if !ok || start+n-1 <= minKey {
			continue
		}
		pairSources := 0
		for r, c := range counts {
			if tripleRanks[r] {
				continue
			}
			if c >= 2 {
				pairSources++
			}
		}
		if pairSources >= n {
			return true
		}
	}
	return false
}
