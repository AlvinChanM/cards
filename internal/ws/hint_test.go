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

func TestRequestHintReturnsLegalPlays(t *testing.T) {
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

	hub.mu.Lock()
	hub.g.Players[0].Hand = []game.Card{
		{Suit: game.Spade, Rank: game.Seven},
		{Suit: game.Club, Rank: game.Four},
	}
	hub.g.Players[1].Hand = []game.Card{
		{Suit: game.Spade, Rank: game.Three},
		{Suit: game.Heart, Rank: game.Nine},
		{Suit: game.Spade, Rank: game.King},
	}
	hub.g.CurrentTurn = 0
	hub.g.LeadPlayer = 0
	hub.g.LastPlay = nil
	hub.mu.Unlock()

	sendEnvelope(t, p0, TypePlayCards, PlayCardsRequest{Cards: []CardDTO{{Suit: "S", Rank: "7"}}})
	drainUntil(t, p0, TypePlayResult, 2*time.Second)
	drainUntil(t, p1, TypePlayResult, 2*time.Second)

	sendEnvelope(t, p1, TypeRequestHint, RequestHintRequest{})
	data := drainUntil(t, p1, TypeHint, 2*time.Second)
	var hint HintMessage
	if err := json.Unmarshal(data, &hint); err != nil {
		t.Fatalf("unmarshal hint: %v", err)
	}

	if len(hint.Plays) != 2 {
		t.Fatalf("expected exactly 2 legal plays (Nine, King singles; Three cannot beat Seven), got %d: %+v", len(hint.Plays), hint.Plays)
	}
	for _, play := range hint.Plays {
		if len(play) != 1 {
			t.Fatalf("expected single-card plays, got %v", play)
		}
		if play[0].Rank == "3" {
			t.Fatalf("hint incorrectly suggested the Three, which cannot beat a Seven")
		}
	}
}

func TestRequestHintRejectedForSpectator(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWS(hub, w, r)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	p0, _, _ := gws.DefaultDialer.Dial(wsURL, nil)
	defer p0.Close()
	drainUntil(t, p0, TypeJoined, 2*time.Second)
	drainUntil(t, p0, TypeWaiting, 2*time.Second)

	p1, _, _ := gws.DefaultDialer.Dial(wsURL, nil)
	defer p1.Close()
	drainUntil(t, p1, TypeJoined, 2*time.Second)
	drainUntil(t, p0, TypeStateSync, 2*time.Second)
	drainUntil(t, p1, TypeStateSync, 2*time.Second)

	spectator, _, _ := gws.DefaultDialer.Dial(wsURL, nil)
	defer spectator.Close()
	drainUntil(t, spectator, TypeJoined, 2*time.Second)

	sendEnvelope(t, spectator, TypeRequestHint, RequestHintRequest{})
	data := drainUntil(t, spectator, TypeError, 2*time.Second)
	var em ErrorMessage
	if err := json.Unmarshal(data, &em); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if em.Code != "NOT_A_PLAYER" {
		t.Fatalf("expected NOT_A_PLAYER, got %s", em.Code)
	}
}
