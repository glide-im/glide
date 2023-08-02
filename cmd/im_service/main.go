package main

import (
	"github.com/glide-im/glide/config"
	"github.com/glide-im/glide/im_service/server"
	"github.com/glide-im/glide/internal/message_store_db"
	"github.com/glide-im/glide/internal/pkg/db"
	"github.com/glide-im/glide/internal/world_channel"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/messaging"
	"github.com/glide-im/glide/pkg/rpc"
	"github.com/glide-im/glide/pkg/store"
	"github.com/glide-im/glide/pkg/subscription/subscription_impl"
)

func main() {

	config.MustLoad()

	err := db.Init(nil, &db.RedisConfig{
		Host:     config.Redis.Host,
		Port:     config.Redis.Port,
		Password: config.Redis.Password,
		Db:       config.Redis.Db,
	})
	if err != nil {
		panic(err)
	}

	gateway := gate.NewWebsocketServer(
		config.WsServer.ID,
		config.WsServer.Addr,
		config.WsServer.Port,
		config.Common.SecretKey,
	)

	var cStore store.MessageStore = &message_store_db.IdleChatMessageStore{}
	var sStore store.SubscriptionStore = &message_store_db.IdleSubscriptionStore{}

	if config.Common.StoreMessageHistory {
		if config.Kafka != nil && len(config.Kafka.Address) != 0 {
			producer, err := store.NewKafkaProducer(config.Kafka.Address)
			if err != nil {
				panic(err)
			}
			cStore = producer
			sStore = producer
			logger.D("Kafka is configured, all message will push to kafka: %v", config.Kafka.Address)
		} else {
			dbStore, err := message_store_db.New(config.MySql)
			if err != nil {
				panic(err)
			}
			cStore = dbStore
			sStore = &message_store_db.SubscriptionMessageStore{}
		}

	} else {
		logger.D("Common.StoreMessageHistory is false, message history will not be stored")
	}

	handler, err := messaging.NewHandlerWithOptions(gateway, &messaging.MessageHandlerOptions{
		MessageStore:           cStore,
		DontInitDefaultHandler: false,
		NotifyOnErr:            true,
	})
	if err != nil {
		panic(err)
	}
	messaging.StoreOfflineMessage = config.Common.StoreOfflineMessage

	subscription := subscription_impl.NewSubscription(sStore, sStore)
	subscription.SetGateInterface(gateway)

	handler.SetSubscription(subscription)
	handler.SetGate(gateway)

	go func() {
		logger.D("websocket listening on %s:%d", config.WsServer.Addr, config.WsServer.Port)

		gateway.SetMessageHandler(func(cliInfo *gate.Info, message *messages.GlideMessage) {
			e := handler.Handle(cliInfo, message)
			if e != nil {
				logger.E("error: %v", e)
			}
		})

		err = gateway.Run()
		if err != nil {
			panic(err)
		}
	}()

	err = world_channel.EnableWorldChannel(subscription_impl.NewSubscribeWrap(subscription))
	if err != nil {
		panic(err)
	}

	rpcOpts := rpc.ServerOptions{
		Name:    config.IMService.Name,
		Network: config.IMService.Network,
		Addr:    config.IMService.Addr,
		Port:    config.IMService.Port,
	}
	logger.D("rpc %s listening on %s %s:%d", rpcOpts.Name, rpcOpts.Network, rpcOpts.Addr, rpcOpts.Port)
	err = server.RunRpcService(&rpcOpts, gateway, subscription)
	if err != nil {
		panic(err)
	}
}
