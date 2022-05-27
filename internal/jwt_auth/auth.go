package jwt_auth

import (
	"errors"
	"fmt"
	"github.com/glide-im/glide/pkg/auth"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"strconv"
	"time"
)

type JwtAuthorize struct {
}

type Response struct {
	Token  string
	Uid    string
	Server []string
}

func NewAuthorizeImpl(secret string) *JwtAuthorize {
	jwtSecret = []byte(secret)
	return &JwtAuthorize{}
}

func (a JwtAuthorize) Auth(c *gate.Info, t *auth.Token) (*auth.Result, error) {
	token, err := parseJwt(t.Token)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	id := c.ID.UID()
	device := c.ID.Device()

	//version, err := userdao.Dao.GetTokenVersion(token.Uid, token.Device)
	//if err != nil || version == 0 || version > token.Ver {
	//	return nil, fmt.Errorf("invalid token")
	//}

	if id == token.Uid && device == token.Device {
		// logged in
		logger.D("auth token for a connection is logged in")
	}

	return &auth.Result{
		ID:      gate.NewID("", token.Uid, strconv.FormatInt(token.Device, 10)),
		Success: true,
		Response: &Response{
			Token:  t.Token,
			Uid:    strconv.FormatInt(token.Uid, 10),
			Server: nil,
		},
	}, nil
}

func (a JwtAuthorize) RemoveToken(t *auth.Token) error {
	return nil
}

func (a JwtAuthorize) GetToken(c *gate.Info) (*auth.Token, error) {
	jt := Claims{
		Uid:    c.ID.UID(),
		Device: c.ID.Device(),
		Ver:    genJwtVersion(),
	}
	expire := time.Now().Add(time.Hour * time.Duration(24*7))
	token, err := genJwtExp(jt, expire)
	if err != nil {
		return nil, errors.New("generate token failed")
	}

	//err = userdao.Dao.SetTokenVersion(jt.Uid, jt.Device, jt.Ver, time.Duration(jt.ExpiresAt))
	//if err != nil {
	//	return "", fmt.Errorf("generate token failed")
	//}

	return &auth.Token{Token: token}, nil
}
