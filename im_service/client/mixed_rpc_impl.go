package client

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/rpc"
	"github.com/glide-im/glide/pkg/subscription"
)

type Client struct {
	sub  *SubscriptionRpcImpl
	gate *GatewayRpcImpl
}

func NewClient(opts *rpc.ClientOptions) (*Client, error) {
	cli, err := rpc.NewBaseClient(opts)
	if err != nil {
		return nil, err
	}
	c := Client{
		sub:  NewSubscriptionRpcImplWithClient(cli),
		gate: NewGatewayRpcImplWithClient(cli),
	}
	return &c, nil
}

func (c *Client) SetClientID(old gate.ID, new_ gate.ID) error {
	return c.gate.SetClientID(old, new_)
}

func (c *Client) ExitClient(id gate.ID) error {
	return c.gate.ExitClient(id)
}

func (c *Client) EnqueueMessage(id gate.ID, message *messages.GlideMessage) error {
	return c.gate.EnqueueMessage(id, message)
}

func (c *Client) Subscribe(ch subscription.ChanID, id subscription.SubscriberID, extra interface{}) error {
	return c.sub.Subscribe(ch, id, extra)
}

func (c *Client) UnSubscribe(ch subscription.ChanID, id subscription.SubscriberID) error {
	return c.sub.UnSubscribe(ch, id)
}

func (c *Client) UpdateSubscriber(ch subscription.ChanID, id subscription.SubscriberID, extra interface{}) error {
	return c.sub.UpdateSubscriber(ch, id, extra)
}

func (c *Client) RemoveChannel(ch subscription.ChanID) error {
	return c.sub.RemoveChannel(ch)
}

func (c *Client) CreateChannel(ch subscription.ChanID, update *subscription.ChanInfo) error {
	return c.sub.CreateChannel(ch, update)
}

func (c *Client) UpdateChannel(ch subscription.ChanID, update *subscription.ChanInfo) error {
	return c.sub.UpdateChannel(ch, update)
}

func (c *Client) Publish(ch subscription.ChanID, msg subscription.Message) error {
	return c.sub.Publish(ch, msg)
}
