package nwep

import (
	"fmt"
	"sync"
	"testing"
)

// startEchoServer is a managed server that echoes "<prefix><body>" for every read,
// returning it and its node_id and port. the caller shuts it down.
func startEchoServer(t *testing.T, prefix string) *RunningServer {
	t.Helper()
	id, err := GenerateIdentity()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(id.Close)
	rs, err := NewServer().
		Identity(id).
		Bind(Loopback(0)).
		OnRequest(func(req *Message, res *Responder) {
			res.OK([]byte(prefix + string(req.Body())))
		}).
		Serve()
	if err != nil {
		t.Fatalf("serve: %v", err)
	}
	return rs
}

// TestManagedClient proves the managed client connects and sends over its own loop.
func TestManagedClient(t *testing.T) {
	server := startEchoServer(t, "echo ")
	defer server.Shutdown()

	clientID, _ := GenerateIdentity()
	defer clientID.Close()

	client, err := NewClient().
		Identity(clientID).
		Target(server.NodeID()).
		Address(Loopback(server.LocalPort())).
		Run() // the managed terminal
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	defer client.Close()

	resp, err := client.Send(Read, "/x", nil, []byte("one"))
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	defer resp.Close()
	if got := string(resp.Body()); got != "echo one" {
		t.Fatalf("body = %q, want %q", got, "echo one")
	}
}

// TestManagedClientConcurrent proves several sends run in flight on one connection.
func TestManagedClientConcurrent(t *testing.T) {
	server := startEchoServer(t, "r:")
	defer server.Shutdown()

	clientID, _ := GenerateIdentity()
	defer clientID.Close()
	client, err := NewClient().
		Identity(clientID).
		Target(server.NodeID()).
		Address(Loopback(server.LocalPort())).
		Run()
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	defer client.Close()

	// fire many sends from separate goroutines, the owner keeps them all in flight.
	const n = 16
	var wg sync.WaitGroup
	errs := make([]error, n)
	got := make([]string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			resp, err := client.Send(Read, "/x", nil, []byte(fmt.Sprintf("%d", i)))
			if err != nil {
				errs[i] = err
				return
			}
			got[i] = string(resp.Body())
			resp.Close()
		}(i)
	}
	wg.Wait()
	for i := 0; i < n; i++ {
		if errs[i] != nil {
			t.Fatalf("send %d: %v", i, errs[i])
		}
		if want := fmt.Sprintf("r:%d", i); got[i] != want {
			t.Fatalf("send %d body = %q, want %q", i, got[i], want)
		}
	}
}

// TestManagedDhtResolve proves a managed dht resolves a node_id it announced.
//
// the announcing node self-bootstraps and announces its address, a second node
// seeded with the first as a bootstrap contact resolves it by node_id alone. the
// managed dht needs at least one contact and a real port, so this binds fixed
// loopback ports NW110800.
func TestManagedDhtResolve(t *testing.T) {
	const portA, portB = 29411, 29412

	// the announcing node, self-bootstrapped (contact is itself) and announcing
	// its own address into the dht.
	idA, _ := GenerateIdentity()
	defer idA.Close()
	selfA := BootstrapEntry{NodeID: idA.NodeID(), Address: Loopback(portA)}
	serverA, err := NewServer().
		Identity(idA).
		Bind(Loopback(portA)).
		OnRequest(func(_ *Message, res *Responder) { res.OK(nil) }).
		Dht([]BootstrapEntry{selfA}).
		AnnounceAs(Loopback(portA)).
		Serve()
	if err != nil {
		t.Skipf("serve A (port %d may be in use): %v", portA, err)
	}
	defer serverA.Shutdown()

	// the resolving node, seeded with A as its bootstrap contact.
	idB, _ := GenerateIdentity()
	defer idB.Close()
	serverB, err := NewServer().
		Identity(idB).
		Bind(Loopback(portB)).
		OnRequest(func(_ *Message, res *Responder) { res.OK(nil) }).
		Dht([]BootstrapEntry{selfA}).
		Serve()
	if err != nil {
		t.Skipf("serve B (port %d may be in use): %v", portB, err)
	}
	defer serverB.Shutdown()

	// resolve A's node_id through B's managed dht, a real find_value lookup that
	// returns A's announced record.
	addr, err := serverB.Resolve(idA.NodeID(), 5000)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if addr.Port() != portA {
		t.Fatalf("resolved port %d, want %d", addr.Port(), portA)
	}
}

// TestResolveWithoutDht confirms Resolve errors cleanly when no dht is attached.
func TestResolveWithoutDht(t *testing.T) {
	server := startEchoServer(t, "")
	defer server.Shutdown()
	if _, err := server.Resolve(server.NodeID(), 100); err == nil {
		t.Fatal("expected a config error resolving without a managed dht")
	}
}
