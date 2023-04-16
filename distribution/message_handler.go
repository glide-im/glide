package distribution

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/messaging"
	"github.com/glide-im/glide/pkg/subscription"
)

type DistributedMessaging struct {
	def messaging.Messaging
}

func New(def messaging.Messaging) {

}

func (m *DistributedMessaging) Handle(clientInfo *gate.Info, msg *messages.GlideMessage) error {
	return m.def.Handle(clientInfo, msg)
}

func (m *DistributedMessaging) AddHandler(i messaging.MessageHandler) {
	m.def.AddHandler(i)
}

func (m *DistributedMessaging) SetSubscription(g subscription.Interface) {
	m.def.SetSubscription(g)
}

func (m *DistributedMessaging) SetGate(g gate.Gateway) {
	m.def.SetGate(g)
}

type gateway struct {
	self gate.Gateway
}

func (g gateway) SetClientID(old gate.ID, new_ gate.ID) error {
	//TODO implement me
	panic("implement me")
}

func (g gateway) ExitClient(id gate.ID) error {
	//TODO implement me
	panic("implement me")
}

func (g gateway) EnqueueMessage(id gate.ID, message *messages.GlideMessage) error {
	//TODO implement me
	panic("implement me")
}
