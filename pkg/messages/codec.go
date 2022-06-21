package messages

import (
	"encoding/json"
	"errors"
	"google.golang.org/protobuf/proto"
	"strings"
)

var ProtoBuffCodec = protobufCodec{}
var JsonCodec = jsonCodec{}
var DefaultCodec = JsonCodec

var errDecode = "message decode error: "

func IsDecodeError(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), errDecode)
}

type Codec interface {
	Decode(data []byte, i interface{}) error
	Encode(i interface{}) ([]byte, error)
}

type protobufCodec struct {
}

func (p protobufCodec) Decode(data []byte, i interface{}) error {
	message, ok := i.(proto.Message)
	if !ok {
		return errors.New("illegal argument, not implement proto.GlideMessage")
	}
	return proto.Unmarshal(data, message)
}

func (p protobufCodec) Encode(i interface{}) ([]byte, error) {
	message, ok := i.(proto.Message)
	if !ok {
		return nil, errors.New("illegal argument, not implement proto.GlideMessage")
	}
	return proto.Marshal(message)
}

type jsonCodec struct {
}

func (j jsonCodec) Decode(data []byte, i interface{}) error {
	err := json.Unmarshal(data, i)
	if err != nil {
		return errors.New(errDecode + err.Error())
	}
	return nil
}

func (j jsonCodec) Encode(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

type GlideProtocol struct {
}
