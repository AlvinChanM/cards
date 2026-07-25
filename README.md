# 跑得快 (Zhao Pao Kuai)

A real-time, two-player web implementation of **跑得快** (Zǒu/Pǎodékuài), a
classic Chinese climbing card game (2-player 关牌 variant). Built with a Go
WebSocket server and a React + TypeScript client.

## Features

- Full rules engine: singles, pairs, triples, triple+one, triple+pair,
  straights, consecutive pairs, airplanes (with/without kickers), and bombs
- Real-time two-player matches over WebSocket, with spectator support
- Server-side auto-pass when a player holds no legal beating play
- Play hints/suggestions (`request_hint`) computed server-side
- Animated UI: dealt-hand fly-in, card selection lift, play/pass feedback,
  bomb shake + flash effects, hand-type stamps, low-card alerts, round-over
  modal, and running score tally
- Lightweight WebAudio sound effects — no audio asset files needed

## Project layout

```
cmd/server/         main entrypoint; serves the API/WebSocket and (in prod) the built web assets
internal/game/      pure game rules engine (deck, hand classification, comparisons, auto-pass, hints)
internal/ws/        WebSocket transport: hub (match/session state), client, wire protocol, DTO conversion
web/                React + TypeScript client (Vite)
```

The rules engine (`internal/game`) has no knowledge of WebSocket or JSON —
it's a small, deterministic state machine that's easy to unit test in
isolation. `internal/ws` wraps it with the transport layer and protocol
encoding.

## Requirements

- Go 1.23+
- Node.js 18+ and npm

## Running locally

Two options, depending on whether you want hot-reload on the frontend.

### Development (hot-reload frontend)

Run the Go API server and the Vite dev server side by side:

```sh
make dev-server   # go run ./cmd/server -- serves the WebSocket API on :8080
make dev-web      # cd web && npm run dev -- Vite dev server with HMR
```

Open the URL Vite prints (typically `http://localhost:5173`). The dev server
proxies WebSocket calls to the Go backend.

### Production-style build

Build the web client and embed it into the Go binary's served assets, then
run a single binary:

```sh
make build   # builds web/, copies dist/ into cmd/server/web/, builds bin/server
make run     # builds (if needed) and runs ./bin/server
```

By default the server listens on `:8080`; open `http://localhost:8080`.

Play a match by opening two browser tabs/windows (or have a second player
connect) — the first two connections take seat 0 and seat 1, any further
connections join as spectators.

## Testing

```sh
make test    # go test ./...
```

Frontend type-checking and build validation:

```sh
cd web
npm run build   # tsc -b && vite build
npm run lint    # oxlint
```

## Protocol

Client and server exchange a single `Envelope { type, data }` JSON message
over WebSocket. The message types and payload shapes are defined in
`internal/ws/protocol.go` (Go) and mirrored by hand in
`web/src/types/protocol.ts` (TypeScript) — keep the two in sync when the
protocol changes.
