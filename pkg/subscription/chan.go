package subscription

//goland:noinspection GoUnusedGlobalVariable
var (
	ErrChanNotExist      = "channel does not exist"
	ErrChanAlreadyExists = "channel already exists"
	ErrChanClosed        = "subscribe channel is closed"
	ErrAlreadySubscribed = "already subscribed"
	ErrNotSubscribed     = "not subscribed"
	ErrNotPublisher      = "not publisher"
)

type Subscriber struct {
	ID   string
	Type string
}

func (s *Subscriber) Notify(msg Message) error {
	return nil
}

type Channel interface {
	Subscribe(id SubscriberID, extra interface{}) error

	Unsubscribe(id SubscriberID) error

	Update(i *ChanInfo) error

	Publish(msg Message) error

	Close() error
}
