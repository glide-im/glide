package gateway

import (
	"errors"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/panjf2000/ants/v2"
	"log"
	"sync"
)

type Options struct {
	MaxMessageConcurrency int
}

var _ gate.Gateway = (*Impl)(nil)

type Impl struct {
	gate.Gateway

	id string

	// clients is a map of all connected clients
	clients map[gate.ID]gate.Client
	mu      sync.RWMutex

	// msgHandler client message handler
	msgHandler gate.MessageHandler

	// pool of ants, used to process messages concurrently.
	pool *ants.Pool
}

func NewServer(options *Options) (*Impl, error) {

	ret := new(Impl)
	ret.clients = map[gate.ID]gate.Client{}
	ret.mu = sync.RWMutex{}

	pool, err := ants.NewPool(options.MaxMessageConcurrency,
		ants.WithNonblocking(true),
		ants.WithPanicHandler(func(i interface{}) {
			log.Printf("panic: %v", i)
		}),
		ants.WithPreAlloc(false),
	)
	if err != nil {
		return nil, err
	}
	ret.pool = pool
	return ret, nil
}

func (c *Impl) AddClient(cs gate.Client) {
	id := cs.GetInfo().ID
	c.clients[id] = cs
}

func (c *Impl) SetClientID(oldID, newID gate.ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	oldID.SetGateway(c.id)
	newID.SetGateway(c.id)

	cli, ok := c.clients[oldID]
	if !ok || cli == nil {
		return errors.New(errClientNotExist)
	}
	cliLogged, logged := c.clients[newID]
	if logged && cliLogged != nil {
		return errors.New(errClientAlreadyExist)
	}

	cli.SetID(newID)
	delete(c.clients, oldID)
	c.clients[newID] = cli
	return nil
}

func (c *Impl) ExitClient(id gate.ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	id.SetGateway(c.id)

	cli, ok := c.clients[id]
	if !ok || cli == nil {
		return errors.New(errClientNotExist)
	}

	delete(c.clients, id)
	cli.Exit()

	return nil
}

// EnqueueMessage to the client with the specified uid and device, device: pass 0 express all device.
func (c *Impl) EnqueueMessage(id gate.ID, msg *messages.GlideMessage) error {

	c.mu.RLock()
	defer c.mu.RUnlock()

	id.SetGateway(c.id)
	cli, ok := c.clients[id]
	if !ok || cli == nil {
		return errors.New(errClientNotExist)
	}

	return c.enqueueMessage(cli, msg)
}

func (c *Impl) enqueueMessage(cli gate.Client, msg *messages.GlideMessage) error {
	if !cli.IsRunning() {
		return errors.New(errClientClosed)
	}
	err := c.pool.Submit(func() {
		_ = cli.EnqueueMessage(msg)
	})
	if err != nil {
		return errors.New("enqueue message to client failed")
	}
	return nil
}

func (c *Impl) IsOnline(id gate.ID) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	id.SetGateway(c.id)
	return c.clients[id] != nil
}
