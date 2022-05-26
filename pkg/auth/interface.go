package auth

import (
	"github.com/glide-im/glide/pkg/gate"
)

type Token struct {
	Token string
}

type Result struct {
	ID      int64
	Token   string
	Servers []string
}

type Interface interface {
	Auth(c *gate.Info, t *Token) error
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
