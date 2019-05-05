package handler

import (
    "io"
    "encoding/json"
    "github.com/gorilla/websocket"

    "proxy/util"
    "proxy/fetcher"

    "sync/atomic"

    // TODO: move to a binary serialization format
    "encoding/base64"
    "bufio"
)

// TODO: read from config file, or convert to namespaced interface
const (
	Addr = "49w5a357GUZjmnzFmVe7CjRQf5AV1uRmFgNuZBZ2xCzYVY2KYm6DE18DC3BRkDXqXh7kbS93K78YzWNw3aa7SiNrHGZdyvs"
	Pass = "goblin:support@goblincompute.com"
	Agent = "goblin-proxy"
)

// TODO: write JSON + b64 encoder/scanners

func Client2Pool(dst io.Writer, src *websocket.Conn, session *util.Session, errc chan error) {
	
	// json.Encode adds a \n
	enc := json.NewEncoder(dst)

	// decoding
	pr, pw := io.Pipe() // should handle pw.Close() at break?
	jDec := json.NewDecoder(pr)

	// loop
	for {
		var (
			inbound MessageFromClient
			outbound MessageToServer
		)

		_, r, err := src.NextReader() // messageType is known (websocket.TextMessage)
		if err != nil {
			// client is offline
			errc <- err
			break
		}

		copyChan := make(chan error)
		go func(errChan chan error) {
			// could be moved up with ReadMessage+buffio
			bDec := base64.NewDecoder(base64.StdEncoding, r)

			if _, err := io.Copy(pw, bDec); err != nil {
				// base64 decode error
				errChan <- err
				return
			}
			errChan <- nil
		}(copyChan)

		if err := jDec.Decode(&inbound); err != nil {
			// JSON decode error
			errc <- err
			break
		}

		if err := <- copyChan; err != nil {
			errc <- err
			break
		}

		switch inbound.Type {
		case ClientTypeAuth:
			outbound.Method = ServerMethodLogin
			outbound.Params = ServerParamsLogin{
				Login: Addr,
				Pass: Pass,
				Agent: Agent,
			}
		case ClientTypeSubmit:
			outbound.Method = ServerMethodSubmit
			outbound.Params = ServerParamsSubmit{
				Id: session.WorkerId,
				SubmitParams: inbound.Params.SubmitParams,
			}
		default:
			// unexpected message
			// TODO: handle error
		}
		outbound.Id = session.Pid

		if err := enc.Encode(&outbound); err != nil {
			// pool no longer receiving data, or JSON encode error
			errc <- err
			break
		}
	}
}

func Pool2Client(dst *websocket.Conn, src io.Reader, session *util.Session, errc chan error) {
	
	// reads newline-deliminated JSON from TCP connection with pool
	dec := json.NewDecoder(src)

	// encoding
	pr, pw := io.Pipe() // should handle pw.Close() at break?
	jEnc := json.NewEncoder(pw)
	jEncScan := bufio.NewScanner(pr)

	// loop
	for {
		var (
			inbound MessageFromServer
			outbound []MessageToClient
		)

		if err := dec.Decode(&inbound); err != nil {
			// pool closed connection, or JSON decode error
			errc <- err
			break
		}

		switch {
		case inbound.Id == session.Pid && inbound.Result != nil && inbound.Result.Id != "":
			// authed & job
			session.WorkerId = inbound.Result.Id
			outbound = []MessageToClient{
				{Type: ClientTypeAuthed, Params: MessageToClientParams{Hashes: &session.Accepted,},},
				{Type: ClientTypeJob, Params: inbound.Result.Job,},
			}
		case inbound.Id == session.Pid && inbound.Result != nil && inbound.Result.Status == ServerStatusOk:
			// hash accepted
			session.Accepted++
			atomic.AddUint32(session.TokenSession.Accepted, 1)
			outbound = []MessageToClient{
				{Type: ClientTypeHashAccepted, Params: MessageToClientParams{
					Hashes: &session.Accepted,
					Chunk: fetcher.ReadChunk(&session.Content),
				},},
			}
		case inbound.Method == ServerMethodJob:
			outbound = []MessageToClient{
				{Type: ClientTypeJob, Params: inbound.Params,},
			}
		case inbound.Id == session.Pid && inbound.Error != nil && inbound.Error.Code == -1:
			outbound = []MessageToClient{
				{Type: ClientTypeError, Params: MessageToClientParams{Error: inbound.Error.Message,},},
			}
		case inbound.Id == session.Pid && inbound.Error != nil && inbound.Error.Code != -1:
			outbound = []MessageToClient{
				{Type: ClientTypeBanned, Params: MessageToClientParams{Banned: session.Pid,},},
			}
		default:
			// unexpected message
			// TODO: handle error
		}

		// write outbound websocket messages
		for _, msg := range outbound {

			w, err := dst.NextWriter(websocket.TextMessage)
			if err != nil {
				// client no longer receiving data
				errc <- err
				break
			}

			copyChan := make(chan error)
			go func(errChan chan error) {
				// could be moved up with WriteMessage+buffio
				bEnc := base64.NewEncoder(base64.StdEncoding, w)

				// TODO: handle error for Scan() method
				jEncScan.Scan()

				if _, err := bEnc.Write(jEncScan.Bytes()); err != nil {
					// base64 encode error
					errChan <- err
					return
				}

				if err := bEnc.Close(); err != nil {
					// base64 encode error?
					errChan <- err
					return
				}

				if err := w.Close(); err != nil {
					// client no longer receiving data?
					errChan <- err
					return
				}

				errChan <- nil
			}(copyChan)

			if err := jEnc.Encode(&msg); err != nil {
				// JSON encode error
				errc <- err
				break
			}

			if err := <- copyChan; err != nil {
				errc <- err
				break
			}
		}

		// if we are done reading content
		if len(session.Content) == 0 {
			// close pool and client connection
			errc <- nil
			break
		}
	}
}
