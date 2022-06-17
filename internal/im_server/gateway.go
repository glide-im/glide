package im_server

import (
	"github.com/glide-im/glide/pkg/conn"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/gate/gateway"
	"github.com/glide-im/glide/pkg/logger"
	"time"
)

type GatewayServer struct {
	*gateway.Impl

	server conn.Server
	h      gate.MessageHandler

	gateID string
	addr   string
	port   int
}

func NewServer(id string, addr string, port int) (gate.Server, error) {
	srv := GatewayServer{}
	srv.Impl, _ = gateway.NewServer(
		&gateway.Options{MaxMessageConcurrency: 30_0000},
	)
	srv.addr = addr
	srv.port = port
	srv.gateID = id

	options := &conn.WsServerOptions{
		ReadTimeout:  time.Minute * 3,
		WriteTimeout: time.Minute * 3,
	}
	srv.server = conn.NewWsServer(options)
	return &srv, nil
}

func (c *GatewayServer) Run() error {
	c.server.SetConnHandler(func(conn conn.Connection) {
		c.HandleConnection(conn)
	})
	return c.server.Run(c.addr, c.port)
}

func (c *GatewayServer) SetMessageHandler(h gate.MessageHandler) {
	c.h = h
	c.Impl.SetMessageHandler(h)
}

// HandleConnection 当一个用户连接建立后, 由该方法创建 Client 实例 Client 并管理该连接, 返回该由连接创建客户端的标识 id
// 返回的标识 id 是一个临时 id, 后续连接认证后会改变
func (c *GatewayServer) HandleConnection(conn conn.Connection) gate.ID {

	// 获取一个临时 uid 标识这个连接
	id, err := gate.GenTempID(c.gateID)
	if err != nil {
		logger.E("[gateway] gen temp id error: %v", err)
		return ""
	}
	ret := gateway.NewClientWithConfig(conn, c, c.h, &gateway.ClientConfig{
		HeartbeatLostLimit:      4,
		ClientHeartbeatDuration: time.Second * 30,
		ServerHeartbeatDuration: time.Second * 30,
		CloseImmediately:        false,
	})
	ret.SetID(id)
	c.Impl.AddClient(ret)

	// 开始处理连接的消息
	ret.Run()
	return id
}
