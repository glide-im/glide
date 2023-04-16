package client

import (
	"context"
	"github.com/glide-im/glide/im_service/proto"
	"github.com/glide-im/glide/im_service/server"
	"github.com/glide-im/glide/pkg/rpc"
)

var _ server.GatewayRpcServer = &GatewayRpcClient{}

type GatewayRpcClient struct {
	cli *rpc.BaseClient
}

func (I *GatewayRpcClient) Status(ctx context.Context, request interface{}, response interface{}) error {
	return I.cli.Call(ctx, "Status", request, response)
}

func (I *GatewayRpcClient) SetClientID(ctx context.Context, request *proto.SetIDRequest, response *proto.Response) error {
	return I.cli.Call(ctx, "SetClientID", request, response)
}

func (I *GatewayRpcClient) ExitClient(ctx context.Context, request *proto.ExitClientRequest, response *proto.Response) error {
	return I.cli.Call(ctx, "ExitClient", request, response)
}

func (I *GatewayRpcClient) IsOnline(ctx context.Context, request *proto.IsOnlineRequest, response *proto.IsOnlineResponse) error {
	return I.cli.Call(ctx, "IsOnline", request, response)
}

func (I *GatewayRpcClient) EnqueueMessage(ctx context.Context, request *proto.EnqueueMessageRequest, response *proto.Response) error {
	return I.cli.Call(ctx, "EnqueueMessage", request, response)
}
