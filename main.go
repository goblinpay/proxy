// design inspired by:
// - https://github.com/koding/websocketproxy/blob/master/websocketproxy.go
// - https://github.com/isobit/ws-tcp-relay/blob/master/ws-tcp-relay.go

package main

import (
    "log"
    "net"
    "net/http"

    "github.com/gorilla/websocket"

    "proxy/handler"
    "proxy/fetcher"
    "proxy/util"
)

// TODO: read from config file
const (
	Listen = ":8181"
	Pool = "gulf.moneroocean.stream:80" // 100 difficulty, TODO: use secure, on port 443?
)

func main() {
    http.HandleFunc("/", httpHandler)

    if err := http.ListenAndServe(Listen, nil); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

func httpHandler(rw http.ResponseWriter, req *http.Request) {

	// special handshake headers
	// - Goblin token -- if it fails, mine for Goblin?
	// - Backend URL -- stop if it fails

	// TODO: read from headers
	const contentURL = "http://localhost:8080/content.partial.html"
	var (
		session util.Session
		err error // https://github.com/golang/go/issues/6842
	)

	if session.Pid, err = util.GetRandomHexString(16); err != nil {
		log.Printf("util.randomHex: error reading from random generator %s", err)
		return
	}

	// TODO: move to better location
	if session.Content, err = fetcher.InitRequest(contentURL); err != nil {
		log.Printf("fetcher.InitRequest: error fetching content %s", err)
		return
	}

	// open TCP socket
	connPool, err := net.Dial("tcp", Pool)
	if err != nil {
		log.Printf("net.Dial: couldn't dial to pool %s", err)
		return
	}
	defer connPool.Close()

	// Only pass those headers to the upgrader.
	upgradeHeader := http.Header{}

	// now upgrade client connection
	connClient, err := upgrader.Upgrade(rw, req, upgradeHeader)
	if err != nil {
		log.Printf("upgrader.Upgrade: couldn't upgrade %s", err)
		return
	}
	defer connClient.Close()

	errClient := make(chan error, 1)
	errPool := make(chan error, 1)

	go handler.Pool2Client(connClient, connPool, &session, errClient)
	go handler.Client2Pool(connPool, connClient, &session, errPool)

	var messageFormat string
	select {
	case err = <-errClient:
		messageFormat = "Pool2Client: Error when forwarding from pool to client: %v"
	case err = <-errPool:
		messageFormat = "Client2Pool: Error when forwarding from client to pool: %v"

	}
	if err != nil {
		log.Printf(messageFormat, err)
	}
}
