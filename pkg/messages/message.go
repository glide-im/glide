package messages

import (
	"encoding/json"
	"github.com/glide-im/glide/pkg/logger"
)

type Data struct {
	des interface{}
}

func NewData(d interface{}) *Data {
	return &Data{
		des: d,
	}
}

func (d *Data) Data() interface{} {
	return d.des
}

func (d *Data) UnmarshalJSON(bytes []byte) error {
	d.des = bytes
	return nil
}

func (d *Data) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.des)
}

func (d *Data) bytes() []byte {
	bytes, ok := d.des.([]byte)
	if ok {
		return bytes
	}
	marshalJSON, err := d.MarshalJSON()
	if err != nil {
		logger.E("message data marshal json error %v", err)
		return nil
	}
	return marshalJSON
}

func (d *Data) Deserialize(i interface{}) error {
	s, ok := d.des.([]byte)
	if ok {
		return json.Unmarshal(s, i)
	}
	return nil
}

type GlideMessage struct {
	Ver    int64
	Seq    int64
	Action string
	Data   *Data
	Extra  map[string]string
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

func (g *GlideMessage) DeserializeData(i interface{}) error {
	return g.Data.Deserialize(i)
}

func NewMessage(seq int64, action Action, data interface{}) *GlideMessage {
	return &GlideMessage{
		Ver:    0,
		Seq:    seq,
		Action: string(action),
		Data:   NewData(data),
	}
}

func NewEmptyMessage() *GlideMessage {
	return &GlideMessage{}
}
