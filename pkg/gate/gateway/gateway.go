package gateway

import (
	"errors"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/panjf2000/ants/v2"
	"log"
	"sync"
)

// Interface is gateway default implements.
type Interface interface {
	gate.Gateway
	GetClient(id gate.ID) gate.Client
	GetAll() map[gate.ID]gate.Info
	SetMessageHandler(h gate.MessageHandler)
	AddClient(cs gate.Client)
}

type Options struct {
	// ID is the gateway id.
	ID string
	// MaxMessageConcurrency is the max message concurrency.
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

	emptyInfo *gate.Info
}

func NewServer(options *Options) (*Impl, error) {

	ret := new(Impl)
	ret.clients = map[gate.ID]gate.Client{}
	ret.mu = sync.RWMutex{}
	ret.id = options.ID
	ret.emptyInfo = &gate.Info{
		ID: gate.NewID(ret.id, "", ""),
	}

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

// GetClient returns the client with specified id
func (c *Impl) GetClient(id gate.ID) gate.Client {
	return c.clients[id]
}

// GetAll returns all clients in the gateway.
func (c *Impl) GetAll() map[gate.ID]gate.Info {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := map[gate.ID]gate.Info{}
	for id, client := range c.clients {
		result[id] = client.GetInfo()
	}
	return result
}

func (c *Impl) SetMessageHandler(h gate.MessageHandler) {
	c.msgHandler = h
}

func (c *Impl) AddClient(cs gate.Client) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := cs.GetInfo().ID
	id.SetGateway(c.id)

	c.clients[id] = cs
	c.msgHandler(nil, messages.NewMessage(0, messages.ActionInternalOnline, id))
}

// SetClientID replace the oldID with newID of the client.
// If the oldID is not exist, return errClientNotExist.
// If the newID is existed, return errClientAlreadyExist.
func (c *Impl) SetClientID(oldID, newID gate.ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	oldID.SetGateway(c.id)
	newID.SetGateway(c.id)

	cli, ok := c.clients[oldID]
	if !ok || cli == nil {
		return errors.New(errClientNotExist)
	}
	cliLogged, exist := c.clients[newID]
	if exist && cliLogged != nil {
		return errors.New(errClientAlreadyExist)
	}

	cli.SetID(newID)
	delete(c.clients, oldID)
	c.msgHandler(c.emptyInfo, messages.NewMessage(0, messages.ActionInternalOffline, oldID))
	c.msgHandler(c.emptyInfo, messages.NewMessage(0, messages.ActionInternalOnline, newID))

	c.clients[newID] = cli
	return nil
}

// ExitClient close the client with the specified id.
// If the client is not exist, return errClientNotExist.
func (c *Impl) ExitClient(id gate.ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	id.SetGateway(c.id)

	cli, ok := c.clients[id]
	if !ok || cli == nil {
		return errors.New(errClientNotExist)
	}

	delete(c.clients, id)
	c.msgHandler(c.emptyInfo, messages.NewMessage(0, messages.ActionInternalOffline, id))
	cli.Exit()

	return nil
}

// EnqueueMessage to the client with the specified id.
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
	client, ok := c.clients[id]
	return ok && client != nil && client.IsRunning()
}
