package gate

const (
	errClientClosed       = "client closed"
	errClientNotExist     = "client does not exist"
	errClientAlreadyExist = "id already exist"
)

func IsClientClosed(err error) bool {
	return err != nil && err.Error() == errClientClosed
}

func IsClientNotExist(err error) bool {
	return err != nil && err.Error() == errClientNotExist
}

// IsIDAlreadyExist returns true if the error is caused by the ID of the client already exist.
// Returns when SetClientID is called with the existing new ID.
func IsIDAlreadyExist(err error) bool {
	return err != nil && err.Error() == errClientAlreadyExist
}
