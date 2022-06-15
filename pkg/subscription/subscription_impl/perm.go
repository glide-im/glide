package subscription_impl

type PermMask int

const (
	MaskPermRead  PermMask = 1 << iota
	MaskPermWrite          = 1 << iota
	MaskPermAdmin          = 1 << iota
)

const (
	PermNone  Permission = 0
	PermRead  Permission = 1 << MaskPermRead
	PermWrite Permission = 1 << MaskPermWrite
	PermAdmin Permission = 1 << MaskPermAdmin
)

type Permission int64

func (perm Permission) allows(masks ...PermMask) bool {
	for _, m := range masks {
		if perm.denies(m) {
			return false
		}
	}
	return true
}

func (perm Permission) denies(mask PermMask) bool {
	b := perm >> mask
	return b&1 != 1
}
