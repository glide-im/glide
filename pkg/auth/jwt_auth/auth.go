package jwt_auth

import (
	"errors"
	"fmt"
	"github.com/glide-im/glide/pkg/auth"
	"github.com/glide-im/glide/pkg/logger"
	"time"
)

type JwtAuthorize struct {
}

type JwtAuthInfo struct {
	UID    string
	Device string
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

	info, ok := c.(*JwtAuthInfo)
	if !ok {
		return nil, errors.New("invalid auth info")
	}

	token, err := parseJwt(t.Token)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	//version, err := userdao.Dao.GetTokenVersion(token.Uid, token.Device)
	//if err != nil || version == 0 || version > token.Ver {
	//	return nil, fmt.Errorf("invalid token")
	//}

	if info.UID == token.Uid && info.Device == token.Device {
		// logged in
		logger.D("auth token for a connection is logged in")
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
