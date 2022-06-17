package auth

type Token struct {
	Token string
}

type Result struct {
	Success  bool
	Msg      string
	Response interface{}
}

type Info interface {
}

type Interface interface {
	Auth(c Info, t *Token) (*Result, error)
}

type Authorize interface {
	Interface

	RemoveToken(t *Token) error

	GetToken(c Info) (*Token, error)
}

type Server interface {
	Authorize

	Run() error
}
