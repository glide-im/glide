package subscribe

import "errors"

//goland:noinspection GoUnusedGlobalVariable
var (
	ErrChanClosed        = errors.New("subscribe channel is closed")
	ErrAlreadySubscribed = errors.New("already subscribed")
	ErrNotSubscribed     = errors.New("not subscribed")
	ErrNotPublisher      = errors.New("not publisher")
)

type Subscriber struct {
	ID   string
	Type string
}

func (s *Subscriber) Notify(msg Message) error {
	return nil
}

type Channel interface {
	GetSubscriber(id string) (Subscriber, error)

	Subscribe(id string, subscriber Subscriber) error

	Unsubscribe(id string) error

	UnsubscribeAll() error

	Publish(msg Message) error

	Close() error
}
