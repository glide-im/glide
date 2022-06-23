package messages

type Hello struct {
	ClientVersion string `json:"client_version,omitempty"`
	ClientName    string `json:"client_name,omitempty"`
	ClientType    string `json:"client_type,omitempty"`
}

type ServerHello struct {
	ServerVersion     string   `json:"server_version,omitempty"`
	TempID            string   `json:"temp_id,omitempty"`
	HeartbeatInterval int      `json:"heartbeat_interval,omitempty"`
	Protocols         []string `json:"protocols,omitempty"`
}
