// the managed dht quickstart, resolve a peer by node_id alone. a server whose
// runtime owns an attached dht, so Resolve(node_id) is a single call, no tick loop.
// one node announces itself into the dht and a second node, seeded with the first
// as a bootstrap contact, resolves it back to an address NW110800. both nodes
// run in this one process on fixed loopback ports for a self-contained demo.

package main

import (
	"fmt"
	"log"

	"nwep"
)

const portA, portB = 29441, 29442

func main() {
	// the announcer: self-bootstraps (its only contact is itself) and announces
	// its own address into the dht.
	idA, err := nwep.GenerateIdentity()
	if err != nil {
		log.Fatal(err)
	}
	defer idA.Close()
	selfA := nwep.BootstrapEntry{NodeID: idA.NodeID(), Address: nwep.Loopback(portA)}

	announcer, err := nwep.NewServer().
		Identity(idA).
		Bind(nwep.Loopback(portA)).
		OnRequest(func(_ *nwep.Message, res *nwep.Responder) { res.OK(nil) }).
		Dht([]nwep.BootstrapEntry{selfA}).
		AnnounceAs(nwep.Loopback(portA)).
		Serve()
	if err != nil {
		log.Fatalf("announcer (port %d may be in use): %v", portA, err)
	}
	defer announcer.Shutdown()
	fmt.Printf("announcer      %s on [::1]:%d\n", idA.NodeID(), portA)

	// the resolver: its runtime owns a dht seeded with the announcer as a contact.
	resolverID, err := nwep.GenerateIdentity()
	if err != nil {
		log.Fatal(err)
	}
	defer resolverID.Close()

	resolver, err := nwep.NewServer().
		Identity(resolverID).
		Bind(nwep.Loopback(portB)).
		OnRequest(func(_ *nwep.Message, res *nwep.Responder) { res.OK(nil) }).
		Dht([]nwep.BootstrapEntry{selfA}).
		Serve()
	if err != nil {
		log.Fatalf("resolver (port %d may be in use): %v", portB, err)
	}
	defer resolver.Shutdown()

	// resolve the announcer's node_id through the managed dht, a single call.
	addr, err := resolver.Resolve(idA.NodeID(), 5000)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("resolved       %s -> [::1]:%d\n", idA.NodeID(), addr.Port())

	m, err := resolver.DhtMetrics()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("dht traffic    %d sent, %d received\n", m.DatagramsSent, m.DatagramsReceived)

	fmt.Println("shutdown       clean")
}
