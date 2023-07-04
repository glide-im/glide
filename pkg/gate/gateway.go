package gate

import (
	"errors"
	"github.com/glide-im/glide/pkg/conn"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/panjf2000/ants/v2"
	"log"
	"sync"
	"time"
)

// Gateway is the basic and common interface for all gate implementations.
// As the basic gate, it is used to provide a common gate interface for other modules to interact with the gate.
type Gateway interface {

	// SetClientID sets the client id with the new id.
	SetClientID(old ID, new_ ID) error

	UpdateClient(id ID, info *ClientSecrets) error

	// ExitClient exits the client with the given id.
	ExitClient(id ID) error

	// EnqueueMessage enqueues the message to the client with the given id.
	EnqueueMessage(id ID, message *messages.GlideMessage) error
}

// Server is the interface for the gateway server, which is used to handle and manager client connections.
type Server interface {
	Gateway

	// SetMessageHandler sets the client message handler.
	SetMessageHandler(h MessageHandler)

	// HandleConnection handles the new client connection and returns the random and temporary id set for the connection.
	HandleConnection(c conn.Connection) ID

	Run() error
}

// MessageHandler used to handle messages from the gate.
type MessageHandler func(cliInfo *Info, message *messages.GlideMessage)

// DefaultGateway is gateway default implements.
type DefaultGateway interface {
	Gateway

	GetClient(id ID) Client

	GetAll() map[ID]Info

	SetMessageHandler(h MessageHandler)

	AddClient(cs Client)
}

type Options struct {
	// ID is the gateway id.
	ID string
	// SecretKey is the secret key used to encrypt and decrypt authentication token.
	SecretKey string
	// MaxMessageConcurrency is the max message concurrency.
	MaxMessageConcurrency int
}

var _ DefaultGateway = (*Impl)(nil)

type Impl struct {
	id string

	// clients is a map of all connected clients
	clients map[ID]Client
	mu      sync.RWMutex

	// msgHandler client message handler
	msgHandler MessageHandler

	authenticator *Authenticator

	// pool of ants, used to process messages concurrently.
	pool *ants.Pool

	emptyInfo *Info
}

func NewServer(options *Options) (*Impl, error) {

	ret := new(Impl)
	ret.clients = map[ID]Client{}
	ret.mu = sync.RWMutex{}
	ret.id = options.ID
	ret.emptyInfo = &Info{
		ID: NewID(ret.id, "", ""),
	}

	if options.SecretKey != "" {
		ret.authenticator = NewAuthenticator(ret, options.SecretKey)
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
func (c *Impl) GetClient(id ID) Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.clients[id]
}

// GetAll returns all clients in the gateway.
func (c *Impl) GetAll() map[ID]Info {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := map[ID]Info{}
	for id, client := range c.clients {
		result[id] = client.GetInfo()
	}
	return result
}

func (c *Impl) SetMessageHandler(h MessageHandler) {
	c.msgHandler = h
}

func (c *Impl) UpdateClient(id ID, info *ClientSecrets) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	id.SetGateway(c.id)

	cli, ok := c.clients[id]
	if !ok || cli == nil {
		return errors.New(errClientNotExist)
	}

	dc, ok := cli.(DefaultClient)
	if ok {
		credentials := dc.GetCredentials()
		credentials.Secrets = info
		dc.SetCredentials(credentials)
		logger.D("gateway", "update client %s, %v", id, info.MessageDeliverSecret)
	}

	return nil
}

func (c *Impl) AddClient(cs Client) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := cs.GetInfo().ID
	id.SetGateway(c.id)

	dc, ok := cs.(DefaultClient)
	if ok {
		dc.AddMessageInterceptor(c.interceptClientMessage)
	}

	c.clients[id] = cs
	c.msgHandler(c.emptyInfo, messages.NewMessage(0, messages.ActionInternalOnline, id))
}

