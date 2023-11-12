package p2p

// Peer represents a remote node in the network.
type Peer interface {
}

// Transport handles the network communication between nodes.
// This can be TCP, UDP, Websockets or any other protocol.
type Transport interface {
	ListenAndAccept() error
}
