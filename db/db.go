package db

import (
	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"

	"sync"
	"sync/atomic"
	"time"

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
		// we must try to load again to properly handle a race for an identical token
		val, _ = tokenSessions.LoadOrStore(token, &util.TokenSession{BaseUrl: baseUrl, Accepted: new(uint32)})
	}
	tokenSession = val.(*util.TokenSession)
	return
}

// TODO: avoid concurrent requests?? -- wait on pending => use channels / sync.Once
// TODO: debounce by 1 sec (cache errors for 1s)
func GetBaseUrl(token string) (baseUrl string, err error) {
	if err = db.Get(&baseUrl, "SELECT base_url FROM tokens WHERE token=$1", token); err != nil {
		return
	}
	// TODO: handle not existing, add err, or test it, already handled?
	return
}

func StartCounterTicker() {
	ticker := time.NewTicker(time.Second * 10)
	go func() {
		defer ticker.Stop() // is this really necessary?
		for {
			<-ticker.C
			CounterTick() // the ticker ensures this only gets called once at a time
			              // by dropping ticks if this is slow
		}
	}()
}

func CounterTick() { // use sync.Map + pointer to sync.Atomic
	// retrieve all values
	tokenSessions.Range(func(tokenVal interface{}, tokenSessionVal interface{}) bool {

		var (
			token = tokenVal.(string)
			tokenSession = tokenSessionVal.(*util.TokenSession)
		)
		
		if inc := atomic.LoadUint32(tokenSession.Accepted); inc > 0 {
			// increment in db, handle panic or return false?
			db.MustExec("UPDATE tokens SET solved_hashes=solved_hashes+$1 WHERE token=$2", inc, token)

			// if it did not fail, atomic decrement
			atomic.AddUint32(tokenSession.Accepted, ^uint32(inc-1))
		}

		return true
	})
}
