// the managed streaming quickstart, pull a body too large for one message. the
// client side is the managed RunningStream (ClientBuilder.Stream): it connects on
// its own owner goroutine, opens the stream, and Recv yields the body chunk by
// chunk until a verified end (the trailer signature checked against the peer
// NW060900). server-side streaming is driven, so a small driven server pushes /big
// across ticks on a background goroutine NWG0600. both halves run in one process.

package main

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"time"

	"nwep"
)

var body = bytes.Repeat([]byte("nwep-stream-"), 20000) // ~240 KB

var epoch = time.Now()

func nowMs() int64 { return int64(time.Since(epoch) / time.Millisecond) }

func main() {
	serverID, err := nwep.GenerateIdentity()
	if err != nil {
		log.Fatal(err)
	}
	defer serverID.Close()

	type sid struct{ conn, stream uint64 }
	opened := make(chan sid, 8)

	server, err := nwep.NewServer().
		Identity(serverID).
		Bind(nwep.Loopback(0)).
		OnRequest(func(req *nwep.Message, res *nwep.Responder) {
			if req.Path() == "/big" {
				opened <- sid{req.ConnID(), req.StreamID()}
				res.Stream("/big", "ok", [][2]string{{"content-type", "application/octet-stream"}})
				return
			}
			res.NotFound()
		}).
		Build()
	if err != nil {
		log.Fatal(err)
	}
	port := server.LocalPort()
	fmt.Printf("serving        /big (%d bytes) on [::1]:%d\n", len(body), port)

	// the driven server loop pushes the body for each opened stream to the end.
	var wg sync.WaitGroup
	stop := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		type active struct {
			conn, stream uint64
			sent         int
		}
		var streams []active
		for {
			select {
			case <-stop:
				server.Close()
				return
			default:
			}
			_ = server.Tick(nowMs())
			for {
				select {
				case s := <-opened:
					streams = append(streams, active{s.conn, s.stream, 0})
					continue
				default:
				}
				break
			}
			kept := streams[:0]
			for _, a := range streams {
				blocked := false
				for a.sent < len(body) {
					n, err := server.StreamSend(a.conn, a.stream, body[a.sent:])
					if err != nil {
						break
					}
					a.sent += n
					if n == 0 {
						blocked = true
						break
					}
				}
				if a.sent < len(body) && blocked {
					kept = append(kept, a)
					continue
				}
				_ = server.StreamEnd(a.conn, a.stream)
			}
			streams = kept
			time.Sleep(time.Millisecond)
		}
	}()
	defer func() { close(stop); wg.Wait() }()

	// the managed client stream pulls the body chunk by chunk.
	clientID, err := nwep.GenerateIdentity()
	if err != nil {
		log.Fatal(err)
	}
	defer clientID.Close()

	stream, err := nwep.NewClient().
		Identity(clientID).
		Target(serverID.NodeID()).
		Address(nwep.Loopback(port)).
		Stream(nwep.Read, "/big")
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()
	fmt.Printf("stream open    status %s\n", stream.Status())

	var got []byte
	chunks := 0
	for {
		chunk, err := stream.Recv()
		if err != nil {
			log.Fatal(err)
		}
		if chunk == nil {
			break // the verified end of the body.
		}
		got = append(got, chunk...)
		chunks++
	}
	if !bytes.Equal(got, body) {
		log.Fatalf("streamed %d bytes, want %d", len(got), len(body))
	}
	fmt.Printf("pulled         %d bytes in %d chunks, trailer verified\n", len(got), chunks)
	fmt.Println("shutdown       clean")
}
