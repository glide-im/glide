package gate

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/glide-im/glide/pkg/messages"
)

type CredentialCrypto interface {
	EncryptCredentials(c *ClientAuthCredentials) ([]byte, error)

	DecryptCredentials(src []byte) (*ClientAuthCredentials, error)
}

// AesCBCCrypto cbc mode PKCS7 padding
type AesCBCCrypto struct {
	Key []byte
}

func NewAesCBCCrypto(key []byte) *AesCBCCrypto {
	keyLen := len(key)
	count := 0
	switch true {
	case keyLen <= 16:
		count = 16 - keyLen
	case keyLen <= 24:
		count = 24 - keyLen
	case keyLen <= 32:
		count = 32 - keyLen
	default:
		key = key[:32]
	}
	if count != 0 {
		key = append(key, bytes.Repeat([]byte{0}, count)...)
	}
	return &AesCBCCrypto{Key: key}
}

func (a *AesCBCCrypto) EncryptCredentials(c *ClientAuthCredentials) ([]byte, error) {
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	// generate random iv
	iv := make([]byte, aes.BlockSize)
	_, err = rand.Read(iv)
	if err != nil {
		return nil, err
	}

	encryptBody, err := a.Encrypt(jsonBytes, iv)
	if err != nil {
		return nil, err
	}

	// NOTE: append iv
	var encrypt []byte
	encrypt = append(encrypt, iv...)
	encrypt = append(encrypt, encryptBody...)

	// base64 encoding encrypted json credentials
	b64Bytes := make([]byte, base64.RawStdEncoding.EncodedLen(len(encrypt)))
	base64.RawStdEncoding.Encode(b64Bytes, encrypt)
	return b64Bytes, nil
}

func (a *AesCBCCrypto) DecryptCredentials(src []byte) (*ClientAuthCredentials, error) {

	encrypt := make([]byte, base64.RawStdEncoding.DecodedLen(len(src)))
	_, err := base64.RawStdEncoding.Decode(encrypt, src)
	if err != nil {
		return nil, err
	}
	var iv []byte
	iv = append(iv, encrypt[:aes.BlockSize]...)
	var encryptBody []byte
	encryptBody = append(encryptBody, encrypt[aes.BlockSize:]...)

	jsonBytes, err := a.Decrypt(encryptBody, iv)
	if err != nil {
		return nil, err
	}

	credentials := ClientAuthCredentials{}
	err = json.Unmarshal(jsonBytes, &credentials)
	if err != nil {
		return nil, err
	}
	return &credentials, nil
}

func (a *AesCBCCrypto) Encrypt(src, iv []byte) ([]byte, error) {

	block, err := aes.NewCipher(a.Key)
	if err != nil {
		return nil, err
	}
	// padding
	blockSize := block.BlockSize()
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	src = append(src, padtext...)

	encryptData := make([]byte, len(src))

	if len(iv) != block.BlockSize() {
		iv = a.cbcIVPending(iv, blockSize)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encryptData, src)

	return encryptData, nil
}

func (a *AesCBCCrypto) Decrypt(src, iv []byte) ([]byte, error) {

	block, err := aes.NewCipher(a.Key)
	if err != nil {
		return nil, err
	}

	dst := make([]byte, len(src))
	blockSize := block.BlockSize()
	if len(iv) != blockSize {
		iv = a.cbcIVPending(iv, blockSize)
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(dst, src)

	length := len(dst)
	if length == 0 {
		return nil, errors.New("unpadding")
	}
	unpadding := int(dst[length-1])
	if length < unpadding {
		return nil, errors.New("unpadding")
	}
	res := dst[:(length - unpadding)]

	return res, nil
}

func (a *AesCBCCrypto) cbcIVPending(iv []byte, blockSize int) []byte {
	k := len(iv)
	if k < blockSize {
		return append(iv, bytes.Repeat([]byte{0}, blockSize-k)...)
	} else if k > blockSize {
		return iv[0:blockSize]
	}
	return iv
}

type Authenticator struct {
	credentialCrypto CredentialCrypto
	gateway          DefaultGateway
}

func NewAuthenticator(key string) *Authenticator {
	k := sha512.New().Sum([]byte(key))
	return &Authenticator{
		credentialCrypto: NewAesCBCCrypto(k),
	}
}

func (a *Authenticator) ClientAuthMessageInterceptor(dc DefaultClient, msg *messages.GlideMessage) (intercept bool) {
	if msg.Action != messages.ActionAuthenticate {
		return false
	}
	intercept = true

	credential := EncryptedCredential{}
	err := msg.Data.Deserialize(&credential)
	if err != nil {
		_ = dc.EnqueueMessage(messages.NewMessage(0, messages.ActionNotifyError, "invalid authenticate message"))
		return
	}

	authCredentials, err := a.credentialCrypto.DecryptCredentials([]byte(credential.Credential))
	if err != nil {
		_ = dc.EnqueueMessage(messages.NewMessage(0, messages.ActionNotifyError, "invalid authenticate message"))
		return
	}

	id, err := a.updateClient(dc, authCredentials)
	if err != nil {
		_ = dc.EnqueueMessage(messages.NewMessage(0, messages.ActionNotifyError, "invalid authenticate message"))
		return
	}
	_ = a.gateway.EnqueueMessage(id, messages.NewMessage(msg.GetSeq(), messages.ActionNotifySuccess, nil))

	return
}

func (a *Authenticator) updateClient(dc DefaultClient, authCredentials *ClientAuthCredentials) (ID, error) {

	dc.SetCredentials(authCredentials)

	newID := NewID("", authCredentials.UserID, authCredentials.DeviceID)
	err := a.gateway.SetClientID(dc.GetInfo().ID, newID)
	if IsIDAlreadyExist(err) {
		tempID, _ := GenTempID("")
		err = a.gateway.SetClientID(newID, tempID)
		if err != nil {
			return "", err
		}
		kickOut := messages.NewMessage(0, messages.ActionNotifyKickOut, &messages.KickOutNotify{
			DeviceName: authCredentials.DeviceName,
			DeviceId:   authCredentials.DeviceID,
		})
		_ = a.gateway.EnqueueMessage(tempID, kickOut)
		err = a.gateway.SetClientID(dc.GetInfo().ID, newID)
		if err != nil {
			return "", err
		}
	}
	return "", err
}
