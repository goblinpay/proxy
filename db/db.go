package db

import (
	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"

	"sync"
)

var Db *sqlx.DB
var baseUrls sync.Map

func MustInitDb() {
	Db = sqlx.MustConnect("postgres", "user=tarik dbname=goblin sslmode=disable")
}

func GetBaseUrlCached(token string) (baseUrl string, err error) {
	if val, ok := baseUrls.Load(token); !ok {
		if baseUrl, err = GetBaseUrl(token); err != nil {
			return
		}
		// only cache if the token was found
		baseUrls.Store(token, baseUrl)
	} else {
		baseUrl = val.(string)
	}
	return
}

// TODO: avoid concurrent requests?? -- wait on pending => use channels
// TODO: debounce by 1 sec (cache errors for 1s)
func GetBaseUrl(token string) (baseUrl string, err error) {
	if err = Db.Get(&baseUrl, "SELECT base_url FROM tokens WHERE token=$1", token); err != nil {
		return
	}
	// TODO: handle not existing, add err, or test it, already handled?
	return
}
