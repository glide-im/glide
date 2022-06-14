package subscription

// Message is a message that can be sent to a subscriber.
type Message interface {
	GetFrom() SubscriberID
}
