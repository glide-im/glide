package auth

import (
	"github.com/glide-im/glide/pkg/gate"
)

type Token struct {
	Token string
}

type Result struct {
	ID       gate.ID
	Success  bool
	Response interface{}
}

type Interface interface {
	Auth(c *gate.Info, t *Token) (*Result, error)
}

type Authorize interface {
	Interface

	RemoveToken(t *Token) error

	GetToken(c *gate.Info) (*Token, error)
}

type Server interface {
	Authorize

	Run() error
}
