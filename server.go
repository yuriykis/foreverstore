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

func (fs *FileServer) broadcast(p *Payload) error {
	peers := []io.Writer{}
	for _, peer := range fs.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)

	return gob.NewEncoder(mw).Encode(p)
}

type Payload struct {
	Key  string
	Data []byte
}

// StoreData stores data in the store and broadcasts it to all peers
func (fs *FileServer) StoreData(key string, r io.Reader) error {
	if err := fs.store.Write(key, r); err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r); err != nil {
		return err
	}

	p := &Payload{
		Key:  key,
		Data: buf.Bytes(),
	}

	return fs.broadcast(p)
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
		case msg := <-fs.Transport.Consume():
			var p Payload
			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&p); err != nil {
				log.Printf("error decoding message: %s\n", err)
			}
			log.Printf("Received message: %s\n", p)
		case <-fs.quitch:
			return
		}
	}
}

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
