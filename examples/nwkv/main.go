// nwkv is a tiny in-memory key-value server over web/1, mirroring the sandbox app.
//
// it serves READ /key and WRITE /key against an in-memory map, showing the managed
// server on-ramp, a handler that branches on method, and the deferred-free common
// path. run it, then point any web/1 client (or the nwcurl example) at the node_id
// and port it prints. see bindings/guide.md for the layering.

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"nwep"
)

// store is a goroutine-safe map standing in for real storage.
type store struct {
	mu sync.RWMutex
	m  map[string][]byte
}

func (s *store) get(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[key]
	return v, ok
}

func (s *store) put(key string, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
}

func main() {
	id, err := nwep.GenerateIdentity()
	if err != nil {
		log.Fatal(err)
	}
	defer id.Close()

	db := &store{m: map[string][]byte{}}

	server, err := nwep.NewServer().
		Identity(id).
		Bind(nwep.Wildcard(0)).
		OnRequest(func(req *nwep.Message, res *nwep.Responder) {
			key := req.Path()
			switch req.Method() {
			case "READ":
				if v, ok := db.get(key); ok {
					res.OK(v)
				} else {
					res.NotFound()
				}
			case "WRITE":
				db.put(key, req.Body())
				res.OK(nil)
			default:
				res.NotAllowed()
			}
		}).
		Serve()
	if err != nil {
		log.Fatal(err)
	}
	defer server.Shutdown()

	fmt.Printf("nwkv listening\n  node_id %s\n  port    %d\n", server.NodeID(), server.LocalPort())

	// run until interrupted, the managed runtime owns the loop in the meantime.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	fmt.Println("\nshutting down")
}
