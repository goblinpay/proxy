// design inspired by:
// - https://github.com/koding/websocketproxy/blob/master/websocketproxy.go
// - https://github.com/isobit/ws-tcp-relay/blob/master/ws-tcp-relay.go

package main

import (
	"log"
	"net/http"
	"os"
	"fmt"

	"proxy/handler"
	"proxy/db"
)

func main() {
	// connect to DB
	db.MustInitDb()

	// init pooled DB increments
	db.StartCounterTicker()

	http.HandleFunc("/", handler.HttpHandler)

	port := os.Getenv("PORT")
	if port == "" {
	port = "8181"
	log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
			log.Fatal("ListenAndServe:", err)
	}
}
