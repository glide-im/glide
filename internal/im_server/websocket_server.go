package im_server

import "github.com/glide-im/glide/pkg/conn"

type websocketServer struct {
	server conn.WsServer
}

func (s *websocketServer) Run(handler func(conn conn.Connection)) error {
	s.server.SetConnHandler(handler)

	return s.server.Run("", 1)
}
