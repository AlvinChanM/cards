package ws

import (
	"log"
	"math/rand"
	"sync"

	"github.com/cards/internal/game"
)

// Hub manages the single global match: at most two player Clients plus
// any number of spectators. Connection bookkeeping (register/unregister)
// runs through the run() goroutine's channels (classic gorilla/websocket
// chat-example pattern); game rules are protected by a plain mutex since
// there is no per-room fan-out to justify an actor-per-room design.
type Hub struct {
	register   chan *Client
	unregister chan *Client

	mu          sync.Mutex
	g           *game.Game
	players     [2]*Client // nil until that seat is taken
	spectators  map[*Client]bool
	rng         *rand.Rand
}

// NewHub creates a Hub with a fresh, not-yet-started Game.
func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		g:          game.NewGame(),
		spectators: make(map[*Client]bool),
		rng:        rand.New(rand.NewSource(1)),
	}
}

// Run processes register/unregister events. Must be started in its
// own goroutine before serving connections.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.handleRegister(c)
		case c := <-h.unregister:
			h.handleUnregister(c)
		}
	}
}

func (h *Hub) handleRegister(c *Client) {
	h.mu.Lock()
	switch {
	case h.players[0] == nil:
		c.seat = Seat0
		h.players[0] = c
	case h.players[1] == nil:
		c.seat = Seat1
		h.players[1] = c
	default:
		c.seat = SeatSpectator
		h.spectators[c] = true
	}

	bothSeated := h.players[0] != nil && h.players[1] != nil
	if bothSeated && h.g.Phase == game.WaitingForPlayers {
		h.g.StartRound(h.rng)
	}
	h.mu.Unlock()

	c.sendEnvelope(TypeJoined, JoinedMessage{
		Seat:        seatLabel(c.seat),
		PlayerNames: h.playerNamesLocked(),
	})
	h.broadcastStateLocked()
}

func (h *Hub) handleUnregister(c *Client) {
	h.mu.Lock()
	switch c.seat {
	case Seat0:
		h.players[0] = nil
	case Seat1:
		h.players[1] = nil
	default:
		delete(h.spectators, c)
	}
	close(c.send)
	h.mu.Unlock()
}

// dispatch routes one decoded client message. Called from the
// client's own readPump goroutine, so all game-state access here must
// go through h.mu. A recover guard turns any unexpected panic into an
// error response instead of killing the connection (and previously,
// before LastAcceptedPlay was introduced, an auto-pass-triggered nil
// dereference here would silently drop both players).
func (h *Hub) dispatch(c *Client, env Envelope) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ws: recovered panic handling %s: %v", env.Type, r)
			c.sendError("INTERNAL_ERROR", "server error handling your request, please retry")
		}
	}()

	switch env.Type {
	case TypeJoin:
		// Seat assignment already happened at connect time; join is a
		// no-op placeholder for future name-setting, acknowledged via a
		// fresh state_sync.
		h.mu.Lock()
		h.broadcastStateLocked()
		h.mu.Unlock()

	case TypePlayCards:
		var req PlayCardsRequest
		if err := unmarshalData(env.Data, &req); err != nil {
			c.sendError("BAD_MESSAGE", "invalid play_cards payload")
			return
		}
		cards, err := dtosToCards(req.Cards)
		if err != nil {
			c.sendError("BAD_MESSAGE", err.Error())
			return
		}
		h.handlePlay(c, cards)

	case TypePass:
		h.handlePass(c)

	case TypeRequestState:
		h.mu.Lock()
		h.broadcastStateLocked()
		h.mu.Unlock()

	case TypeRequestHint:
		h.handleHint(c)

	default:
		c.sendError("UNKNOWN_TYPE", "unrecognized message type: "+env.Type)
	}
}

