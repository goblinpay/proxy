package main

import (
    "log"
    "net/http"
    "io"

    "github.com/gorilla/websocket"

    "proxy/handler"
    "proxy/fetcher"
)

const PoolProxyUrl = "ws://localhost:8181/"
const testContent = "http://localhost:8081/public/content.partial.html"

var (
	upgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	dialer = websocket.DefaultDialer
)

func handler(rw http.ResponseWriter, req *http.Request) {

	// Pass headers from the incoming request to the dialer to forward them to
	// the final destinations.
	requestHeader := http.Header{}

	// special handshake headers
	// - Goblin token -- if it fails, mine for Goblin?
	// - Backend URL -- stop if it fails

	respBuffer, err := fetcher.InitRequest(testContent) // TODO: move to better location
	if err != nil {
		// handle errors here
	}
	
	connBackend, resp, err := dialer.Dial(PoolProxyUrl, requestHeader)
	if err != nil {
		log.Printf("websocketproxy: couldn't dial to remote backend url %s", err)
		if resp != nil {
			// If the WebSocket handshake fails, ErrBadHandshake is returned
			// along with a non-nil *http.Response so that callers can handle
			// redirects, authentication, etcetera.
			if err := copyResponse(rw, resp); err != nil {
				log.Printf("websocketproxy: couldn't write response after failed remote backend handshake: %s", err)
			}
		} else {
			http.Error(rw, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		}
		return
	}
	defer connBackend.Close()

	// Only pass those headers to the upgrader.
	upgradeHeader := http.Header{}

	// now upgrade client connection
	connPub, err := upgrader.Upgrade(rw, req, upgradeHeader)
	if err != nil {
		log.Printf("websocketproxy: couldn't upgrade %s", err)
		return
	}
	defer connPub.Close()

	errClient := make(chan error, 1)
	errBackend := make(chan error, 1)

	go proxy.ReplicateWebsocketConnFromPool(connPub, connBackend, errClient, respBuffer)
	go proxy.ReplicateWebsocketConnFromClient(connBackend, connPub, errBackend)


	var message string
	select {
	case err = <-errClient:
		message = "websocketproxy: Error when copying from backend to client: %v"
	case err = <-errBackend:
		message = "websocketproxy: Error when copying from client to backend: %v"

	}
	if e, ok := err.(*websocket.CloseError); !ok || e.Code == websocket.CloseAbnormalClosure {
		log.Printf(message, err)
	}
}

func main() {
    http.HandleFunc("/", handler)

    if err := http.ListenAndServe(":8383", nil); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func copyResponse(rw http.ResponseWriter, resp *http.Response) error {
	copyHeader(rw.Header(), resp.Header)
	rw.WriteHeader(resp.StatusCode)
	defer resp.Body.Close()

	_, err := io.Copy(rw, resp.Body)
	return err
}
