package p2p

// HandshakeFunc is a function that is called when a new connection is established.
type HandshakeFunc func(Peer) error

// NOPHandshakeFunc is a no-op handshake function.
func NOPHandshakeFunc(Peer) error { return nil }
