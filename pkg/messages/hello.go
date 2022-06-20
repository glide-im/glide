package messages

type Hello struct {
	ClientVersion string `json:"client_version,omitempty"`

	ClientName string `json:"client_name,omitempty"`
	ClientType string `json:"client_type,omitempty"`
}
