package util

type TokenSession struct {
	BaseUrl string
	Accepted *uint32 // to increment with sync/atomic
}

type Session struct {
  TokenSession *TokenSession
  Pid string // generated
  Content []byte
  WorkerId string // given by pool
  Accepted uint
}

type Config struct {

}