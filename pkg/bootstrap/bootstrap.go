package bootstrap

import (
	"errors"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/messaging"
	"github.com/glide-im/glide/pkg/subscription"
)

type Options struct {
	Messaging    messaging.Interface
	Gate         gate.Interface
	Subscription subscription.Interface
}

func Bootstrap(opts *Options) error {

	_, ok := opts.Gate.(gate.Server)
	if ok {
		return bootGatewayServer(opts)
	}
	_, ok = opts.Messaging.(messaging.Server)
	if ok {
		return bootMessagingServer(opts)
	}
	_, ok = opts.Subscription.(subscription.Server)
	if ok {
		return bootSubscriptionServer(opts)
	}

	return errors.New("no server found")
}

func bootSubscriptionServer(opts *Options) error {
	server, ok := opts.Subscription.(subscription.Server)
	if !ok {
		return errors.New("subscription server not implemented")
	}
	server.SetGateInterface(opts.Gate)
	return server.Run()
}

func bootMessagingServer(opts *Options) error {
	server, ok := opts.Messaging.(messaging.Server)
	if !ok {
		return errors.New("messaging does not implement Messaging.impl")
	}

	manager, ok := opts.Gate.(gate.Gateway)
	if ok {
		server.SetGate(manager)
	}
	server.SetSubscription(opts.Subscription)
	return server.Run()
}

func bootGatewayServer(opts *Options) error {

	gateway, ok := opts.Gate.(gate.Server)
	if !ok {
		return errors.New("Gate is not a gateway server")
	}

	if opts.Messaging == nil {
		return errors.New("can't boot a gateway server without a Messaging interface")
	}

	gateway.SetMessageHandler(func(cliInfo *gate.Info, message *messages.GlideMessage) {
		err := opts.Messaging.Handle(cliInfo, message)
		if err != nil {
			// TODO: Log error
		}
	})

	return gateway.Run()
}
