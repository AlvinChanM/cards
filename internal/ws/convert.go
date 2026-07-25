package ws

import (
	"fmt"

	"github.com/cards/internal/game"
)

var suitToStr = map[game.Suit]string{
	game.Spade: "S", game.Heart: "H", game.Club: "C", game.Diamond: "D",
}
var strToSuit = map[string]game.Suit{
	"S": game.Spade, "H": game.Heart, "C": game.Club, "D": game.Diamond,
}

var rankToStr = map[game.Rank]string{
	game.Three: "3", game.Four: "4", game.Five: "5", game.Six: "6",
	game.Seven: "7", game.Eight: "8", game.Nine: "9", game.Ten: "10",
	game.Jack: "J", game.Queen: "Q", game.King: "K", game.Ace: "A", game.Two: "2",
}
var strToRank = map[string]game.Rank{
	"3": game.Three, "4": game.Four, "5": game.Five, "6": game.Six,
	"7": game.Seven, "8": game.Eight, "9": game.Nine, "10": game.Ten,
	"J": game.Jack, "Q": game.Queen, "K": game.King, "A": game.Ace, "2": game.Two,
}

var categoryToStr = map[game.Category]string{
	game.Invalid:             "invalid",
	game.Single:              "single",
	game.Pair:                "pair",
	game.Triple:              "triple",
	game.TriplePlusOne:       "triple_plus_one",
	game.TriplePlusPair:      "triple_plus_pair",
	game.Straight:            "straight",
	game.ConsecutivePairs:    "consecutive_pairs",
	game.Airplane:            "airplane",
	game.AirplanePlusSingles: "airplane_plus_singles",
	game.AirplanePlusPairs:   "airplane_plus_pairs",
	game.Bomb:                "bomb",
}

func cardToDTO(c game.Card) CardDTO {
	return CardDTO{Suit: suitToStr[c.Suit], Rank: rankToStr[c.Rank]}
}

func cardsToDTO(cards []game.Card) []CardDTO {
	out := make([]CardDTO, len(cards))
	for i, c := range cards {
		out[i] = cardToDTO(c)
	}
	return out
}

func dtoToCard(d CardDTO) (game.Card, error) {
	s, ok := strToSuit[d.Suit]
	if !ok {
		return game.Card{}, fmt.Errorf("ws: unknown suit %q", d.Suit)
	}
	r, ok := strToRank[d.Rank]
	if !ok {
		return game.Card{}, fmt.Errorf("ws: unknown rank %q", d.Rank)
	}
	return game.Card{Suit: s, Rank: r}, nil
}

func dtosToCards(dtos []CardDTO) ([]game.Card, error) {
	out := make([]game.Card, len(dtos))
	for i, d := range dtos {
		c, err := dtoToCard(d)
		if err != nil {
			return nil, err
		}
		out[i] = c
	}
	return out, nil
}

func phaseToStr(p game.Phase) string {
	switch p {
	case game.WaitingForPlayers:
		return "waiting"
	case game.Playing:
		return "playing"
	case game.RoundOver:
		return "round_over"
	default:
		return "unknown"
	}
}
