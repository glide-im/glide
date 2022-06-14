package group_subscription

const (
	maskPermRead  = 1 << iota
	maskPermWrite = 1 << iota
	maskPermAdmin = 1 << iota
)

const (
	PermNone  Permission = 0
	PermRead             = PermNone | 1<<maskPermRead
	PermWrite            = PermNone | 1<<maskPermWrite
	PermAdmin            = PermNone | 1<<maskPermAdmin
)

type Permission int64

func (perm Permission) Allows(permissions ...Permission) bool {
	for _, p := range permissions {
		if p&perm != p {
			return false
		}
	}
	return true
}

func (perm Permission) Denies(permissions ...Permission) bool {
	for _, p := range permissions {
		if p&perm == p {
			return false
		}
	}
	return true
}
