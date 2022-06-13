package gateway

import (
	"github.com/glide-im/glide/pkg/conn"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func mockMsgHandler(cliInfo *gate.Info, message *messages.GlideMessage) {
}

func TestClient_RunReadHeartbeatTimeout(t *testing.T) {
	fn, ch := mockReadFn()
	go func() {
		time.Sleep(time.Millisecond * 100)
		ch <- messages.NewMessage(1, messages.ActionHeartbeat, nil)
		time.Sleep(time.Millisecond * 100)
		ch <- messages.NewMessage(1, messages.ActionHeartbeat, nil)
		time.Sleep(time.Millisecond * 100)
		ch <- messages.NewMessage(1, messages.ActionHeartbeat, nil)
	}()

	client := NewClient(&mockConnection{
		writeDelayMilliSec: 200,
		mockRead:           fn,
	}, mockGateway{}, mockMsgHandler)

	client.config.ClientHeartbeatDuration = time.Millisecond * 200
	client.Run()

	time.Sleep(time.Second * 1)
	client.Exit()
}

func TestClient_RunServerHeartbeat(t *testing.T) {
	fn, _ := mockReadFn()
	client := NewClientWithConfig(&mockConnection{
		writeDelayMilliSec: 100,
		mockRead:           fn,
	}, mockGateway{}, mockMsgHandler, &ClientConfig{
		ServerHeartbeatDuration: time.Millisecond * 400,
		CloseImmediately:        true,
	})
	client.Run()

	time.Sleep(time.Second * 1)
	client.Exit()
}

func TestClient_ExitImmediately(t *testing.T) {

	fn, ch := mockReadFn()
	go func() {
		time.Sleep(time.Second * 1)
		ch <- messages.NewMessage(1, messages.ActionHeartbeat, nil)
	}()

	client := NewClient(&mockConnection{
		writeDelayMilliSec: 200,
		mockRead:           fn,
	}, mockGateway{}, mockMsgHandler)
	client.config.CloseImmediately = true
	client.Run()

	for i := 0; i < 10; i++ {
		err := client.EnqueueMessage(messages.NewMessage(1, messages.ActionHeartbeat, nil))
		if err != nil {
			t.Error(err)
		}
	}
	time.Sleep(time.Second * 1)
	client.Exit()

	assert.False(t, client.IsRunning())
	assert.Error(t, client.EnqueueMessage(messages.NewMessage(1, messages.ActionHeartbeat, nil)))
	assert.Equal(t, client.state, stateClosed)
}

func mockReadFn() (func() ([]byte, error), chan<- *messages.GlideMessage) {
	ch := make(chan *messages.GlideMessage)
	return func() ([]byte, error) {
		m := <-ch
		encode, err := messages.JsonCodec.Encode(m)
		return encode, err
	}, ch
}

type mockConnection struct {
	writeDelayMilliSec time.Duration
	mockRead           func() ([]byte, error)
}

func (m *mockConnection) Write(data []byte) error {
	time.Sleep(time.Millisecond * m.writeDelayMilliSec)
	log.Println("write:", string(data))
	return nil
}

func (m *mockConnection) Read() ([]byte, error) {
	for {
		return m.mockRead()
	}
}

func (m *mockConnection) Close() error {
	log.Println("close connection")
	return nil
}

func (m *mockConnection) GetConnInfo() *conn.ConnectionInfo {
	return &conn.ConnectionInfo{
		Ip:   "127.0.0.1",
		Port: 9999,
		Addr: "[::1]:9999",
	}
}

type mockGateway struct {
}

func (m mockGateway) SetClientID(old gate.ID, new_ gate.ID) error {
	return nil
}

func (m mockGateway) ExitClient(id gate.ID) error {
	log.Println("exit client:", id)
	return nil
}

func (m mockGateway) IsOnline(id gate.ID) bool {
	return true
}

func (m mockGateway) EnqueueMessage(id gate.ID, message *messages.GlideMessage) error {
	return nil
}
