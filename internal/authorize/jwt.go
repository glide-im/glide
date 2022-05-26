package authorize

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"time"
)

var jwtSecret []byte = []byte("glide-jwt-secret")

type Claims struct {
	jwt.StandardClaims
	Uid    int64 `json:"uid"`
	Device int64 `json:"device"`
	Ver    int64 `json:"ver"`
}

func genJwt(payload Claims) (string, error) {
	expireAt := time.Now().Add(time.Hour * 24)
	return genJwtExp(payload, expireAt)
}

func genJwtExp(payload Claims, expiredAt time.Time) (string, error) {
	payload.ExpiresAt = expiredAt.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	t, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return t, nil
}

func parseJwt(token string) (*Claims, error) {
	j := Claims{}
	t, err := jwt.ParseWithClaims(token, &j, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	jwtToken, ok := t.Claims.(*Claims)
	if !ok {
		return nil, errors.New("parse jwt error")
	}
	return jwtToken, nil
}

func genJwtVersion() int64 {
	return time.Now().Unix()
}
