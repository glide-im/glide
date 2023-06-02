package messages

import (
	"errors"
	"fmt"
	"reflect"
)

var messageVersion int64 = 1

// GlideMessage common data of all message
type GlideMessage struct {
	Ver    int64  `json:"ver,omitempty"`
	Seq    int64  `json:"seq,omitempty"`
	Action string `json:"action"`
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
	Data   *Data  `json:"data,omitempty"`

	Extra map[string]string `json:"extra,omitempty"`
}

func NewMessage(seq int64, action Action, data interface{}) *GlideMessage {
	return &GlideMessage{
		Ver:    messageVersion,
		Seq:    seq,
		Action: string(action),
		Data:   NewData(data),
		Extra:  nil,
	}
}

func NewEmptyMessage() *GlideMessage {
	return &GlideMessage{
		Ver:   messageVersion,
		Data:  nil,
		Extra: nil,
	}
}

func (g *GlideMessage) GetSeq() int64 {
	return g.Seq
}

func (g *GlideMessage) GetAction() Action {
	return Action(g.Action)
}

func (g *GlideMessage) SetSeq(seq int64) {
	g.Seq = seq
}

func (g *GlideMessage) String() string {
	if g == nil {
		return "<nil>"
	}
	return fmt.Sprintf("&Message{Ver:%d, Action:%s, Data:%s}", g.Ver, g.Action, g.Data)
}

// Data used to wrap message data.
// Server received a message, the data type is []byte, it's waiting for deserialize to specified struct.
// When server push a message to client, the data type is specific struct.
type Data struct {
	des interface{}
}

func NewData(d interface{}) *Data {
	return &Data{
		des: d,
	}
}

func (d *Data) UnmarshalJSON(bytes []byte) error {
	d.des = bytes
	return nil
}

func (d *Data) MarshalJSON() ([]byte, error) {
	bytes, ok := d.des.([]byte)
	if ok {
		return bytes, nil
	}
	return JsonCodec.Encode(d.des)
}

func (d *Data) GetData() interface{} {
	return d.des
}

func (d *Data) Deserialize(i interface{}) error {
	if d == nil {
		return errors.New("data is nil")
	}
	s, ok := d.des.([]byte)
	if ok {
		return JsonCodec.Decode(s, i)
	} else {
		t1 := reflect.TypeOf(i)
		t2 := reflect.TypeOf(d.des)
		if t1 == t2 {
			reflect.ValueOf(i).Elem().Set(reflect.ValueOf(d.des).Elem())
			return nil
		}
	}
	return errors.New("invalid data")
}

func (d *Data) String() string {
	b, ok := d.des.([]byte)
	var s interface{}
	if ok {
		s = string(b)
	} else {
		if d.des == nil {
			s = "<nil>"
		} else {
			s = d.des
		}
	}
	return fmt.Sprintf("%s", s)
}
