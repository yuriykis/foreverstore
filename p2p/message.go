package p2p

// Message represents any arbitrary data that can be sent over the each transport between two nodes
type Message struct {
	Payload []byte
}
