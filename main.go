// design inspired by:
// - https://github.com/koding/websocketproxy/blob/master/websocketproxy.go
// - https://github.com/isobit/ws-tcp-relay/blob/master/ws-tcp-relay.go

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"fmt"

	"proxy/handler"
	"proxy/db"
)

func main() {
	// connect to DB
	db.MustInitDb()

	// init pooled DB increments
	db.StartCounterTicker()

	// http.DefaultServeMux
	http.HandleFunc("/", handler.HttpHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8181"
		log.Printf("Defaulting to port %s", port)
	}

	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", port),
		Handler: nil, // nil => http.DefaultServeMux
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigint

		// We received an interrupt signal, shut down.
		log.Printf("Shutdown request (signal: %v)", sig)
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("Listening on port %s", port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
	log.Printf("HTTP server has been Shutdown")
	// no longer accepting http connections, but may still have ws conns

	log.Printf("Waiting for Websocket connections to finish")
	handler.Wg.Wait()

	// when shutdown ready: stop db tick and flush
	log.Printf("Waiting for TokenSessions to flush")
	db.ShutdownCounterTicker()
}
