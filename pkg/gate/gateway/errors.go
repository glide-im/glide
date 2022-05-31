package gateway

const (
	errClientClosed       = "client closed"
	errClientNotExist     = "client does not exist"
	errClientAlreadyExist = "client already exist"
)

func IsClientClosed(err error) bool {
	return err.Error() == errClientClosed
}

func IsClientNotExist(err error) bool {
	return err.Error() == errClientNotExist
}

func IsAlreadyExist(err error) bool {
	return err.Error() == errClientAlreadyExist
}
