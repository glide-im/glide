package im_server

import (
	"context"
	"encoding/json"
	"github.com/glide-im/glide/im_service/proto"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/rpc"
	"github.com/glide-im/glide/pkg/subscription"
	"github.com/glide-im/glide/pkg/subscription/subscription_impl"
)

type RpcServer struct {
	gateway gate.Gateway
	sub     subscription_impl.SubscribeWrap
}

func RunRpcServer(options *rpc.ServerOptions, gate gate.Gateway, subscribe subscription.Subscribe) error {
	server := rpc.NewBaseServer(options)
	rpcServer := RpcServer{
		gateway: gate,
		sub:     subscription_impl.NewSubscribeWrap(subscribe),
	}
	server.Register(options.Name, &rpcServer)
	return server.Run()
}

func (r *RpcServer) SetClientID(ctx context.Context, request *proto.SetIDRequest, response *proto.Response) error {
	err := r.gateway.SetClientID(gate.ID(request.OldId), gate.ID(request.NewId))
	if err != nil {
		response.Code = int32(proto.Response_ERROR)
		response.Msg = err.Error()
	}
	return nil
}

func (r *RpcServer) ExitClient(ctx context.Context, request *proto.ExitClientRequest, response *proto.Response) error {
	err := r.gateway.ExitClient(gate.ID(request.Id))
	if err != nil {
		response.Code = int32(proto.Response_ERROR)
		response.Msg = err.Error()
	}
	return nil
}

func (r *RpcServer) IsOnline(ctx context.Context, request *proto.IsOnlineRequest, response *proto.IsOnlineResponse) error {
	response.Online = r.gateway.IsOnline(gate.ID(request.Id))
	return nil
}

func (r *RpcServer) EnqueueMessage(ctx context.Context, request *proto.EnqueueMessageRequest, response *proto.Response) error {

	msg := messages.GlideMessage{}
	err := json.Unmarshal(request.Msg, &msg)
	if err != nil {
		response.Code = int32(proto.Response_ERROR)
		response.Msg = err.Error()
		return nil
	}

	err = r.gateway.EnqueueMessage(gate.ID(request.Id), &msg)
	if err != nil {
		response.Code = int32(proto.Response_ERROR)
		response.Msg = err.Error()
	}
	return nil
}

////////////////////////////////////// Subscription //////////////////////////////////////////////

func (r *RpcServer) Subscribe(ctx context.Context, request *proto.SubscribeRequest, response *proto.Response) error {

	subscriberID := subscription.SubscriberID(request.SubscriberID)
	channelID := subscription.ChanID(request.ChannelID)

	info := subscription_impl.SubscriberOptions{}
	err := json.Unmarshal(request.GetExtra(), &info)
	if err != nil {
		response.Code = int32(proto.Response_ERROR)
		response.Msg = err.Error()
		return nil
	}

	err = r.sub.Subscribe(channelID, subscriberID, &info)
	if err != nil {
		response.Code = int32(proto.Response_ERROR)
		response.Msg = err.Error()
	}
	return nil
}

func (r *RpcServer) UnSubscribe(ctx context.Context, request *proto.UnsubscribeRequest, response *proto.Response) error {
	chanId := subscription.ChanID(request.ChannelID)
	subscriberId := subscription.SubscriberID(request.SubscriberID)
	err := r.sub.UnSubscribe(chanId, subscriberId)
	if err != nil {
		response.Code = int32(proto.Response_ERROR)
		response.Msg = err.Error()
	}
	return nil
}

func (r *RpcServer) UpdateSubscriber(ctx context.Context, request *proto.UpdateSubscriberRequest, response *proto.Response) error {

	chanId := subscription.ChanID(request.ChannelID)
	subscriberID := subscription.SubscriberID(request.SubscriberID)
	info := subscription_impl.SubscriberOptions{}
	err := json.Unmarshal(request.GetExtra(), &info)
	if err != nil {
		response.Code = int32(proto.Response_ERROR)
		response.Msg = err.Error()
		return nil
	}
	err = r.sub.UpdateSubscriber(chanId, subscriberID, &info)
	if err != nil {
		response.Code = int32(proto.Response_ERROR)
		response.Msg = err.Error()
	}
	return nil
}

func (r *RpcServer) RemoveChannel(ctx context.Context, request *proto.RemoveChannelRequest, response *proto.Response) error {
	chanId := subscription.ChanID(request.ChannelID)
	err := r.sub.RemoveChannel(chanId)
	if err != nil {
		response.Code = int32(proto.Response_ERROR)
		response.Msg = err.Error()
	}
	return nil
}

func (r *RpcServer) CreateChannel(ctx context.Context, request *proto.CreateChannelRequest, response *proto.Response) error {
	chanId := subscription.ChanID(request.ChannelID)
	cInfo := request.GetChannelInfo()

	info := subscription.ChanInfo{
		ID:      chanId,
		Type:    subscription.ChanType(cInfo.Type),
		Muted:   cInfo.Muted,
		Blocked: cInfo.Blocked,
		Closed:  cInfo.Closed,
	}
	err := r.sub.CreateChannel(chanId, &info)
	if err != nil {
		response.Code = int32(proto.Response_ERROR)
		response.Msg = err.Error()
	}
	return nil
}

func (r *RpcServer) UpdateChannel(ctx context.Context, request *proto.UpdateChannelRequest, response *proto.Response) error {
	chanId := subscription.ChanID(request.ChannelID)
	cInfo := request.GetChannelInfo()
	info := subscription.ChanInfo{
		ID:      chanId,
		Type:    subscription.ChanType(cInfo.Type),
		Muted:   cInfo.Muted,
		Blocked: cInfo.Blocked,
		Closed:  cInfo.Closed,
	}
	err := r.sub.UpdateChannel(chanId, &info)
	if err != nil {
		response.Code = int32(proto.Response_ERROR)
		response.Msg = err.Error()
	}
	return nil
}

func (r *RpcServer) Publish(ctx context.Context, request *proto.PublishRequest, response *proto.Response) error {
	chanId := subscription.ChanID(request.ChannelID)
	msg := subscription_impl.PublishMessage{}
	err := json.Unmarshal(request.GetMessage(), &msg)
	if err != nil {
		response.Code = int32(proto.Response_ERROR)
		response.Msg = err.Error()
		return nil
	}
	err = r.sub.Publish(chanId, &msg)
	if err != nil {
		response.Code = int32(proto.Response_ERROR)
		response.Msg = err.Error()
	}
	return nil
}
