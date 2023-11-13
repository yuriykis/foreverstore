package p2p

import (
	"fmt"
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
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcCh    chan RPC

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		peers:            make(map[net.Addr]Peer),
		rpcCh:            make(chan RPC),
	}
}

// Consume returns a read-only channel of RPCs that can be consumed by the application, implementing the Transport interface
// This is the channel that the application will read from to receive messages from other peers
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcCh
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()
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

	if err := t.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Println("TCP error shaking hands: ", err)
		return
	}

	// Read loop
	rpc := RPC{}
	for {
		// n, err := conn.Read(buf)
		// if err != nil {
		// 	fmt.Println("TCP error reading message: ", err)
		// 	continue
		// }
		if err := t.Decoder.Decode(conn, &rpc); err != nil {
			fmt.Println("TCP error decoding message: ", err)
			continue
		}
		rpc.From = conn.RemoteAddr() // probably not ok
		t.rpcCh <- rpc
	}
}
