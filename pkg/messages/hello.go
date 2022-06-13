package messages

type Hello struct {
	ClientVersion string `json:"client_version"`

	ClientName string `json:"client_name"`
	ClientType string `json:"client_type"`
}
