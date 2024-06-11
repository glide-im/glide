package messages

// ChatMessage chat message in single/group chat
type ChatMessage struct {
	/// client message id to identity unique a message.
	/// for identity a message
	/// and wait for the server ack receipt and return `mid` for it.
	CliMid string `json:"cliMid,omitempty"`
	/// server message id in the database.
	// when a client sends a message for the first time or  client retry to send a message that
	// the server does not ack, the 'Mid' is empty.
	/// if this field is not empty that this message is server acked, need not store to database again.
	Mid int64 `json:"mid,omitempty"`
	/// message sequence for a chat, use to check message whether the message lost.
	Seq int64 `json:"seq,omitempty"`
	/// message sender
	From string `json:"from,omitempty"`
	/// message send to
	To string `json:"to,omitempty"`
	/// message type
	Type int32 `json:"type,omitempty"`
	/// message content
	Content string `json:"content,omitempty"`
	/// message send time, server store message time.
	SendAt int64 `json:"sendAt,omitempty"`
}

// ClientCustom client custom message, server does not store to database.
type ClientCustom struct {
	Type    string      `json:"type,omitempty"`
	Content interface{} `json:"content,omitempty"`
}

// AckRequest 接收者回复给服务端确认收到消息
type AckRequest struct {
	CliMid string `json:"cli_mid,omitempty"`
	Seq    int64  `json:"seq,omitempty"`
	Mid    int64  `json:"mid,omitempty"`
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
}

// AckGroupMessage 发送群消息服务器回执
type AckGroupMessage struct {
	CliMid string `json:"cli_mid,omitempty"`
	Gid    int64  `json:"gid,omitempty"`
	Mid    int64  `json:"mid,omitempty"`
	Seq    int64  `json:"seq,omitempty"`
}

// AckMessage 服务端通知发送者的服务端收到消息
type AckMessage struct {
	CliMid string `json:"cli_mid,omitempty"`
	/// message id to tall the client
	Mid  int64  `json:"mid,omitempty"`
	From string `json:"from,omitempty"`
	Seq  int64  `json:"seq,omitempty"`
}

// AckNotify 服务端下发给发送者的消息送达通知
type AckNotify struct {
	CliMid string `json:"cli_mid,omitempty"`
	Seq    int64  `json:"seq,omitempty"`
	Mid    int64  `json:"mid,omitempty"`
	From   string `json:"from,omitempty"`
}

type KickOutNotify struct {
	DeviceId   string `json:"device_id,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
}
