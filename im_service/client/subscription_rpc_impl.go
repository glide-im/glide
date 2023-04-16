package client

import (
	"context"
	"encoding/json"
	"github.com/glide-im/glide/im_service/proto"
	"github.com/glide-im/glide/pkg/rpc"
	"github.com/glide-im/glide/pkg/subscription"
)

type SubscriptionRpcImpl struct {
	rpcCli *subscriptionRpcClient
}

func NewSubscriptionRpcImplWithClient(cli *rpc.BaseClient) *SubscriptionRpcImpl {
	return &SubscriptionRpcImpl{
		rpcCli: &subscriptionRpcClient{cli: cli},
	}
}

func NewSubscriptionRpcImpl(opts *rpc.ClientOptions) (*SubscriptionRpcImpl, error) {
	c, err := rpc.NewBaseClient(opts)
	if err != nil {
		return nil, err
	}
	return NewSubscriptionRpcImplWithClient(c), nil
}

func (s *SubscriptionRpcImpl) Close() error {
	return s.rpcCli.Close()
}

func (s *SubscriptionRpcImpl) Subscribe(ch subscription.ChanID, id subscription.SubscriberID, extra interface{}) error {

	marshal, err := json.Marshal(extra)
	if err != nil {
		return err
	}
	request := &proto.SubscribeRequest{
		ChannelID:    string(ch),
		SubscriberID: string(id),
		Extra:        marshal,
	}
	reply := &proto.Response{}
	err = s.rpcCli.Subscribe(context.Background(), request, reply)
	if err != nil {
		return err
	}
	return getResponseError(reply)
}

func (s *SubscriptionRpcImpl) UnSubscribe(ch subscription.ChanID, id subscription.SubscriberID) error {
	request := &proto.UnsubscribeRequest{
		ChannelID:    string(ch),
		SubscriberID: string(id),
	}
	reply := &proto.Response{}
	err := s.rpcCli.UnSubscribe(context.Background(), request, reply)
	if err != nil {
		return err
	}
	return getResponseError(reply)
}

func (s *SubscriptionRpcImpl) UpdateSubscriber(ch subscription.ChanID, id subscription.SubscriberID, extra interface{}) error {
	marshal, err := json.Marshal(extra)
	if err != nil {
		return err
	}

	request := &proto.UpdateSubscriberRequest{
		ChannelID:    string(ch),
		SubscriberID: string(id),
		Extra:        marshal,
	}
	reply := &proto.Response{}
	err = s.rpcCli.UpdateSubscriber(context.Background(), request, reply)
	if err != nil {
		return err
	}
	return getResponseError(reply)
}

func (s *SubscriptionRpcImpl) RemoveChannel(ch subscription.ChanID) error {
	request := &proto.RemoveChannelRequest{
		ChannelID: string(ch),
	}
	reply := &proto.Response{}
	err := s.rpcCli.RemoveChannel(context.Background(), request, reply)
	if err != nil {
		return err
	}
	return getResponseError(reply)
}

func (s *SubscriptionRpcImpl) CreateChannel(ch subscription.ChanID, update *subscription.ChanInfo) error {

	var children []string
	for _, child := range update.Child {
		children = append(children, string(child))
	}
	parent := ""
	if update.Parent != nil {
		parent = string(*update.Parent)
	}
	channelInfo := &proto.ChannelInfo{
		ID:       string(update.ID),
		Type:     int32(update.Type),
		Muted:    update.Muted,
		Blocked:  update.Blocked,
		Closed:   update.Closed,
		Parent:   parent,
		Children: children,
	}

	request := &proto.CreateChannelRequest{
		ChannelID:   string(ch),
		ChannelInfo: channelInfo,
	}
	reply := &proto.Response{}
	err := s.rpcCli.CreateChannel(context.Background(), request, reply)
	if err != nil {
		return err
	}
	return getResponseError(reply)
}

func (s *SubscriptionRpcImpl) UpdateChannel(ch subscription.ChanID, update *subscription.ChanInfo) error {
	var children []string
	for _, child := range update.Child {
		children = append(children, string(child))
	}
	parent := ""
	if update.Parent != nil {
		parent = string(*update.Parent)
	}
	channelInfo := &proto.ChannelInfo{
		ID:       string(update.ID),
		Type:     int32(update.Type),
		Muted:    update.Muted,
		Blocked:  update.Blocked,
		Closed:   update.Closed,
		Parent:   parent,
		Children: children,
	}

	request := &proto.UpdateChannelRequest{
		ChannelID:   string(ch),
		ChannelInfo: channelInfo,
	}
	reply := &proto.Response{}
	err := s.rpcCli.UpdateChannel(context.Background(), request, reply)
	if err != nil {
		return err
	}
	return getResponseError(reply)
}

func (s *SubscriptionRpcImpl) Publish(ch subscription.ChanID, msg subscription.Message) error {

	marshal, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	request := &proto.PublishRequest{
		ChannelID: string(ch),
		Message:   marshal,
	}
	reply := &proto.Response{}
	err = s.rpcCli.Publish(context.Background(), request, reply)
	if err != nil {
		return err
	}
	return getResponseError(reply)
}
