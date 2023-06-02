package im_server

import (
	"github.com/glide-im/glide/pkg/conn"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/rcrowley/go-metrics"
	"time"
)

type GatewayMetrics struct {
	ServerId     string
	Addr         string
	Port         int
	StartAt      time.Time
	RunningHours float64
	Message      *MessageMetrics
	Conn         *ConnectionMetrics
}

type GatewayServer struct {
	*gate.Impl

	server conn.Server
	h      gate.MessageHandler

	gateID string
	addr   string
	port   int

	metrics *GatewayMetrics
}

func NewServer(id string, addr string, port int) (*GatewayServer, error) {
	srv := GatewayServer{}
	srv.Impl, _ = gate.NewServer(
		&gate.Options{
			ID:                    id,
			MaxMessageConcurrency: 30_0000,
		},
	)
	srv.metrics = &GatewayMetrics{
		ServerId: id,
		Addr:     addr,
		Port:     port,
		Message:  NewMessageMetrics(),
		Conn:     NewConnectionMetrics(),
	}
	srv.addr = addr
	srv.port = port
	srv.gateID = id

	sample := metrics.NewExpDecaySample(1024, 0.015)
	histogram := metrics.NewHistogram(sample)
	_ = metrics.Register("s", histogram)
	histogram.Update(1)

	options := &conn.WsServerOptions{
		ReadTimeout:  time.Minute * 3,
		WriteTimeout: time.Minute * 3,
	}
	srv.server = conn.NewWsServer(options)
	return &srv, nil
}

func (c *GatewayServer) Run() error {

	c.metrics.StartAt = time.Now()

	c.server.SetConnHandler(func(conn conn.Connection) {
		c.HandleConnection(conn)
		c.metrics.Conn.Connected()
	})
	return c.server.Run(c.addr, c.port)
}

func (c *GatewayServer) SetMessageHandler(h gate.MessageHandler) {
	handler := func(id *gate.Info, msg *messages.GlideMessage) {
		c.metrics.Message.In()
		h(id, msg)
	}
	c.h = handler
	c.Impl.SetMessageHandler(handler)
}

// HandleConnection 当一个用户连接建立后, 由该方法创建 UserClient 实例 UserClient 并管理该连接, 返回该由连接创建客户端的标识 id
// 返回的标识 id 是一个临时 id, 后续连接认证后会改变
func (c *GatewayServer) HandleConnection(conn conn.Connection) gate.ID {

	// 获取一个临时 uid 标识这个连接
	id, err := gate.GenTempID(c.gateID)
	if err != nil {
		logger.E("[gateway] gen temp id error: %v", err)
		return ""
	}
	ret := gate.NewClientWithConfig(conn, c, c.h, &gate.ClientConfig{
		HeartbeatLostLimit:      3,
		ClientHeartbeatDuration: time.Second * 30,
		ServerHeartbeatDuration: time.Second * 30,
		CloseImmediately:        false,
	})
	ret.SetID(id)
	c.Impl.AddClient(ret)

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

func (c *GatewayServer) EnqueueMessage(id gate.ID, msg *messages.GlideMessage) error {
	err := c.Impl.EnqueueMessage(id, msg)
	if err != nil {
		c.metrics.Message.Out()
	} else {
		c.metrics.Message.OutFailed()
	}
	return err
}

func (c *GatewayServer) SetClientID(oldID, newID gate.ID) error {
	err := c.Impl.SetClientID(oldID, newID)
	if err == nil {
		c.metrics.Conn.Login()
	}
	return err
}

func (c *GatewayServer) ExitClient(id gate.ID) error {
	id.SetGateway(c.gateID)

	client := c.Impl.GetClient(id)
	if client != nil {
		c.metrics.Conn.Exit(client.GetInfo())
	}

	err := c.Impl.ExitClient(id)
	return err
}

func (c *GatewayServer) GetState() GatewayMetrics {
	span := time.Now().Unix() - c.metrics.StartAt.Unix()
	c.metrics.RunningHours = float64(span) / 60.0 / 60.0
	return *c.metrics
}
