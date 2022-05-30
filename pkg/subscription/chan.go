package subscription

import "errors"

//goland:noinspection GoUnusedGlobalVariable
var (
	ErrChanNotExist      = errors.New("channel does not exist")
	ErrChanAlreadyExists = errors.New("channel already exists")
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
	GetSubscriber(id SubscriberID) (Subscriber, error)

	Subscribe(id SubscriberID, extra interface{}) error

	Unsubscribe(id SubscriberID) error

	UpdateSubscribe(id SubscriberID, extra interface{}) error

	Update(extra interface{}) error

	UnsubscribeAll() error

	Publish(msg Message) error

	Close() error
}
