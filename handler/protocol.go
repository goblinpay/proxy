package handler

// client messages

// received:
// {"type":"auth","params":{"version":7,"userID":"deepMiner_wasm"}}
// {"type":"submit","params":{"version":7,"job_id":"yV44h4xagakrDcSeUe4x6jTgqWQf","nonce":"5ef9910b","result":"9a984afab161177a09e1b13dae4dcd795d477880733a50902d582feb733dbb01"}}
// sent:
// {"type":"authed","params":{"hashes":0}}
// {"type":"job","params":{"blob":"0b0bc89ddce40592ae5e7a6ce78ecd69eb1c1ceb0b853663a270c67a0c938ded7eefba2cf3882200000000909bab062025b5973d6c49cc6805a67276a9fef2c969645a02291906650609e702","algo":"cn/r","variant":"r","height":1797516,"job_id":"yV44h4xagakrDcSeUe4x6jTgqWQf","target":"285c8f02","id":"9e5e1bbe-601c-4035-bb07-4a2882b4b6e3"}}
// {"type":"hash_accepted","params":{"hashes":1}}

// inbound
type MessageFromClient struct {

	// enum: auth, submit
	Type string 					`json:"type"`
	Params MessageFromClientParams 	`json:"params"`
}
type MessageFromClientParams struct {

	Version uint 	`json:"version"`

	// auth
	UserId string 	`json:"userID"`

	// submit
	SubmitParams
}

const (
	// from client
	ClientTypeAuth = "auth"
	ClientTypeSubmit = "submit"

	// to client
	ClientTypeAuthed = "authed"
	ClientTypeJob = "job"
	ClientTypeHashAccepted = "hash_accepted"
	ClientTypeError = "error"
	ClientTypeBanned = "banned"
)

// outbound
type MessageToClient struct {

	// enum: authed, job, hash_accepted
	Type string 					`json:"type"`
	// MessageToClientParams (authed, hash_accepted) or JobParams (job)
	Params interface{} 				`json:"params"`
}
type MessageToClientParams struct {

	// authed, hash_accepted
	Hashes *uint 	`json:"hashes,omitempty"`
	Chunk string 	`json:"chunk,omitempty"`

	// error, banned
	Error string 	`json:"error,omitempty"`
	Banned string 	`json:"banned,omitempty"`
}


// common params

type JobParams struct {
	Blob string 	`json:"blob"`
	Algo string 	`json:"algo"`
	Variant string 	`json:"variant"`
	Height uint 	`json:"height"`
	JobId string 	`json:"job_id"`
	Target string 	`json:"target"`
	Id string 		`json:"id"`
}

type SubmitParams struct {
	JobId string 	`json:"job_id"`
	Nonce string 	`json:"nonce"`
	Result string 	`json:"result"`
}


// server messages

// received:
// {"id":"AE08F415F7850C33","jsonrpc":"2.0","error":null,"result":{"id":"7eaeb87f-9387-4941-bf13-965775e8a237","job":{"blob":"0b0bb5a0dce4054ed5d95723fa9c5ba35d0451ad11d5f3432abdfe19a73a420ad6965f4a0c5d5800000000917eb282ef71e4c6c33d76438726def7de194e198de4b5cdffdf14a53a31342a01","algo":"cn/r","variant":"r","height":1797518,"job_id":"u6I7dmCs0ky3sU/vs/ZyOjq/rbN2","target":"285c8f02","id":"7eaeb87f-9387-4941-bf13-965775e8a237"},"status":"OK"}}
// {"id":"AE08F415F7850C33","jsonrpc":"2.0","error":null,"result":{"status":"OK"}}
// {"jsonrpc":"2.0","method":"job","params":{"blob":"0b0be1a1dce405e5c52a1fbe5bfd46c817f9401c2701c4740c3bfd98a42c089a34929543d3273f0000000030a65b58fbeffedb8e5f5e26a0af3b2809866a19049beb308770a07f09c800bf04","algo":"cn/r","variant":"r","height":1797519,"job_id":"ffA8FaN8TQsiHSVBlzJaX5UEkYyR","target":"285c8f02","id":"7eaeb87f-9387-4941-bf13-965775e8a237"}}
// sent:
// {"method":"login","params":{"login":"41ynfGBUDbGJYYzz2jgSPG5mHrHJL4iMXEKh9EX6RfEiM9JuqHP66vuS2tRjYehJ3eRSt7FfoTdeVBfbvZ7Tesu1LKxioRU","pass":"x","rigid":"","agent":"deepMiner"},"id":"93F59B5B11F16F63"}
// {"method":"submit","params":{"id":"7eaeb87f-9387-4941-bf13-965775e8a237","job_id":"u6I7dmCs0ky3sU/vs/ZyOjq/rbN2","nonce":"ecf4fc6c","result":"5904d5f0c72bba781ada08d614f7f4da0036c361b02ecba19aa62c1316d48201"},"id":"AE08F415F7850C33"}

// inbound

type MessageFromServer struct {

	// used for authed & job, hash_accepted
	Id string 				`json:"id"`
	Error *ServerError 		`json:"error"`
	Result *ServerResult 	`json:"result"`

	// used for job
	Method string 			`json:"method"`
	Params *JobParams		`json:"params"`
}
type ServerResult struct {

	Status string 	`json:"status"`

	// authed & job
	Id string 		`json:"id"`
	Job *JobParams	`json:"job"`
}
type ServerError struct {
	Code int 		`json:"code"`
	Message string 	`json:"message"`
}

const (
	// from server
	ServerMethodJob = "job"

	// to server
	ServerMethodLogin = "login"
	ServerMethodSubmit = "submit"
)
const ServerStatusOk = "OK"

// outbound

type MessageToServer struct {

	// enum: login, submit
	Method string 			`json:"method"`
	Id string 				`json:"id"`
	Params interface{} 		`json:"params"`
}
type ServerParamsLogin struct {

	Login string 	`json:"login"`
	Pass string 	`json:"pass"`
	Rigid string 	`json:"rigid"`
	Agent string 	`json:"agent"`
}
type ServerParamsSubmit struct {

	Id string 		`json:"id"`
	SubmitParams
}

