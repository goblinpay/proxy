package handler

import (
    "io"
    "encoding/json"
    "github.com/gorilla/websocket"

    "proxy/util"
    "proxy/fetcher"
)

// TODO: read from config file, or convert to namespaced interface
const (
	Addr = "49w5a357GUZjmnzFmVe7CjRQf5AV1uRmFgNuZBZ2xCzYVY2KYm6DE18DC3BRkDXqXh7kbS93K78YzWNw3aa7SiNrHGZdyvs"
	Pass = "goblin:support@goblincompute.com"
	Agent = "goblin-proxy"
)

func Client2Pool(dst io.Writer, src *websocket.Conn, session *util.Session, errc chan error) {
	// json.Encode adds a \n
	enc := json.NewEncoder(dst)
	for {
		var (
			inbound MessageFromClient
			outbound MessageToServer
		)

		if err := src.ReadJSON(&inbound); err != nil {
			// client is offline, or JSON decode error
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

			if err := dst.WriteJSON(&msg); err != nil {
				// client no longer receiving data, or JSON encode error
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
