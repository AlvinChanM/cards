// Package server assembles the HTTP server: the /ws upgrade endpoint
// plus static file serving for the embedded React build.
package server

import (
	"io/fs"
	"net/http"

	"github.com/cards/internal/ws"
)

// New builds the top-level http.Handler. staticFS should be the
// embedded web/dist build output (or nil to disable static serving,
// e.g. during frontend dev-server iteration where Vite serves assets
// itself and only /ws needs to be reached).
func New(hub *ws.Hub, staticFS fs.FS) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWS(hub, w, r)
	})

	if staticFS != nil {
		mux.Handle("/", http.FileServer(http.FS(staticFS)))
	}

	return mux
}
