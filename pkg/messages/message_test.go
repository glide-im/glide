package messages

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGlideMessage_Decode(t *testing.T) {
	cm := AckMessage{
		Mid: 1,
		Seq: 2,
	}
	message := NewMessage(1, ActionHeartbeat, &cm)
	bytes, err := JsonCodec.Encode(message)
	assert.Nil(t, err)

	m := NewEmptyMessage()
	err = JsonCodec.Decode(bytes, m)
	assert.Nil(t, err)

	assert.Equal(t, m.Action, message.Action)
}

func TestData_MarshalJSON(t *testing.T) {

	data := NewData("foo")
	encode, err := JsonCodec.Encode(data)
	assert.Nil(t, err)

	d := Data{}
	err = JsonCodec.Decode(encode, &d)
	assert.Nil(t, err)

	var s string
	err = d.Deserialize(&s)
	assert.Nil(t, err)

	assert.Equal(t, s, data.des)
}
