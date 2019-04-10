package db

import (
	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"

	"sync"

	"proxy/util"
)

var tokenSessions sync.Map

var db *sqlx.DB

func MustInitDb() {
	db = sqlx.MustConnect("postgres", "user=tarik dbname=goblin sslmode=disable")
}

func GetTokenSession(token string) (tokenSession *util.TokenSession, err error) {
	var (
		val interface{}
		ok bool
	)
	if val, ok = tokenSessions.Load(token); !ok {
		var baseUrl string // or create tokenSession object right here
		if baseUrl, err = GetBaseUrl(token); err != nil { // does this write the err for return??
			return
		}
		// only cache if the token was found
		// we must try to load again to avoid a race for an identical token
		val, _ = tokenSessions.LoadOrStore(token, &util.TokenSession{BaseUrl: baseUrl, Accepted: new(uint32)})
	}
	tokenSession = val.(*util.TokenSession)
	return
}

// TODO: avoid concurrent requests?? -- wait on pending => use channels HIGH PRIORITY use sync.Once
// TODO: debounce by 1 sec (cache errors for 1s)
func GetBaseUrl(token string) (baseUrl string, err error) {
	if err = db.Get(&baseUrl, "SELECT base_url FROM tokens WHERE token=$1", token); err != nil {
		return
	}
	// TODO: handle not existing, add err, or test it, already handled?
	return
}

// func PooledIncrementSharesForToken(token string) {
// 	// atomic increment
// }

// func StartCoutingPool() {
// 	// time.Tick
// 	// can also be embed in PooledIncrementSharesForToken
// }

// func CounterPool() { // use sync.Map + pointer to sync.Atomic
// 	// retrieve all values

// 	// for anything > 0, increment in db

// 	// decrement by values
// }
