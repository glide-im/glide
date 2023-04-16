package server

import (
	"context"
	"github.com/glide-im/glide/im_service/proto"
)

type GatewayRpcServer interface {
	UpdateClient(ctx context.Context, request *proto.UpdateClient, response *proto.Response) error

	EnqueueMessage(ctx context.Context, request *proto.EnqueueMessageRequest, response *proto.Response) error
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
