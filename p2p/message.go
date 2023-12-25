package p2p

import "net"

// RPC represents any arbitrary data that can be sent over the each transport between two nodes
// can be a message, a file, a request, a response, etc.
type RPC struct {
	From    net.Addr
	Payload []byte
}
