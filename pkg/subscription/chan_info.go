package subscription

type ChanType int32

//goland:noinspection GoUnusedConst
const (
	ChanTypeUnknown ChanType = 0
)

type ChanInfo struct {
	ID   ChanID
	Type ChanType

	Parent *ChanID
	Child  []ChanID
}

func NewChanInfo(id ChanID, type_ ChanType) *ChanInfo {
	return &ChanInfo{
		ID:   id,
		Type: type_,
	}
}

type Chan struct {
}