// SetClientID replace the oldID with newID of the client.
// If the oldID is not exist, return errClientNotExist.
// If the newID is existed, return errClientAlreadyExist.
func (c *Impl) SetClientID(oldID, newID ID) error {
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
func (c *Impl) ExitClient(id ID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	id.SetGateway(c.id)

	cli, ok := c.clients[id]
	if !ok || cli == nil {
		return errors.New(errClientNotExist)
	}

	cli.SetID("")
	delete(c.clients, id)
	c.msgHandler(c.emptyInfo, messages.NewMessage(0, messages.ActionInternalOffline, id))
	cli.Exit()

	return nil
}

// EnqueueMessage to the client with the specified id.
func (c *Impl) EnqueueMessage(id ID, msg *messages.GlideMessage) error {

	c.mu.RLock()
	defer c.mu.RUnlock()

	id.SetGateway(c.id)
	cli, ok := c.clients[id]
	if !ok || cli == nil {
		return errors.New(errClientNotExist)
	}

	return c.enqueueMessage(cli, msg)
}

func (c *Impl) interceptClientMessage(dc DefaultClient, m *messages.GlideMessage) bool {

	if m.Action == messages.ActionAuthenticate {
		if c.authenticator != nil {
			return c.authenticator.ClientAuthMessageInterceptor(dc, m)
		}
	}

	return c.authenticator.MessageInterceptor(dc, m)
}

func (c *Impl) enqueueMessage(cli Client, msg *messages.GlideMessage) error {
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

type WebsocketGatewayServer struct {
	gateId    string
	addr      string
	port      int
	server    conn.Server
	decorator DefaultGateway
	h         MessageHandler
}

func NewWebsocketServer(gateId string, addr string, port int, secretKey string) *WebsocketGatewayServer {
	srv := WebsocketGatewayServer{}
	srv.decorator, _ = NewServer(
		&Options{
			ID:                    gateId,
			MaxMessageConcurrency: 30_0000,
			SecretKey:             secretKey,
		},
	)
	srv.addr = addr
	srv.port = port
	srv.gateId = gateId
	options := &conn.WsServerOptions{
		ReadTimeout:  time.Minute * 3,
		WriteTimeout: time.Minute * 3,
	}
	srv.server = conn.NewWsServer(options)
	return &srv
}

func (w *WebsocketGatewayServer) SetMessageHandler(h MessageHandler) {
	w.h = h
	w.decorator.SetMessageHandler(h)
}

func (w *WebsocketGatewayServer) HandleConnection(c conn.Connection) ID {
	// 获取一个临时 uid 标识这个连接
	id, err := GenTempID(w.gateId)
	if err != nil {
		logger.E("[gateway] gen temp id error: %v", err)
		return ""
	}
	ret := NewClientWithConfig(c, w, w.h, &ClientConfig{
		HeartbeatLostLimit:      3,
		ClientHeartbeatDuration: time.Second * 30,
		ServerHeartbeatDuration: time.Second * 30,
		CloseImmediately:        false,
	})
	ret.SetID(id)
	w.decorator.AddClient(ret)

	// 开始处理连接的消息
	ret.Run()

	hello := messages.ServerHello{
		TempID:            id.UID(),
		HeartbeatInterval: 30,
	}

	m := messages.NewMessage(0, messages.ActionHello, hello)
	_ = ret.EnqueueMessage(m)

	return id
}

func (w *WebsocketGatewayServer) Run() error {
	w.server.SetConnHandler(func(conn conn.Connection) {
		w.HandleConnection(conn)
	})
	return w.server.Run(w.addr, w.port)
}

func (w *WebsocketGatewayServer) GetClient(id ID) Client {
	return w.decorator.GetClient(id)
}

func (w *WebsocketGatewayServer) GetAll() map[ID]Info {
	return w.decorator.GetAll()
}

func (w *WebsocketGatewayServer) AddClient(cs Client) {
	w.decorator.AddClient(cs)
}

func (w *WebsocketGatewayServer) SetClientID(old ID, new_ ID) error {
	return w.decorator.SetClientID(old, new_)
}

func (w *WebsocketGatewayServer) UpdateClient(id ID, info *ClientSecrets) error {
	return w.decorator.UpdateClient(id, info)
}

func (w *WebsocketGatewayServer) ExitClient(id ID) error {
	return w.decorator.ExitClient(id)
}

func (w *WebsocketGatewayServer) EnqueueMessage(id ID, message *messages.GlideMessage) error {
	return w.decorator.EnqueueMessage(id, message)
}
