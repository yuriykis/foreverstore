package p2p

import "net"

// RPC represents any arbitrary data that can be sent over the each transport between two nodes
type RPC struct {
	From    net.Addr
	Payload []byte
}
