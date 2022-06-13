package messages

// ChatMessage chat message in single/group chat
type ChatMessage struct {
	Mid     int64  `json:"mid"`
	Seq     int64  `json:"seq"`
	From    string `json:"from"`
	To      string `json:"to"`
	Type    int32  `json:"type"`
	Content string `json:"content"`
	SendAt  int64  `json:"sendAt"`
}

// ClientCustom client custom message, server does not store to database.
type ClientCustom struct {
	From    int64  `json:"from"`
	To      int64  `json:"to"`
	Type    int32  `json:"type"`
	Content string `json:"content"`
}

// AckRequest 接收者回复给服务端确认收到消息
type AckRequest struct {
	Seq  int64 `json:"seq"`
	Mid  int64 `json:"mid"`
	From int64 `json:"from"`
}

// AckGroupMessage 发送群消息服务器回执
type AckGroupMessage struct {
	Gid int64 `json:"gid"`
	Mid int64 `json:"mid"`
	Seq int64 `json:"seq"`
}

// AckMessage 服务端通知发送者的服务端收到消息
type AckMessage struct {
	Mid int64 `json:"mid"`
	Seq int64 `json:"seq"`
}

// AckNotify 服务端下发给发送者的消息送达通知
type AckNotify struct {
	Mid int64 `json:"mid"`
}
