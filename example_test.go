package nwep_test

import (
	"fmt"

	"nwep"
)

// Example shows the five-minute hello-world, the managed on-ramp NWG0200.
//
// a server is started with Serve (the runtime owns the loop), a client connects
// and reads a path, and both shut down cleanly. this mirrors the simplest sandbox
// app, a single request and response over the loopback.
func Example() {
	// the server proves this identity in the handshake NW090000.
	serverID, _ := nwep.GenerateIdentity()
	defer serverID.Close()

	server, _ := nwep.NewServer().
		Identity(serverID).
		Bind(nwep.Loopback(0)).
		OnRequest(func(req *nwep.Message, res *nwep.Responder) {
			res.OK([]byte("hello from " + req.Path()))
		}).
		Serve()
	defer server.Shutdown()

	// a client dials the server by node_id at its loopback address.
	clientID, _ := nwep.GenerateIdentity()
	defer clientID.Close()

	client, _ := nwep.NewClient().
		Identity(clientID).
		Target(server.NodeID()).
		Address(nwep.Loopback(server.LocalPort())).
		Connect()
	defer client.Close()

	resp, _ := client.Send(nwep.Read, "/greeting", nil, nil)
	defer resp.Close()

	fmt.Println(string(resp.Body()))
	// Output: hello from /greeting
}

// ExampleRunningClient shows the managed client, whose loop a goroutine owns.
//
// Run (not Connect) returns a client whose tick loop the runtime owns, so Send is
// just a call from any goroutine and several may be in flight at once NWG0600.
func ExampleRunningClient() {
	serverID, _ := nwep.GenerateIdentity()
	defer serverID.Close()
	server, _ := nwep.NewServer().
		Identity(serverID).
		Bind(nwep.Loopback(0)).
		OnRequest(func(_ *nwep.Message, res *nwep.Responder) {
			res.OK([]byte("pong"))
		}).
		Serve()
	defer server.Shutdown()

	clientID, _ := nwep.GenerateIdentity()
	defer clientID.Close()

	client, _ := nwep.NewClient().
		Identity(clientID).
		Target(server.NodeID()).
		Address(nwep.Loopback(server.LocalPort())).
		Run() // the managed terminal, no tick loop to write
	defer client.Close()

	resp, _ := client.Send(nwep.Read, "/ping", nil, nil)
	defer resp.Close()
	fmt.Println(string(resp.Body()))
	// Output: pong
}

// ExampleServer_driven shows the driven layer, where the caller owns the loop.
//
// Build (not Serve) returns a server you tick yourself, so it folds into any
// reactor, the only portable floor across every target NWG0600 NWG1200. here the
// loop is a plain goroutine, but the same Fd, Tick, and NextTimeout drop into
// epoll, io_uring, or a select reactor unchanged.
func ExampleServer_driven() {
	id, _ := nwep.GenerateIdentity()
	defer id.Close()

	server, err := nwep.NewServer().
		Identity(id).
		Bind(nwep.Loopback(0)).
		OnRequest(func(req *nwep.Message, res *nwep.Responder) {
			res.OK(req.Body())
		}).
		Build()
	if err != nil {
		return
	}
	defer server.Close()

	// the fd and tick are yours to fold into an existing event loop.
	_ = server.Fd()
	_ = server.Tick(0)
	fmt.Println("listening on", server.LocalPort() != 0)
	// Output: listening on true
}
