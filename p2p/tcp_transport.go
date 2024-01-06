package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPPeer represents a remote node in the TCP network
type TCPPeer struct {
	// the underlying TCP connection of the peer
	net.Conn

	// if we dial a connection => outbound = true
	// if we accept a connection => outbound = false
	outbound bool

	Wg *sync.WaitGroup
}

// Send sends a message to the remote peer, implementing the Peer interface
func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		Wg:       &sync.WaitGroup{},
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

// Dial implements the Transport interface
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConn(conn, true)
	return nil
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
		if errors.Is(err, net.ErrClosed) {
			log.Println("TCP transport stopped")
			return
		}
		if err != nil {
			log.Printf("error accepting connection: %s\n", err)
			return
		}
		log.Printf("accepted connection from %s\n", conn.RemoteAddr())
		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
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
		rpc.From = conn.RemoteAddr().String()
		peer.Wg.Add(1)
		log.Printf("waiting till stream is read")
		t.rpcCh <- rpc
		peer.Wg.Wait()
		log.Printf("stream is read")
	}
}
