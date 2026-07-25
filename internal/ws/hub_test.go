package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gws "github.com/gorilla/websocket"

	"github.com/cards/internal/game"
)

// TestBombPlayWithSubsequentAutoPass reproduces a real bug: playing a
// bomb that triggers auto-pass for the opponent used to crash the
// server (nil pointer dereference reading h.g.LastPlay.Hand.Category
// after autoPassIfNoBeatingPlay had already cleared LastPlay to nil).
// This drives the full real WS path (not just the game package) since
// the bug lived entirely in the ws.Hub layer.
func TestBombPlayWithSubsequentAutoPass(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWS(hub, w, r)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	p0, _, err := gws.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial p0: %v", err)
	}
	defer p0.Close()
	drainUntil(t, p0, TypeJoined, 2*time.Second)
	drainUntil(t, p0, TypeWaiting, 2*time.Second)

	p1, _, err := gws.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial p1: %v", err)
	}
	defer p1.Close()
	drainUntil(t, p1, TypeJoined, 2*time.Second)

	drainUntil(t, p0, TypeStateSync, 2*time.Second)
	drainUntil(t, p1, TypeStateSync, 2*time.Second)

	// p0 leads with consecutive pairs 4-4-5-5-6-6-7-7; p1 holds a bomb
	// (four Queens) that beats it, and after p1 plays the bomb, p0's
	// remaining single card (a Nine) cannot beat a bomb, so p0 should
	// be auto-passed -- this is exactly the sequence that used to crash
	// the server.
	hub.mu.Lock()
	hub.g.Players[0].Hand = []game.Card{
		{Suit: game.Spade, Rank: game.Four}, {Suit: game.Heart, Rank: game.Four},
		{Suit: game.Spade, Rank: game.Five}, {Suit: game.Heart, Rank: game.Five},
		{Suit: game.Spade, Rank: game.Six}, {Suit: game.Heart, Rank: game.Six},
		{Suit: game.Spade, Rank: game.Seven}, {Suit: game.Heart, Rank: game.Seven},
		{Suit: game.Club, Rank: game.Nine},
	}
	hub.g.Players[1].Hand = []game.Card{
		{Suit: game.Spade, Rank: game.Queen}, {Suit: game.Heart, Rank: game.Queen},
		{Suit: game.Club, Rank: game.Queen}, {Suit: game.Diamond, Rank: game.Queen},
		{Suit: game.Club, Rank: game.Three},
	}
	hub.g.CurrentTurn = 0
	hub.g.LeadPlayer = 0
	hub.g.LastPlay = nil
	hub.mu.Unlock()

	sendEnvelope(t, p0, TypePlayCards, PlayCardsRequest{Cards: []CardDTO{
		{Suit: "S", Rank: "4"}, {Suit: "H", Rank: "4"},
		{Suit: "S", Rank: "5"}, {Suit: "H", Rank: "5"},
		{Suit: "S", Rank: "6"}, {Suit: "H", Rank: "6"},
		{Suit: "S", Rank: "7"}, {Suit: "H", Rank: "7"},
	}})
	drainUntil(t, p0, TypePlayResult, 2*time.Second)
	drainUntil(t, p1, TypePlayResult, 2*time.Second)

	// p1 plays the bomb -- this must succeed and must NOT crash/close
	// the connection, even though it triggers an auto-pass for p0.
	sendEnvelope(t, p1, TypePlayCards, PlayCardsRequest{Cards: []CardDTO{
		{Suit: "S", Rank: "Q"}, {Suit: "H", Rank: "Q"}, {Suit: "C", Rank: "Q"}, {Suit: "D", Rank: "Q"},
	}})

	data := drainUntil(t, p1, TypePlayResult, 2*time.Second)
	var pr PlayResultMessage
	if err := json.Unmarshal(data, &pr); err != nil {
		t.Fatalf("unmarshal play_result: %v", err)
	}
	if pr.Category != "bomb" {
		t.Fatalf("expected category=bomb, got %q", pr.Category)
	}

	// The connection must still be alive and usable -- prove it by
	// requesting fresh state and getting a real response back.
	sendEnvelope(t, p1, TypeRequestState, RequestStateRequest{})
	drainUntil(t, p1, TypeStateSync, 2*time.Second)
}

func drainUntil(t *testing.T, conn *gws.Conn, msgType string, timeout time.Duration) json.RawMessage {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(deadline)
		var env Envelope
		if err := conn.ReadJSON(&env); err != nil {
			t.Fatalf("waiting for %s: %v", msgType, err)
		}
		if env.Type == msgType {
			return env.Data
		}
	}
	t.Fatalf("timed out waiting for %s", msgType)
	return nil
}

func sendEnvelope(t *testing.T, conn *gws.Conn, msgType string, data any) {
	t.Helper()
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := conn.WriteJSON(Envelope{Type: msgType, Data: raw}); err != nil {
		t.Fatalf("send: %v", err)
	}
}
