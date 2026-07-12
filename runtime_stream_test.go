package nwep

import (
	"bytes"
	"testing"
	"time"
)

type streamSid struct{ conn, stream uint64 }
type streamActive struct {
	conn, stream uint64
	sent         int
}

// TestManagedStream proves the managed RunningStream pulls a streamed body and
// verifies its trailer at the end NW060200 NW060900.
//
// a driven server streams a large /big body across ticks (server-side streaming is
// driven NWG0600), and the managed client stream reads it chunk by chunk.
func TestManagedStream(t *testing.T) {
	// the body is larger than one message chunk, so it must stream.
	body := bytes.Repeat([]byte("nwep-stream-"), 20000) // ~240 KB

	serverID, _ := GenerateIdentity()
	defer serverID.Close()

	opened := make(chan streamSid, 8)

	server, err := NewServer().
		Identity(serverID).
		Bind(Loopback(0)).
		OnRequest(func(req *Message, res *Responder) {
			if req.Path() == "/big" {
				opened <- streamSid{req.ConnID(), req.StreamID()}
				res.Stream("/big", "ok", [][2]string{{"content-type", "application/octet-stream"}})
				return
			}
			res.NotFound()
		}).
		Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	port := server.LocalPort()

	// the server owner loop: tick, and push body for each opened stream to the end.
	stop := make(chan struct{})
	stopped := make(chan struct{})
	go runStreamServerLoop(server, body, opened, stop, stopped)
	defer func() { close(stop); <-stopped }()

	clientID, _ := GenerateIdentity()
	defer clientID.Close()

	stream, err := NewClient().
		Identity(clientID).
		Target(serverID.NodeID()).
		Address(Loopback(port)).
		Stream(Read, "/big")
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	defer stream.Close()

	if stream.Status() != "ok" {
		t.Fatalf("stream status = %q, want ok", stream.Status())
	}
	got := collectStreamChunks(t, stream)
	if !bytes.Equal(got, body) {
		t.Fatalf("streamed %d bytes, want %d", len(got), len(body))
	}
}

func collectStreamChunks(t *testing.T, stream *RunningStream) []byte {
	t.Helper()
	var got []byte
	for {
		chunk, err := stream.Recv()
		if err != nil {
			t.Fatalf("recv: %v", err)
		}
		if chunk == nil {
			return got
		}
		got = append(got, chunk...)
	}
}

func runStreamServerLoop(server *Server, body []byte, opened <-chan streamSid, stop <-chan struct{}, stopped chan<- struct{}) {
	defer close(stopped)
	var streams []streamActive
	for {
		select {
		case <-stop:
			server.Close()
			return
		default:
		}
		_ = server.Tick(monoMs())
		streams = drainStreamOpened(opened, streams)
		streams = pumpStreamSends(server, body, streams)
		time.Sleep(time.Millisecond)
	}
}

func drainStreamOpened(opened <-chan streamSid, streams []streamActive) []streamActive {
	for {
		select {
		case s := <-opened:
			streams = append(streams, streamActive{conn: s.conn, stream: s.stream})
		default:
			return streams
		}
	}
}

func pumpStreamSends(server *Server, body []byte, streams []streamActive) []streamActive {
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
	return kept
}
