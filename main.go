package main

import (
	"log"

	"github.com/yuriykis/foreverstore/p2p"
)

func main() {
	tcpOpt := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tr := p2p.NewTCPTransport(tcpOpt)

	go func() {
		for {
			select {
			case rpc := <-tr.Consume():
				log.Println("Received message: ", string(rpc.Payload))
			}
		}
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}
	select {}
}
