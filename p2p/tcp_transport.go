package p2p

import (
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPPeer represents a remote node in the TCP network
type TCPPeer struct {
	// conn is the underlying TCP peer connection
	conn net.Conn

	// if we dial a connection => outbound = true
	// if we accept a connection => outbound = false
	outbound bool
}

// Close closes the underlying TCP connection, implementing the Peer interface
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcCh    chan RPC

	mu sync.RWMutex
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcCh:            make(chan RPC),
	}
}

// Consume returns a read-only channel of RPCs that can be consumed by the application, implementing the Transport interface
// This is the channel that the application will read from to receive messages from other peers
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcCh
}

// Close implements the Transport interface
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	log.Printf("TCP transport listening on %s\n", t.ListenAddr)

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Println("error accepting connection: ", err)
			return
		}
		fmt.Printf("received connection from %+v\n", conn)
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, false)

	var err error
	defer func() {
		fmt.Printf("dropping connection: %s", err)
		conn.Close()
	}()

	if err := t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// Read loop
	rpc := RPC{}
	for {
		// n, err := conn.Read(buf)
		// if err != nil {
		// 	fmt.Println("TCP error reading message: ", err)
		// 	continue
		// }
		err := t.Decoder.Decode(conn, &rpc)

		if _, ok := err.(*net.OpError); ok {
			fmt.Println("connection closed")
			return
		}

		if err != nil {
			fmt.Println("TCP error decoding message: ", err)
			continue
		}
		rpc.From = conn.RemoteAddr() // probably not ok
		t.rpcCh <- rpc
	}
}
