package engine

const (
	maxMsgSize = 100
)

// HostID unique host id generated by user
type HostID string
// MsgType basic message type
type MsgType int

const (
	_ MsgType = iota
// MsgOpen open connection msg
	MsgOpen
// MsgSend msg to server
	MsgSend
// MsgReceived msg from server
	MsgReceived
// MsgClose close msg
	MsgClose
// MsgError error msg
	MsgError
)

func (m MsgType) String() string {
	switch m {
	case MsgOpen:
		return "Open"
	case MsgSend:
		return "Send"
	case MsgReceived:
		return "Received"
	case MsgClose:
		return "Close"
	case MsgError:
		return "Error"
	}
	return "Invalid MsgType"
}

// Message between sever and client
type Message struct {
	Type MsgType
	Host Host
	Msg  string
}

