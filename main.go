package main

import (
	"log"

	"github.com/yuriykis/foreverstore/p2p"
)

func main() {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		// TODO: OnPeer
	}
	tcpTrasport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot:       "3000_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTrasport,
	}
	s := NewFileServer(fileServerOpts)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}

	select {}
}
