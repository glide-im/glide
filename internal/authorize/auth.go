package authorize

import (
	"errors"
	"fmt"
	"github.com/glide-im/glide/pkg/auth"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"time"
)

type AuthorizeImpl struct {
}

func NewAuthorizeImpl() *AuthorizeImpl {
	return &AuthorizeImpl{}
}

func (a AuthorizeImpl) Auth(c *gate.Info, t *auth.Token) error {
	token, err := parseJwt(t.Token)
	if err != nil {
		return fmt.Errorf("invalid token")
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

	return nil
}

func (a AuthorizeImpl) RemoveToken(t *auth.Token) error {
	return nil
}

func (a AuthorizeImpl) GetToken(c *gate.Info) (*auth.Token, error) {
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
