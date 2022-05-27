package im_server

import (
	"github.com/glide-im/glide/internal/gateway"
	"github.com/glide-im/glide/pkg/conn"
	"github.com/glide-im/glide/pkg/gate"
	"time"
)

type GatewayServer struct {
	*gateway.Impl

	server conn.Server
	h      gate.MessageHandler

	addr string
	port int
}

func NewServer(addr string, port int) (gate.Server, error) {
	srv := GatewayServer{}
	srv.Impl, _ = gateway.NewServer()
	srv.addr = addr
	srv.port = port

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
}

// HandleConnection 当一个用户连接建立后, 由该方法创建 Client 实例 Client 并管理该连接, 返回该由连接创建客户端的标识 id
// 返回的标识 id 是一个临时 id, 后续连接认证后会改变
func (c *GatewayServer) HandleConnection(conn conn.Connection) gate.ID {

	// 获取一个临时 uid 标识这个连接
	id := gate.NewID("", "0", "0")
	ret := gateway.NewClient(conn, c, c.h)
	ret.SetID(id)
	c.Impl.AddClient(ret)

	// 开始处理连接的消息
	ret.Run()
	return id
}
