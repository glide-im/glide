package bootstrap

import (
	"errors"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/messaging"
	"github.com/glide-im/glide/pkg/subscription"
)

type options struct {
	messaging messaging.Interface
	gate      gate.Interface
	group     subscription.Interface
}

func bootGroupServer(opts options) error {
	server, ok := opts.group.(subscription.Server)
	if !ok {
		return errors.New("group server not implemented")
	}
	server.SetGate(opts.gate)
	return server.Run()
}

func bootMessagingServer(opts options) error {
	server, ok := opts.messaging.(messaging.Server)
	if !ok {
		return errors.New("messaging does not implement messaging.Server")
	}
	server.SetGate(opts.gate)
	server.SetGroup(opts.group)
	return server.Run()
}

func bootGatewayServer(opts options) error {

	gateway, ok := opts.gate.(gate.Server)
	if !ok {
		return errors.New("gate is not a gateway server")
	}

	if opts.messaging == nil {
		return errors.New("can't boot a gateway server without a messaging interface")
	}

	gateway.SetMessageHandler(func(cliInfo *gate.Info, message *messages.GlideMessage) {
		err := opts.messaging.Handle(cliInfo, message)
		if err != nil {
			// TODO: Log error
		}
	})

	return gateway.Run()
}
