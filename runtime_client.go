// the managed client, a connection whose loop a goroutine owns NWG0200 NWG0600.
//
// the driven Client (client.go) makes the caller run the loop. RunningClient pins
// the connection to one owner goroutine that runs the tick and poll loop, and the
// public Send marshals onto it through the actor bridge. the owner submits every
// request without blocking and polls all in flight at once, so two goroutines that
// Send concurrently run their requests in parallel on the one connection.

package nwep

import (
	"runtime"
)

// clientCmd is one queued send, the request plus where to deliver its response.
type clientCmd struct {
	method  Method
	path    string
	headers [][2]string
	body    []byte
	reply   chan clientResult
}

// clientResult is a completed send, a response or the error that ended it.
type clientResult struct {
	resp *Message
	err  error
}

// RunningClient is a connection whose tick and poll loop the runtime owns NWG0600.
//
// obtained from ClientBuilder.Run. Send is safe from any goroutine and several may
// be in flight at once. close it with Close.
type RunningClient struct {
	cmds chan clientCmd
	stop chan struct{}
	done chan struct{}
}

// Run connects on a dedicated goroutine and returns a managed client, the L2
// terminal NWG0300.
//
// the same builder chain that ends in Connect (driven) ends here in Run (managed).
// the blocking handshake runs on the owner goroutine, and Run returns once the
// connection is up. the owner then keeps many requests in flight at once.
// errors with whatever Connect would, surfaced from the owner goroutine.
func (b *ClientBuilder) Run() (*RunningClient, error) {
	rc := &RunningClient{
		cmds: make(chan clientCmd),
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	ready := make(chan error, 1)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		defer close(rc.done)

		client, err := b.Connect()
		if err != nil {
			ready <- err
			return
		}
		ready <- nil
		rc.loop(client)
	}()

	if err := <-ready; err != nil {
		<-rc.done
		return nil, err
	}
	return rc, nil
}

// loop is the owner goroutine's body, submit, tick, poll all in flight, repeat.
func (rc *RunningClient) loop(client *Client) {
	// in flight requests by id, each mapped to the channel awaiting its reply.
	pending := map[uint64]chan clientResult{}

	failAll := func(err error) {
		for id, reply := range pending {
			reply <- clientResult{err: err}
			delete(pending, id)
		}
	}

	for {
		select {
		case <-rc.stop:
			failAll(ErrNetworkClosed)
			client.Close()
			return
		default:
		}

		// drain newly enqueued sends, submitting each without blocking.
		draining := true
		for draining {
			select {
			case cmd := <-rc.cmds:
				id, err := client.Submit(cmd.method, cmd.path, cmd.headers, cmd.body)
				if err != nil {
					cmd.reply <- clientResult{err: err}
					continue
				}
				pending[id] = cmd.reply
			default:
				draining = false
			}
		}

		// a terminally closed connection cannot recover, fail in flight and exit.
		if client.Tick(monoClock()) != nil || !client.IsAlive() {
			failAll(ErrNetworkClosed)
			client.Close()
			return
		}

		// poll every in flight request, deliver and retire the finished ones.
		for id, reply := range pending {
			resp, ready, err := client.Poll(id)
			switch {
			case err != nil:
				reply <- clientResult{err: err}
				delete(pending, id)
			case ready:
				reply <- clientResult{resp: resp}
				delete(pending, id)
			}
		}

		wait := maxWaitMs
		if t := client.NextTimeoutMs(monoClock()); t >= 0 && t < wait {
			wait = t
		}
		waitReadable(client.Fd(), capWait(wait))
	}
}

// Send sends a request and waits for its response, safe from any goroutine NW060000.
//
// the owner submits it without blocking and polls it alongside any other in flight
// requests, so awaiting two Sends concurrently runs them in parallel on the one
// connection. returns an owned response Message, close it with Close.
// errors NetworkClosed once the client has shut down, and any transport error the
// request itself produces.
func (rc *RunningClient) Send(method Method, path string, headers [][2]string, body []byte) (*Message, error) {
	reply := make(chan clientResult, 1)
	cmd := clientCmd{method: method, path: path, headers: headers, body: body, reply: reply}
	select {
	case rc.cmds <- cmd:
	case <-rc.done:
		return nil, ErrNetworkClosed
	}
	select {
	case r := <-reply:
		return r.resp, r.err
	case <-rc.done:
		return nil, ErrNetworkClosed
	}
}

// Close stops the owner goroutine and tears the connection down on it.
//
// in flight sends fail with NetworkClosed rather than hanging. safe to call more
// than once.
func (rc *RunningClient) Close() {
	select {
	case <-rc.done:
		return
	default:
	}
	close(rc.stop)
	<-rc.done
}
