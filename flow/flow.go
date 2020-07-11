package flow

import "net"

type Flow struct {
	SrcIP       net.IP
	DstIP       net.IP
	SrcPort     uint16
	DstPort     uint16
	Proto       uint8
	CommunityID string
	FlowID      uint64
	Timestamp   string
}
