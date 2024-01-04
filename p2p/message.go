package p2p

// RPC represents any arbitrary data that can be sent over the each transport between two nodes
// can be a message, a file, a request, a response, etc.
type RPC struct {
	From    string
	Payload []byte
}
