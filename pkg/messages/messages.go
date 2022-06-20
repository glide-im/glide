package messages

// ChatMessage chat message in single/group chat
type ChatMessage struct {
	Mid     int64  `json:"mid,omitempty"`
	Seq     int64  `json:"seq,omitempty"`
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	Type    int32  `json:"type,omitempty"`
	Content string `json:"content,omitempty"`
	SendAt  int64  `json:"sendAt,omitempty"`
}

// ClientCustom client custom message, server does not store to database.
type ClientCustom struct {
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	Type    int32  `json:"type,omitempty"`
	Content string `json:"content,omitempty"`
}

// AckRequest 接收者回复给服务端确认收到消息
type AckRequest struct {
	Seq  int64  `json:"seq,omitempty"`
	Mid  int64  `json:"mid,omitempty"`
	From string `json:"from,omitempty"`
}

// AckGroupMessage 发送群消息服务器回执
type AckGroupMessage struct {
	Gid int64 `json:"gid,omitempty"`
	Mid int64 `json:"mid,omitempty"`
	Seq int64 `json:"seq,omitempty"`
}

// AckMessage 服务端通知发送者的服务端收到消息
type AckMessage struct {
	Mid int64 `json:"mid,omitempty"`
	Seq int64 `json:"seq,omitempty"`
}

// AckNotify 服务端下发给发送者的消息送达通知
type AckNotify struct {
	Mid int64 `json:"mid,omitempty"`
}
