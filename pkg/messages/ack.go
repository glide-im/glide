package messages

// AckRequest 接收者回复给服务端确认收到消息
type AckRequest struct {
	Seq  int64
	Mid  int64
	From int64
}

type AckGroupMessage struct {
	Gid int64
	Mid int64
	Seq int64
}

// AckMessage 服务端通知发送者的服务端收到消息
type AckMessage struct {
	Mid int64
	Seq int64
}

func NewAckMessage(mid int64, seq int64) AckMessage {
	return AckMessage{Mid: mid, Seq: seq}
}

// AckNotify 服务端下发给发送者的消息送达通知
type AckNotify struct {
	Mid int64
}

func NewAckNotify(mid int64) AckNotify {
	return AckNotify{Mid: mid}
}

type GroupNotify struct {
	Mid       int64
	Gid       int64
	Type      int64
	Seq       int64
	Timestamp int64
	Data      interface{}
}
