package p2p

import "net"

// Message represents any arbitrary data that can be sent over the each transport between two nodes
type Message struct {
	From    net.Addr
	Payload []byte
}
