// the managed (L2) happy path, the five minute quickstart. a server and client,
// no tick loop, no fd, no goroutine in sight, the runtime owns the loops behind the
// actor bridge NWG0600. run it and it serves, connects, sends, and shuts down.

package main

import (
	"fmt"
	"log"

	"github.com/levresearch/nwep-go"
)

func main() {
	// a server that answers read /hello, running on its own owned loop.
	serverID, err := nwep.GenerateIdentity()
	if err != nil {
		log.Fatal(err)
	}
	defer serverID.Close()

	server, err := nwep.NewServer().
		Identity(serverID).
		Bind(nwep.Loopback(0)).
		OnRequest(func(req *nwep.Message, res *nwep.Responder) {
			if req.Path() == "/hello" {
				res.OK([]byte("hi from the managed runtime"))
			} else {
				res.NotFound()
			}
		}).
		Serve()
	if err != nil {
		log.Fatal(err)
	}
	defer server.Shutdown()
	fmt.Printf("serving        %s on [::1]:%d\n", server.NodeID(), server.LocalPort())

	// a managed client that dials it and sends, no loop to write.
	clientID, err := nwep.GenerateIdentity()
	if err != nil {
		log.Fatal(err)
	}
	defer clientID.Close()

	client, err := nwep.NewClient().
		Identity(clientID).
		Target(server.NodeID()).
		Address(nwep.Loopback(server.LocalPort())).
		Run()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	resp, err := client.Send(nwep.Read, "/hello", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Close()
	fmt.Printf("read /hello    %s %q\n", resp.Status(), resp.Body())

	fmt.Println("shutdown       clean")
}
