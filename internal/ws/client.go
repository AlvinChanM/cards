package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

// Seat identifies what role a connection plays.
type Seat int

const (
	SeatSpectator Seat = -1
	Seat0         Seat = 0
	Seat1         Seat = 1
)

// Client wraps one WebSocket connection: a seat assignment, the
// underlying conn, and buffered outbound channel drained by a single
// writer goroutine (per gorilla/websocket's standard pattern -- never
// write to the same conn from two goroutines).
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	seat Seat
	name string
	send chan []byte
}

func newClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		seat: SeatSpectator,
		send: make(chan []byte, 16),
	}
}

// readPump reads client messages and hands them to the hub. Must run
// in its own goroutine; returns (and triggers unregister) on any
// connection error.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		var env Envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			c.sendError("BAD_MESSAGE", "malformed envelope")
			continue
		}
		c.hub.dispatch(c, env)
	}
}

// writePump drains the send channel and writes to the connection, and
// sends periodic pings. Must run in its own goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendEnvelope(msgType string, data any) {
	raw, err := encode(msgType, data)
	if err != nil {
		log.Printf("ws: encode %s: %v", msgType, err)
		return
	}
	select {
	case c.send <- raw:
	default:
		log.Printf("ws: dropping message for slow client (seat=%v)", c.seat)
	}
}

func (c *Client) sendError(code, message string) {
	c.sendEnvelope(TypeError, ErrorMessage{Code: code, Message: message})
}
