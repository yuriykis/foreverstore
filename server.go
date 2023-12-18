package main

import (
	"fmt"
	"log"

	"github.com/yuriykis/foreverstore/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
}

type FileServer struct {
	FileServerOpts

	store  *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
	}
}

func (fs *FileServer) Stop() {
	fmt.Println("Stopping FileServer")
	close(fs.quitch)
}

func (fs *FileServer) loop() {

	defer func() {
		log.Println("FileServer stopped due to user request")
	}()

	for {
		select {
		case <-fs.quitch:
			return
		case msg := <-fs.Transport.Consume():
			fmt.Println(msg)
		}
	}
}

func (fs *FileServer) Start() error {
	if err := fs.Transport.ListenAndAccept(); err != nil {
		return err
	}

	fs.loop()

	return nil
}
