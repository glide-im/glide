package client

import (
	"context"
	"github.com/glide-im/glide/pkg/rpc"
	"github.com/glide-im/im-service/pkg/proto"
)

type subscriptionRpcClient struct {
	cli *rpc.BaseClient
}

func (s *subscriptionRpcClient) Close() error {
	return s.cli.Close()
}

func (s *subscriptionRpcClient) Subscribe(ctx context.Context, request *proto.SubscribeRequest, response *proto.Response) error {
	return s.cli.Call(ctx, "Subscribe", request, response)
}

func (s *subscriptionRpcClient) UnSubscribe(ctx context.Context, request *proto.UnsubscribeRequest, response *proto.Response) error {
	return s.cli.Call(ctx, "UnSubscribe", request, response)
}

func (s *subscriptionRpcClient) UpdateSubscriber(ctx context.Context, request *proto.UpdateSubscriberRequest, response *proto.Response) error {
	return s.cli.Call(ctx, "UpdateSubscriber", request, response)
}

func (s *subscriptionRpcClient) RemoveChannel(ctx context.Context, request *proto.RemoveChannelRequest, response *proto.Response) error {
	return s.cli.Call(ctx, "RemoveChannel", request, response)
}

func (s *subscriptionRpcClient) CreateChannel(ctx context.Context, request *proto.CreateChannelRequest, response *proto.Response) error {
	return s.cli.Call(ctx, "CreateChannel", request, response)
}

func (s *subscriptionRpcClient) UpdateChannel(ctx context.Context, request *proto.UpdateChannelRequest, response *proto.Response) error {
	return s.cli.Call(ctx, "UpdateChannel", request, response)
}

func (s *subscriptionRpcClient) Publish(ctx context.Context, request *proto.PublishRequest, response *proto.Response) error {
	return s.cli.Call(ctx, "Publish", request, response)
}
