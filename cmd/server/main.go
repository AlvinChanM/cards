// Command server runs the 二人关牌 single-room match server: one HTTP
// listener serving the embedded React frontend and a /ws endpoint
// backing the one global game.
package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"

	"github.com/cards/internal/server"
	"github.com/cards/internal/ws"
)

//go:embed all:web
var embeddedWeb embed.FS

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	staticFS, err := fs.Sub(embeddedWeb, "web")
	if err != nil {
		log.Fatalf("server: embed sub fs: %v", err)
	}

	hub := ws.NewHub()
	go hub.Run()

	handler := server.New(hub, staticFS)

	log.Printf("listening on %s", *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal(err)
	}
}
