package server

import (
	"context"
	"github.com/glide-im/glide/im_service/proto"
)

type GatewayRpcServer interface {
	SetClientID(ctx context.Context, request *proto.SetIDRequest, response *proto.Response) error

	ExitClient(ctx context.Context, request *proto.ExitClientRequest, response *proto.Response) error

	IsOnline(ctx context.Context, request *proto.IsOnlineRequest, response *proto.IsOnlineResponse) error

	EnqueueMessage(ctx context.Context, request *proto.EnqueueMessageRequest, response *proto.Response) error

	Status(ctx context.Context, request interface{}, response interface{}) error

	// TODO 查询单个连接信息
}

type SubscriptionRpcServer interface {
	Subscribe(ctx context.Context, request *proto.SubscribeRequest, response *proto.Response) error

	UnSubscribe(ctx context.Context, request *proto.UnsubscribeRequest, response *proto.Response) error

	UpdateSubscriber(ctx context.Context, request *proto.UpdateSubscriberRequest, response *proto.Response) error

	RemoveChannel(ctx context.Context, request *proto.RemoveChannelRequest, response *proto.Response) error

	CreateChannel(ctx context.Context, request *proto.CreateChannelRequest, response *proto.Response) error

	UpdateChannel(ctx context.Context, request *proto.UpdateChannelRequest, response *proto.Response) error

	Publish(ctx context.Context, request *proto.PublishRequest, response *proto.Response) error
}
