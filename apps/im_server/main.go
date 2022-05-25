package main

import (
	"github.com/glide-im/glide/internal/im_server"
	"github.com/glide-im/glide/internal/message_handler"
	"github.com/glide-im/glide/internal/message_store_db"
	"github.com/glide-im/glide/internal/subscription"
	"github.com/glide-im/glide/pkg/bootstrap"
)

func main() {

	gateway, err := im_server.NewServer("0.0.0.0", 9090)
	if err != nil {
		panic(err)
	}

	handler, err := message_handler.NewHandler(message_store_db.New())
	if err != nil {
		panic(err)
	}

	options := bootstrap.Options{
		Messaging:    handler,
		Gate:         gateway,
		Subscription: subscription.NewSubscription(),
	}

	err = bootstrap.Bootstrap(&options)

	if err != nil {
		panic(err)
	}
}