func (h *Hub) handlePlay(c *Client, cards []game.Card) {
	if c.seat != Seat0 && c.seat != Seat1 {
		c.sendError("NOT_A_PLAYER", "spectators cannot play cards")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	seatIdx := int(c.seat)
	if err := h.g.PlayCards(seatIdx, cards); err != nil {
		c.sendError("INVALID_PLAY", err.Error())
		return
	}

	h.broadcastEnvelopeLocked(TypePlayResult, PlayResultMessage{
		PlayerIdx: seatIdx,
		Cards:     cardsToDTO(cards),
		Category:  categoryToStr[h.g.LastAcceptedPlay.Hand.Category],
	})

	if h.g.LastAutoPass >= 0 {
		h.broadcastEnvelopeLocked(TypePassResult, PassResultMessage{PlayerIdx: h.g.LastAutoPass})
	}

	if h.g.Phase == game.RoundOver {
		h.broadcastEnvelopeLocked(TypeRoundOver, RoundOverMessage{
			Winner:    h.g.RoundWinner,
			RoundsWon: h.g.RoundsWon,
		})
		h.g.StartRound(h.rng)
	}

	h.broadcastStateLocked()
}

func (h *Hub) handlePass(c *Client) {
	if c.seat != Seat0 && c.seat != Seat1 {
		c.sendError("NOT_A_PLAYER", "spectators cannot pass")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	seatIdx := int(c.seat)
	if err := h.g.Pass(seatIdx); err != nil {
		c.sendError("INVALID_PASS", err.Error())
		return
	}

	h.broadcastEnvelopeLocked(TypePassResult, PassResultMessage{PlayerIdx: seatIdx})
	h.broadcastStateLocked()
}

func (h *Hub) handleHint(c *Client) {
	if c.seat != Seat0 && c.seat != Seat1 {
		c.sendError("NOT_A_PLAYER", "spectators have no hand to suggest plays for")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.g.Phase != game.Playing {
		c.sendError("NOT_PLAYING", "hints are only available while a round is in progress")
		return
	}

	seatIdx := int(c.seat)
	var target *game.ClassifiedHand
	if h.g.LastPlay != nil && h.g.LastPlay.PlayerIdx != seatIdx {
		target = &h.g.LastPlay.Hand
	}

	plays := game.SuggestPlays(h.g.Players[seatIdx].Hand, target)
	dtoPlays := make([][]CardDTO, len(plays))
	for i, p := range plays {
		dtoPlays[i] = cardsToDTO(p)
	}
	c.sendEnvelope(TypeHint, HintMessage{Plays: dtoPlays})
}

// broadcastStateLocked sends a per-recipient-redacted state_sync to
// every connected client. Caller must hold h.mu.
func (h *Hub) broadcastStateLocked() {
	if h.g.Phase == game.WaitingForPlayers {
		count := 0
		for _, p := range h.players {
			if p != nil {
				count++
			}
		}
		for _, p := range h.players {
			if p != nil {
				p.sendEnvelope(TypeWaiting, WaitingMessage{PlayersConnected: count})
			}
		}
		for sp := range h.spectators {
			sp.sendEnvelope(TypeWaiting, WaitingMessage{PlayersConnected: count})
		}
		return
	}

	names := h.playerNamesLocked()
	for i, p := range h.players {
		if p != nil {
			p.sendEnvelope(TypeStateSync, h.stateSyncFor(i, names))
		}
	}
	for sp := range h.spectators {
		sp.sendEnvelope(TypeStateSync, h.stateSyncFor(-1, names))
	}
}

func (h *Hub) stateSyncFor(seat int, names [2]string) StateSyncMessage {
	msg := StateSyncMessage{
		Phase:       phaseToStr(h.g.Phase),
		YourSeat:    seat,
		CurrentTurn: h.g.CurrentTurn,
		LeadPlayer:  h.g.LeadPlayer,
		RoundsWon:   h.g.RoundsWon,
		PlayerNames: names,
	}

	if seat == 0 || seat == 1 {
		msg.YourHand = cardsToDTO(h.g.Players[seat].Hand)
		msg.OpponentCardCount = len(h.g.Players[1-seat].Hand)
	} else {
		// spectators get both counts folded into OpponentCardCount slot
		// is ambiguous for 2 hands; expose neither hand and rely on
		// clients not needing per-seat counts beyond what's useful for
		// a spectator view (both counts derivable if we extend later).
		msg.OpponentCardCount = 0
	}

	if h.g.LastPlay != nil {
		msg.LastPlay = &LastPlayDTO{
			PlayerIdx: h.g.LastPlay.PlayerIdx,
			Cards:     cardsToDTO(h.g.LastPlay.Cards),
			Category:  categoryToStr[h.g.LastPlay.Hand.Category],
		}
	}

	return msg
}

func (h *Hub) broadcastEnvelopeLocked(msgType string, data any) {
	for _, p := range h.players {
		if p != nil {
			p.sendEnvelope(msgType, data)
		}
	}
	for sp := range h.spectators {
		sp.sendEnvelope(msgType, data)
	}
}

func (h *Hub) playerNamesLocked() [2]string {
	var names [2]string
	for i, p := range h.players {
		if p != nil {
			names[i] = p.name
		}
	}
	return names
}

func seatLabel(s Seat) string {
	switch s {
	case Seat0:
		return "0"
	case Seat1:
		return "1"
	default:
		return "spectator"
	}
}
