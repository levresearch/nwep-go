package nwep

import "testing"

// TestManagedServe proves the L2 managed terminal owns the loop end to end.
//
// the server is started with Serve (the runtime owns the goroutine and poll loop),
// a driven client connects and exchanges a request, and the actor-bridge Do path
// scrapes metrics from the test goroutine. this exercises NWG0200 NWG0600.
func TestManagedServe(t *testing.T) {
	serverID, err := GenerateIdentity()
	if err != nil {
		t.Fatal(err)
	}
	defer serverID.Close()

	rs, err := NewServer().
		Identity(serverID).
		Bind(Loopback(0)).
		OnRequest(func(req *Message, res *Responder) {
			res.OK([]byte("managed " + string(req.Body())))
		}).
		Serve()
	if err != nil {
		t.Fatalf("serve: %v", err)
	}
	defer rs.Shutdown()

	if rs.NodeID() != serverID.NodeID() {
		t.Fatal("running server reports a different node_id")
	}

	clientID, err := GenerateIdentity()
	if err != nil {
		t.Fatal(err)
	}
	defer clientID.Close()

	client, err := NewClient().
		Identity(clientID).
		Target(serverID.NodeID()).
		Address(Loopback(rs.LocalPort())).
		Connect()
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer client.Close()

	resp, err := client.Send(Read, "/x", nil, []byte("hello"))
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	defer resp.Close()
	if got := string(resp.Body()); got != "managed hello" {
		t.Fatalf("body = %q, want %q", got, "managed hello")
	}

	// the actor-bridge Do path reaches the handle from this goroutine safely.
	m, err := rs.Metrics()
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}
	if m.ConnectionsAccepted == 0 {
		t.Error("expected at least one accepted connection in the metrics")
	}
}
