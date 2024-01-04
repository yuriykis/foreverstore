package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/yuriykis/foreverstore/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.RWMutex
	peers    map[string]p2p.Peer
	store    *Store
	quitch   chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		peers:          make(map[string]p2p.Peer),
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
	}
}

// broadcast sends a message to all peers
func (fs *FileServer) broadcast(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range fs.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)

	return gob.NewEncoder(mw).Encode(msg)
}

type Message struct {
	Payload any
}

// StoreData stores data in the store and broadcasts it to all peers
func (fs *FileServer) StoreData(key string, r io.Reader) error {

	buf := new(bytes.Buffer)
	msg := &Message{
		Payload: []byte("storegekey"),
	}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range fs.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	payload := []byte("THIS LARGE FILE IS STORED IN THE FOREVERSTORE")
	for _, peer := range fs.peers {
		if err := peer.Send(payload); err != nil {
			return err
		}
	}

	return nil
	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r, buf)

	// if err := fs.store.Write(key, tee); err != nil {
	// 	return err
	// }

	// p := &DataMessage{
	// 	Key:  key,
	// 	Data: buf.Bytes(),
	// }

	// return fs.broadcast(&Message{
	// 	From:    "todo",
	// 	Payload: p,
	// })
}

func (fs *FileServer) Stop() {
	fmt.Println("Stopping FileServer")
	close(fs.quitch)
}

func (fs *FileServer) OnPeer(p p2p.Peer) error {
	fs.peerLock.Lock()
	defer fs.peerLock.Unlock()

	fs.peers[p.RemoteAddr().String()] = p

	log.Printf("New peer connected: %s\n", p.RemoteAddr())

	return nil
}

func (fs *FileServer) loop() {

	defer func() {
		log.Println("FileServer stopped due to user request")
		fs.Transport.Close()
	}()

	for {
		select {
		case rpc := <-fs.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Printf("error decoding message: %s\n", err)
			}

			peer, ok := fs.peers[rpc.From]
			if !ok {
				log.Printf("received message from unknown peer %s\n", rpc.From)
				continue
			}

			log.Printf("Peer: %d", peer)

			log.Printf("Received message: %s\n", string(msg.Payload.([]byte)))
			// if err := fs.handleMessage(msg); err != nil {
			// 	log.Printf("error handling message: %s\n", err)
			// }
		case <-fs.quitch:
			return
		}
	}
}

// func (fs *FileServer) handleMessage(msg Message) error {
// 	switch v := msg.Payload.(type) {
// 	case *DataMessage:
// 		log.Printf("Received data message: %s\n", v.Key)
// 		// fs.store.Write(v.Key, bytes.NewReader(v.Data))
// 	}
// 	return nil
// }

// bootstrapNetwork tries to connect to all bootstrap nodes
func (fs *FileServer) bootstrapNetwork() error {
	for _, addr := range fs.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			if err := fs.Transport.Dial(addr); err != nil {
				log.Printf("error dialing bootstrap node %s: %s\n", addr, err)
			}
		}(addr)
	}
	return nil
}

func (fs *FileServer) Start() error {
	if err := fs.Transport.ListenAndAccept(); err != nil {
		return err
	}

	fs.bootstrapNetwork()
	fs.loop()

	return nil
}
