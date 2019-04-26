// design inspired by:
// - https://github.com/koding/websocketproxy/blob/master/websocketproxy.go
// - https://github.com/isobit/ws-tcp-relay/blob/master/ws-tcp-relay.go

package main

import (
    "log"
    "net"
    "net/http"
    "time"
    "strings"
    "os"
    "fmt"

    "github.com/gorilla/websocket"

    "proxy/handler"
    "proxy/fetcher"
    "proxy/util"
    "proxy/db"
)

// TODO: read from config file
const (
	Pool = "gulf.moneroocean.stream:443" // 100 difficulty, TODO: use secure, on port 443?
)

func main() {
	// connect to DB
	db.MustInitDb()

	// init pooled DB increments
	db.StartCounterTicker()

    http.HandleFunc("/", httpHandler)

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

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

func httpHandler(rw http.ResponseWriter, req *http.Request) {

	// special handshake headers
	// - Goblin token -- if it fails, mine for Goblin?
	// - Backend URL -- stop if it fails

	// use url.Parse for path validation?

	// health check required by kube's ingress
	if req.RequestURI == "/" {
		rw.WriteHeader(http.StatusOK)
		return
	}

	var (
		session util.Session
		err error // https://github.com/golang/go/issues/6842
	)

	// read token and path
	parts := strings.SplitN(req.RequestURI, "/", 3);
	if len(parts) < 3 {
		log.Printf("cannot parse token/path")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	token := parts[1]
	path := parts[2]

	// TOOD: validate token before triggering a possible db select
	if session.TokenSession, err = db.GetTokenSession(token); err != nil {
		log.Printf("db.GetTokenSession: cannot retrieve token session %s", token)
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if session.Pid, err = util.GetRandomHexString(16); err != nil {
		log.Printf("util.randomHex: error reading from random generator %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: move to better location
	if session.Content, err = fetcher.Fetch(session.TokenSession.BaseUrl + path); err != nil {
		log.Printf("fetcher.Fetch: error fetching content %s", err)
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// open TCP socket
	connPool, err := net.DialTLS("tcp", Pool)
	if err != nil {
		log.Printf("net.Dial: couldn't dial to pool %s", err)
		rw.WriteHeader(http.StatusBadGateway)
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
		if err != nil {
			messageFormat = "Pool2Client: Error when forwarding from pool to client: %v"
		} else {
			connClient.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				time.Now().Add(time.Duration(10)*time.Second))
			// TODO: should introduce a delay as this races with the defer, should conserve order?
		}
	case err = <-errPool:
		messageFormat = "Client2Pool: Error when forwarding from client to pool: %v"

	}
	if err != nil {
		log.Printf(messageFormat, err)
	}
}
