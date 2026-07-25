// Package ws implements the WebSocket transport layer for the single
// global 二人关牌 match. It has no game-rule logic of its own -- all
// rules enforcement is delegated to internal/game.Game.
package ws

import "encoding/json"

// Envelope is the single message wrapper used in both directions.
type Envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func encode(msgType string, data any) ([]byte, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{Type: msgType, Data: raw})
}

func unmarshalData(raw json.RawMessage, v any) error {
	return json.Unmarshal(raw, v)
}

// --- Client -> Server message bodies ---

// JoinRequest is sent immediately after connecting.
type JoinRequest struct {
	Name string `json:"name"`
}

// PlayCardsRequest asks the server to play the given cards.
type PlayCardsRequest struct {
	Cards []CardDTO `json:"cards"`
}

// PassRequest asks the server to pass the current turn.
type PassRequest struct{}

// RequestStateRequest asks the server to resend a full state snapshot.
type RequestStateRequest struct{}

// RequestHintRequest asks the server for the list of legal plays
// given the player's current hand and the play they'd need to beat
// (or any shape, if they hold free lead).
type RequestHintRequest struct{}

// --- Server -> Client message bodies ---

// CardDTO is the wire representation of a game.Card.
type CardDTO struct {
	Suit string `json:"suit"` // "S","H","C","D"
	Rank string `json:"rank"` // "3".."10","J","Q","K","A","2"
}

// JoinedMessage confirms a successful join and seat assignment.
type JoinedMessage struct {
	Seat        string   `json:"seat"` // "0", "1", or "spectator"
	PlayerNames [2]string `json:"playerNames"`
}

// LastPlayDTO describes the most recent accepted play, if any.
type LastPlayDTO struct {
	PlayerIdx int       `json:"playerIdx"`
	Cards     []CardDTO `json:"cards"`
	Category  string    `json:"category"`
}

// StateSyncMessage is a full, per-recipient-redacted state snapshot.
type StateSyncMessage struct {
	Phase             string       `json:"phase"`
	YourSeat          int          `json:"yourSeat"` // -1 for spectators
	YourHand          []CardDTO    `json:"yourHand,omitempty"`
	OpponentCardCount int          `json:"opponentCardCount"`
	CurrentTurn       int          `json:"currentTurn"`
	LeadPlayer        int          `json:"leadPlayer"`
	LastPlay          *LastPlayDTO `json:"lastPlay"`
	RoundsWon         [2]int       `json:"roundsWon"`
	PlayerNames       [2]string    `json:"playerNames"`
}

// PlayResultMessage is broadcast after a play is accepted.
type PlayResultMessage struct {
	PlayerIdx int       `json:"playerIdx"`
	Cards     []CardDTO `json:"cards"`
	Category  string    `json:"category"`
}

// PassResultMessage is broadcast after a pass is accepted.
type PassResultMessage struct {
	PlayerIdx int `json:"playerIdx"`
}

// RoundOverMessage announces a round's winner.
type RoundOverMessage struct {
	Winner    int    `json:"winner"`
	RoundsWon [2]int `json:"roundsWon"`
}

// WaitingMessage indicates the match is waiting for a second player.
type WaitingMessage struct {
	PlayersConnected int `json:"playersConnected"`
}

// ErrorMessage reports a rejected client action.
type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// HintMessage lists the legal plays available to the requesting
// player right now, weakest-to-strongest (bombs last). Empty if the
// player holds nothing that can legally be played (i.e. they must
// pass, or the server will auto-pass them on the opponent's next
// move).
type HintMessage struct {
	Plays [][]CardDTO `json:"plays"`
}

const (
	TypeJoin         = "join"
	TypePlayCards    = "play_cards"
	TypePass         = "pass"
	TypeRequestState = "request_state"
	TypeRequestHint  = "request_hint"

	TypeJoined     = "joined"
	TypeStateSync  = "state_sync"
	TypePlayResult = "play_result"
	TypePassResult = "pass_result"
	TypeRoundOver  = "round_over"
	TypeWaiting    = "waiting_for_opponent"
	TypeError      = "error"
	TypeHint       = "hint"
)
