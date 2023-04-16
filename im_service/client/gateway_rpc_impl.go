package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/glide-im/glide/im_service/proto"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/rpc"
	"strings"
)

const (
	errRpcInvocation = "gate invocation error: "
)

type IMServiceError struct {
	Code    int32
	Message string
}

func (e *IMServiceError) Error() string {
	return fmt.Sprintf("IM Service Error: %d, %s", e.Code, e.Message)
}

// IsRpcInvocationError
// Rpc invocation failed errors are returned by the gate client when the rpc call fails.
func IsRpcInvocationError(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), errRpcInvocation)
}

type GatewayRpcImpl struct {
	gate *GatewayRpcClient
}

func (i *GatewayRpcImpl) SetClientID(old gate.ID, new_ gate.ID) error {
	//TODO implement me update client
	panic("implement me")
}

func (i *GatewayRpcImpl) ExitClient(id gate.ID) error {
	//TODO implement me update client
	panic("implement me")
}

func NewGatewayRpcImplWithClient(client *rpc.BaseClient) *GatewayRpcImpl {
	return &GatewayRpcImpl{
		gate: &GatewayRpcClient{
			cli: client,
		},
	}
}

func NewGatewayRpcImpl(opts *rpc.ClientOptions) (*GatewayRpcImpl, error) {
	cli, err := rpc.NewBaseClient(opts)
	if err != nil {
		return nil, err
	}
	return NewGatewayRpcImplWithClient(cli), nil
}

func (i *GatewayRpcImpl) EnqueueMessage(id gate.ID, message *messages.GlideMessage) error {

	marshal, err := json.Marshal(message)
	if err != nil {
		return err
	}
	ctx := context.TODO()
	request := proto.EnqueueMessageRequest{
		Id:  string(id),
		Msg: marshal,
	}
	response := proto.Response{}
	err = i.gate.EnqueueMessage(ctx, &request, &response)
	if err != nil {
		return errors.New(errRpcInvocation + err.Error())
	}
	return getResponseError(&response)
}

func (i *GatewayRpcImpl) Close() error {
	return i.gate.cli.Close()
}

func getResponseError(response *proto.Response) error {
	if proto.Response_ResponseCode(response.GetCode()) != proto.Response_OK {
		return &IMServiceError{
			Code:    response.GetCode(),
			Message: response.GetMsg(),
		}
	}
	return nil
}
