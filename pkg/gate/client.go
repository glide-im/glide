package gate

import (
	"github.com/glide-im/glide/pkg/messages"
	"strings"
)

var (
	ClientTypeRobot = 1
	ClientTypeUser  = 2
)

// tempIdPrefix is the prefix for temporary IDs in the second part of the ID.
const tempIdPrefix = "tmp@"

// idSeparator is the separator used to separate the part of the ID.
const idSeparator = "_"

// ID is used to identify the client, the ID is consist of multiple parts, some of them are optional.
// The ID is constructed by concatenating the parts with a '_' separator, and the parts are:
//   - The gateway id (optional): the string id of the gateway that the client is connected to.
//   - The client id (required): the string id  of the client, it is unique for user.
//   - if the client is temporary, this id is a string generated by the gateway and start with `tmp`.
//   - The client type (optional): the int type of the client, like 'web', 'mobile', 'desktop', etc.
type ID string

// NewID2 creates a new ID from the given user id, use the empty gateway id and the empty client type.
func NewID2(uid string) ID {
	return ID(strings.Join([]string{"", uid, ""}, idSeparator))
}

// NewID creates a new ID from the given user id, gateway id and client type.
func NewID(gate string, uid string, device string) ID {
	return ID(strings.Join([]string{gate, uid, device}, idSeparator))
}

// Device returns the device type of the client, if the client device type is not set, it returns "".
func (i *ID) Device() string {
	return i.getPart(2)
}

// UID returns the user id of the client, if the client is temporary, it returns "".
func (i *ID) UID() string {
	return i.getPart(1)
}

// Gateway returns the gateway id of the client, if not set, it returns an empty string.
func (i *ID) Gateway() string {
	return i.getPart(0)
}

// SetGateway sets the gateway part of the ID.
func (i *ID) SetGateway(gateway string) bool {
	if strings.HasPrefix(string(*i), gateway) {
		return false
	}
	s := strings.Split(string(*i), idSeparator)
	if len(s) != 3 {
		return false
	}
	s[0] = gateway
	*i = ID(strings.Join(s, idSeparator))
	return true
}

// SetDevice sets the device type of the client.
func (i *ID) SetDevice(device string) bool {
	if strings.HasSuffix(string(*i), device) {
		return false
	}
	s := strings.Split(string(*i), idSeparator)
	if len(s) != 3 {
		return false
	}
	s[2] = device
	*i = ID(strings.Join(s, idSeparator))
	return true
}

// IsTemp returns true if the ID is a temporary.
func (i *ID) IsTemp() bool {
	return strings.HasPrefix(i.getPart(1), tempIdPrefix)
}

func (i *ID) Equals(other ID) bool {
	return i.UID()+i.Device() == other.UID()+other.Device()
}

func (i *ID) getPart(index int) string {
	s := strings.Split(string(*i), idSeparator)
	if index >= len(s) {
		return ""
	}
	return s[index]
}

// Info represents a client's information.
type Info struct {

	// ID is the unique identifier for the client.
	ID ID

	// ConnectionId generated by client, used to identify the client connection.
	ConnectionId string

	// Version is the version of the client.
	Version string

	// AliveAt is the time the client was last seen.
	AliveAt int64

	// ConnectionAt is the time the client was connected.
	ConnectionAt int64

	// Gateway is the name of the gateway the client is connected to.
	Gateway string

	// CliAddr is the address of the client.
	CliAddr string
}

// Client is a client connection abstraction.
type Client interface {

	// SetID sets the ID of the client.
	SetID(id ID)

	// IsRunning returns true if the client is running/alive.
	IsRunning() bool

	// EnqueueMessage enqueues a message to be sent to the client.
	EnqueueMessage(message *messages.GlideMessage) error

	// Exit the client and close the connection.
	Exit()

	// Run starts the client message handling loop and blocks until the client.
	Run()

	// GetInfo returns the client's information.
	GetInfo() Info
}

// ClientSecrets used to control client permission.
type ClientSecrets struct {
	// MessageDeliverSecret is the secret of the client, used to authenticate the client message.
	// The secret is generated by the business service, saved in business service, client should not know it.
	// When client send a message to someone else, it should get the sign of the message target, and send it
	// with the message. If business service want to control which one the client can send message to,
	// business service can generate different secret for client, and notify the gateway update the secret, to make
	// client old sign invalid.
	MessageDeliverSecret string `json:"message_deliver_secret"`
	OnlineStateSecret    string `json:"online_state_secret"`
}

// EncryptedCredential represents the encrypted credential.
type EncryptedCredential struct {
	// Version is the version of the credential.
	Version int `json:"version"`

	// Credential is the encrypted credential string.
	Credential string `json:"credential"`
}

// ConnectionConfig _
type ConnectionConfig struct {
	AllowMaxHeartbeatLost int
	HeartbeatDuration     int
	CloseImmediately      bool
}

type RiskControl struct {
	MaxMessagesPeerSecond int
}

// ClientAuthCredentials represents the client authentication credentials.
// Used to client authentication when connecting to the gateway, credentials are generated by business service,
// encrypted use the gateway's secret key, and sent to the client.
type ClientAuthCredentials struct {

	// Type is the type of the client.
	Type int `json:"type"`

	// UserID uniquely identifies the client.
	UserID string `json:"user_id"`

	// DeviceID is the id of the client device, it is unique for same client.
	DeviceID string `json:"device_id"`

	DeviceName string `json:"device_name"`

	Secrets *ClientSecrets `json:"secrets"`

	RiskControl *RiskControl `json:"risk_control"`

	ConnectionConfig *ConnectionConfig `json:"connection_config"`

	// ConnectionID is the temporary connection id of the client, generated by the client.
	ConnectionID string `json:"connection_id"`

	// Timestamp of credentials creation.
	Timestamp int64 `json:"timestamp"`
}

func (a *ClientAuthCredentials) validate() error {

	return nil
}
