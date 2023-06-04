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

func (I *GatewayRpcClient) UpdateClient(ctx context.Context, request *proto.UpdateClient, response *proto.Response) error {
	return I.cli.Call(ctx, "UpdateClient", request, response)
}

func (I *GatewayRpcClient) EnqueueMessage(ctx context.Context, request *proto.EnqueueMessageRequest, response *proto.Response) error {
	return I.cli.Call(ctx, "EnqueueMessage", request, response)
}
