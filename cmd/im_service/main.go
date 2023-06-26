package main

import (
	"flag"
	"github.com/glide-im/glide/config"
	"github.com/glide-im/glide/internal/action_handler"
	"github.com/glide-im/glide/internal/im_server"
	"github.com/glide-im/glide/internal/message_store_db"
	"github.com/glide-im/glide/internal/pkg/db"
	"github.com/glide-im/glide/internal/server_state"
	"github.com/glide-im/glide/internal/world_channel"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/messaging"
	"github.com/glide-im/glide/pkg/rpc"
	"github.com/glide-im/glide/pkg/store"
	"github.com/glide-im/glide/pkg/subscription/subscription_impl"
)

var state *string

func init() {
	state = flag.String("state", "", "show im server run state")
	flag.Parse()
}

func main() {

	if *state != "" {
		server_state.ShowServerState("localhost:9091")
		return
	}

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

	gateway, err := im_server.NewServer(
		config.WsServer.ID,
		config.WsServer.Addr,
		config.WsServer.Port,
		config.Common.SecretKey,
	)
	if err != nil {
		panic(err)
	}

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

	handler, err := messaging.NewHandlerWithOptions(&messaging.MessageHandlerOptions{
		MessageStore:           cStore,
		DontInitDefaultHandler: true,
		NotifyOnErr:            true,
	})
	if err != nil {
		panic(err)
	}
	if config.Common.StoreOfflineMessage {
		messaging.Enable = true
		handler.SetOfflineMessageHandler(messaging.GetHandleFn())
	}
	action_handler.Setup(handler)
	handler.InitDefaultHandler(nil)
	handler.SetGate(gateway)

	subscription := subscription_impl.NewSubscription(sStore, sStore)
	subscription.SetGateInterface(gateway)

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

	go func() {
		logger.D("state server is listening on 0.0.0.0:%d", 9091)
		server_state.StartSrv(9091, gateway)
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
	err = im_server.RunRpcServer(&rpcOpts, gateway, subscription)
	if err != nil {
		panic(err)
	}
}
