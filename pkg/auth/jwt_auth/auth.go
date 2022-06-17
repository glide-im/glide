package jwt_auth

import (
	"errors"
	"github.com/glide-im/glide/pkg/auth"
	"time"
)

type JwtAuthorize struct {
}

type JwtAuthInfo struct {
	UID         string
	Device      string
	ExpiredHour int64
}

type Response struct {
	Token  string
	Uid    string
	Device string
	Server []string
}

func NewAuthorizeImpl(secret string) *JwtAuthorize {
	jwtSecret = []byte(secret)
	return &JwtAuthorize{}
}

func (a JwtAuthorize) Auth(c auth.Info, t *auth.Token) (*auth.Result, error) {

	token, err := parseJwt(t.Token)
	if err != nil {

		result := auth.Result{
			Success: false,
			Msg:     "invalid token",
		}
		return &result, nil
	}

	return &auth.Result{
		Success: true,
		Response: &Response{
			Token:  t.Token,
			Uid:    token.Uid,
			Server: nil,
		},
	}, nil
}

func (a JwtAuthorize) RemoveToken(t *auth.Token) error {
	return nil
}

func (a JwtAuthorize) GetToken(c auth.Info) (*auth.Token, error) {

	info, ok := c.(*JwtAuthInfo)
	if !ok {
		return nil, errors.New("invalid auth info")
	}

	jt := Claims{
		Uid:    info.UID,
		Device: info.Device,
		Ver:    genJwtVersion(),
	}
	if info.ExpiredHour == 0 {
		info.ExpiredHour = 7 * 24
	}
	expire := time.Now().Add(time.Hour * time.Duration(info.ExpiredHour))
	token, err := genJwtExp(jt, expire)
	if err != nil {
		return nil, errors.New("generate token failed")
	}
	return &auth.Token{Token: token}, nil
}
