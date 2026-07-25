package game

import (
	"errors"
	"math/rand"
)

// Phase describes the current stage of the game/match.
type Phase int

const (
	WaitingForPlayers Phase = iota
	Playing
	RoundOver
)

var (
	ErrNotYourTurn  = errors.New("game: not your turn")
	ErrGameNotReady = errors.New("game: game is not in a playable phase")
	ErrCardsNotHeld = errors.New("game: player does not hold all of the played cards")
)

// PlayerState holds one seat's hand.
type PlayerState struct {
	Hand []Card
}

// LastPlay records the most recent accepted play.
type LastPlay struct {
	PlayerIdx int
	Cards     []Card
	Hand      ClassifiedHand
}

// Game is the full two-player match state machine. All exported
// methods are safe to call concurrently IF the caller wraps access
// with its own mutex (see internal/ws.Hub) -- Game itself holds no
// lock so it can be used directly and deterministically in tests.
type Game struct {
	Phase       Phase
	Players     [2]PlayerState
	CurrentTurn int  // seat index (0 or 1) whose turn it is
	LeadPlayer  int  // seat that currently holds free lead (no lastPlay to beat)
	LastPlay    *LastPlay
	FirstRound  bool
	RoundWinner int // -1 until a round has been won
	RoundsWon   [2]int

	// LastAutoPass is set to the seat index that was just auto-passed
	// by PlayCards (because they held no beating combination), or -1
	// if the most recent PlayCards call did not trigger an auto-pass.
	// Callers (e.g. the ws Hub) should check this after each PlayCards
	// call to know whether to broadcast an additional pass_result.
	LastAutoPass int

	// LastAcceptedPlay records the play just accepted by the most
	// recent successful PlayCards call. Unlike LastPlay (which
	// autoPassIfNoBeatingPlay may immediately clear back to nil via
	// the internal Pass call), this field is only overwritten by the
	// next PlayCards call, so callers can safely read it after
	// PlayCards returns even if an auto-pass happened in the same call.
	LastAcceptedPlay *LastPlay
}

// NewGame creates a fresh match in WaitingForPlayers phase.
func NewGame() *Game {
	return &Game{
		Phase:        WaitingForPlayers,
		RoundWinner:  -1,
		LastAutoPass: -1,
	}
}

// StartRound shuffles, deals, and determines the leader for a new
// round. On the very first round, whoever holds Spade-Three leads.
// On subsequent rounds, the previous round's winner leads.
func (g *Game) StartRound(rng *rand.Rand) {
	deck := NewDeck()
	Shuffle(deck, rng)
	h0, h1 := Deal(deck)
	g.Players[0].Hand = h0
	g.Players[1].Hand = h1
	g.LastPlay = nil

	if g.RoundWinner == -1 {
		g.LeadPlayer = leaderHoldingSpadeThree(h0, h1)
	} else {
		g.LeadPlayer = g.RoundWinner
	}
	g.CurrentTurn = g.LeadPlayer
	g.Phase = Playing
}

func leaderHoldingSpadeThree(h0, h1 []Card) int {
	for _, c := range h0 {
		if c.Suit == Spade && c.Rank == Three {
			return 0
		}
	}
	for _, c := range h1 {
		if c.Suit == Spade && c.Rank == Three {
			return 1
		}
	}
	// Spade-Three was one of the removed cards in some deck variant edge
	// case; fall back to seat 0 rather than panic.
	return 0
}

// PlayCards attempts to play the given cards for playerIdx. On
// success, the cards are removed from the player's hand, LastPlay is
// updated, and the turn advances (or the round ends if the hand is
// now empty).
func (g *Game) PlayCards(playerIdx int, cards []Card) error {
	g.LastAutoPass = -1
	if g.Phase != Playing {
		return ErrGameNotReady
	}
	if playerIdx != g.CurrentTurn {
		return ErrNotYourTurn
	}
	if !handContains(g.Players[playerIdx].Hand, cards) {
		return ErrCardsNotHeld
	}

	classified, err := Classify(cards)
	if err != nil {
		return err
	}

	// If there's an active LastPlay from the OTHER player that hasn't
	// been cleared by a pass, this play must beat it.
	if g.LastPlay != nil && g.LastPlay.PlayerIdx != playerIdx {
		cmp, err := Compare(classified, g.LastPlay.Hand)
		if err != nil {
			return err // incomparable shape -> reject, caller should Pass instead
		}
		if cmp <= 0 {
			return errors.New("game: play does not beat the last play")
		}
	}

	g.Players[playerIdx].Hand = removeCards(g.Players[playerIdx].Hand, cards)
	played := &LastPlay{PlayerIdx: playerIdx, Cards: cards, Hand: classified}
	g.LastPlay = played
	g.LastAcceptedPlay = played

	if len(g.Players[playerIdx].Hand) == 0 {
		g.Phase = RoundOver
		g.RoundWinner = playerIdx
		g.RoundsWon[playerIdx]++
		return nil
	}

	g.CurrentTurn = 1 - playerIdx
	g.autoPassIfNoBeatingPlay()
	return nil
}

// autoPassIfNoBeatingPlay checks whether the player now on turn holds
// any card combination that could beat the current LastPlay; if not,
// it silently passes on their behalf so the human doesn't have to
// click Pass when it's not a real decision. Safe to call when
// LastPlay is nil (free lead) -- it's a no-op in that case.
func (g *Game) autoPassIfNoBeatingPlay() {
	if g.Phase != Playing || g.LastPlay == nil {
		return
	}
	if g.LastPlay.PlayerIdx == g.CurrentTurn {
		return // already holds free lead somehow; nothing to pass
	}
	if HasBeatingPlay(g.Players[g.CurrentTurn].Hand, g.LastPlay.Hand) {
		return
	}
	passed := g.CurrentTurn
	_ = g.Pass(passed)
	g.LastAutoPass = passed
}

// Pass gives up the current player's turn. In this two-player game a
// pass always immediately hands free lead back to whoever played
// LastPlay (there's no third player to skip to), so LastPlay is
// cleared and that player may lead with any shape next.
func (g *Game) Pass(playerIdx int) error {
	if g.Phase != Playing {
		return ErrGameNotReady
	}
	if playerIdx != g.CurrentTurn {
		return ErrNotYourTurn
	}
	if g.LastPlay == nil {
		return errors.New("game: cannot pass while holding free lead")
	}

	winner := g.LastPlay.PlayerIdx
	g.LastPlay = nil
	g.LeadPlayer = winner
	g.CurrentTurn = winner
	return nil
}

func handContains(hand []Card, cards []Card) bool {
	pool := make(map[Card]int, len(hand))
	for _, c := range hand {
		pool[c]++
	}
	for _, c := range cards {
		if pool[c] == 0 {
			return false
		}
		pool[c]--
	}
	return true
}

func removeCards(hand []Card, cards []Card) []Card {
	toRemove := make(map[Card]int, len(cards))
	for _, c := range cards {
		toRemove[c]++
	}
	out := make([]Card, 0, len(hand)-len(cards))
	for _, c := range hand {
		if toRemove[c] > 0 {
			toRemove[c]--
			continue
		}
		out = append(out, c)
	}
	return out
}
