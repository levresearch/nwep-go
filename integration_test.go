package nwep

import (
	"testing"
	"time"
)

// monoStart anchors the monotonic millisecond clock the ticks use.
var monoStart = time.Now()

// monoMs returns a monotonic millisecond timestamp for tick NW070000.
func monoMs() int64 { return int64(time.Since(monoStart) / time.Millisecond) }

// TestLoopbackRequestResponse drives a real server and client through the web/1
// handshake on the loopback and exchanges a request and a response. it is the
// proof that the driven layer wires the protocol end to end NWG0600.
func TestLoopbackRequestResponse(t *testing.T) {
	serverID, err := GenerateIdentity()
	if err != nil {
		t.Fatal(err)
	}
	defer serverID.Close()

	server, err := NewServer().
		Identity(serverID).
		Bind(Loopback(0)).
		OnRequest(func(req *Message, res *Responder) {
			res.OK([]byte("hello " + string(req.Body())))
		}).
		Build()
	if err != nil {
		t.Fatalf("build server: %v", err)
	}
	port := server.LocalPort()

	// the server handle is single-threaded, so one goroutine owns its tick loop
	// while the test goroutine drives the (separate) client NWG0900.
	stop := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		for {
			select {
			case <-stop:
				return
			default:
			}
			_ = server.Tick(monoMs())
			time.Sleep(500 * time.Microsecond)
		}
	}()
	defer func() {
		close(stop)
		<-stopped
		server.Close()
	}()

	clientID, err := GenerateIdentity()
	if err != nil {
		t.Fatal(err)
	}
	defer clientID.Close()

	client, err := NewClient().
		Identity(clientID).
		Target(serverID.NodeID()).
		Address(Loopback(port)).
		Connect()
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer client.Close()

	resp, err := client.Send(Read, "/greet", nil, []byte("world"))
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	defer resp.Close()

	if got := string(resp.Body()); got != "hello world" {
		t.Fatalf("response body = %q, want %q", got, "hello world")
	}

	// the server learned and verified the client's identity in the handshake.
	if !client.IsAlive() {
		t.Error("client reports the connection is not alive")
	}
}
