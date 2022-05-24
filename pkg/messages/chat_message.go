package messages

// ChatMessage 上行消息, 表示服务端收到发送者的消息
type ChatMessage struct {
	// Mid 消息ID
	Mid int64
	// Seq 发送者消息 seq
	Seq int64
	// From internal
	From int64
	// To 接收者 ID
	To int64
	// Type 消息类型
	Type int32
	// Content 消息内容
	Content string
	// SendAt 发送时间
	SendAt int64
}

type ClientCustom struct {
	From    int64
	To      int64
	Type    int32
	Content string
}
