package proxy

import (
    "fmt" // TODO: change to log...
    "errors"

    "encoding/json"
    "github.com/gorilla/websocket"

    "../fetcher"
)

// TODO: enforce enums
type PoolIdentifier string
const (
	Job PoolIdentifier = "job"
	HashSolved PoolIdentifier = "hashsolved"
)
type PoolMethod struct {
	Identifier PoolIdentifier `json:"identifier"`
}

type ClientIdentifier string
const (
	Handshake ClientIdentifier = "handshake"
	Solved ClientIdentifier = "solved"
)
type ClientMethod struct {
	Identifier ClientIdentifier `json:"identifier"`
}


type ProxyHashSolved struct {
	Identifier PoolIdentifier `json:"identifier"`
	Chunk string 							`json:"chunk"`
}


func ReplicateWebsocketConnFromPool(dst, src *websocket.Conn, errc chan error, contentBuffer []byte) {
	for {
		msgType, msg, err := src.ReadMessage()
		if err != nil {
			m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
			if e, ok := err.(*websocket.CloseError); ok {
				if e.Code != websocket.CloseNoStatusReceived {
					m = websocket.FormatCloseMessage(e.Code, e.Text)
				}
			}
			errc <- err
			dst.WriteMessage(websocket.CloseMessage, m)
			break
		}

		var method PoolMethod
		err = json.Unmarshal(msg, &method)
		if err != nil {
			fmt.Println("json.Unmarshal error:", err)
		}

		if method.Identifier == HashSolved {
			var proxyResponse ProxyHashSolved

			proxyResponse.Identifier = HashSolved
			proxyResponse.Chunk = fetcher.ReadChunk(&contentBuffer)

			msg, err := json.Marshal(proxyResponse)
			if err != nil {
				fmt.Println("json.Marshal error:", err)
			}
			
			err = dst.WriteMessage(msgType, msg)
			if err != nil {
				errc <- err
				break
			}

			if len(contentBuffer) == 0 { // we are actually done reading content, close connection
				errc <- errors.New("Done reading.")
				break
			}
		} else {
			// just regular proxying
			err = dst.WriteMessage(msgType, msg)
			if err != nil {
				errc <- err
				break
			}
		}
	}
}

func ReplicateWebsocketConnFromClient(dst, src *websocket.Conn, errc chan error) {
	for {
		msgType, msg, err := src.ReadMessage()
		if err != nil {
			m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
			if e, ok := err.(*websocket.CloseError); ok {
				if e.Code != websocket.CloseNoStatusReceived {
					m = websocket.FormatCloseMessage(e.Code, e.Text)
				}
			}
			errc <- err
			dst.WriteMessage(websocket.CloseMessage, m)
			break
		}

		err = dst.WriteMessage(msgType, msg)
		if err != nil {
			errc <- err
			break
		}
	}
}

// parse JSON
// identifier:
// - client:
//   - handshake (pass on, later will add params)
//   - solved (pass on)
// - pool:
//   - job (pass on)
//   - hashsolved (pass on but add content)
