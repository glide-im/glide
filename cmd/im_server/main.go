package main

import (
	"github.com/glide-im/glide/internal/config"
	"github.com/glide-im/glide/internal/im_server"
	"github.com/glide-im/glide/internal/store_db"
	"github.com/glide-im/glide/pkg/auth/jwt_auth"
	"github.com/glide-im/glide/pkg/bootstrap"
	"github.com/glide-im/glide/pkg/messaging/message_handler"
	"github.com/glide-im/glide/pkg/subscription/subscription_impl"
)

func main() {

	config.MustLoad()

	gateway, err := im_server.NewServer(config.WsServer.ID, config.WsServer.Addr, config.WsServer.Port)
	if err != nil {
		panic(err)
	}

	auth := jwt_auth.NewAuthorizeImpl(config.WsServer.JwtSecret)
	dbStore, err := store_db.New(config.MySql)
	if err != nil {
		panic(err)
	}
	handler, err := message_handler.NewHandler(dbStore, auth)
	if err != nil {
		panic(err)
	}

	store := &store_db.SubscriptionStore{}
	options := bootstrap.Options{
		Messaging:    handler,
		Gate:         gateway,
		Subscription: subscription_impl.NewSubscription(store, store),
	}

	err = bootstrap.Bootstrap(&options)

	if err != nil {
		panic(err)
	}
}
