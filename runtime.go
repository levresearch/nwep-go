// the managed runtime, a goroutine that owns the loop for you NWG0200 NWG0600.
//
// the c handles are single-threaded and caller-driven, so they cannot be touched
// from any goroutine. the managed layer pins a handle to one owner goroutine that
// runs the real tick and poll loop, and the public api talks to it by message
// passing, a call sends a closure to the owner and waits for the reply (the
// actor-bridge rule of NWG0600). the owner goroutine is locked to its os thread
// because the abi is non-reentrant and not thread safe NWG0900.

package nwep

import (
	"runtime"
	"time"
)

// maxWaitMs caps how long the owner loop blocks on the socket between ticks, so a
// command sent to the owner is serviced within this bound without a self-pipe
// waker (no self-pipe needed, the command channel unblocks the select).
const maxWaitMs = 50

var monoEpoch = time.Now()

// monoClock returns a monotonic millisecond clock for tick NW070000.
func monoClock() int64 { return int64(time.Since(monoEpoch) / time.Millisecond) }

// nowSecs returns a unix-seconds clock for the dht, as its layer expects NW110000.
func nowSecs() uint64 { return uint64(time.Now().Unix()) } //nolint:gosec // Unix timestamps are non-negative

// command is a closure to run on the owner goroutine plus a channel to signal it ran.
type command struct {
	fn   func()
	done chan struct{}
}

// owner runs a handle's tick and poll loop on one locked os thread NWG0600 NWG0900.
type owner struct {
	cmds chan command
	stop chan struct{}
	done chan struct{}
}

// newOwner starts an owner goroutine running body, a tick-poll loop that checks
// the owner's stop and command channels each cycle.
func newOwner() *owner {
	return &owner{
		cmds: make(chan command, 16),
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
}

// run executes fn on the owner goroutine and waits for it, the actor-bridge call.
//
// returns false when the owner has already stopped, so the call could not run.
func (o *owner) run(fn func()) bool {
	cmd := command{fn: fn, done: make(chan struct{})}
	select {
	case o.cmds <- cmd:
	case <-o.done:
		return false
	}
	select {
	case <-cmd.done:
		return true
	case <-o.done:
		return false
	}
}

// drainCommands runs every queued command, called by the owner between ticks.
func (o *owner) drainCommands() {
	for {
		select {
		case cmd := <-o.cmds:
			cmd.fn()
			close(cmd.done)
		default:
			return
		}
	}
}

// RunningServer is a Server whose loop the runtime owns on a goroutine NWG0200 NWG0600.
//
// the server runs on its own owner goroutine, dispatching to the handler inside
// tick (so the handler must not block), until Shutdown. obtained from
// ServerBuilder.Serve. to call a server method from another goroutine, use Do,
// which marshals the call onto the owner.
type RunningServer struct {
	server *Server
	owner  *owner
	nodeID NodeID
	port   uint16
	// dht is non-nil when the server was built with ServerBuilder.Dht. it lives on
	// the owner goroutine and is only touched from inside owner.run closures, which
	// run there serialized with the loop's ticks (so no lock is needed).
	dht *Dht
}

// Serve builds the server and runs its loop on a goroutine, the L2 terminal NWG0300.
//
// the same builder chain that ends in Build (driven) ends here in Serve (managed).
// the server is built on the owner goroutine, so the handle never crosses a thread,
// then driven there with a real poll until Shutdown. returns the running server.
// errors with whatever Build would, surfaced from the owner goroutine.
func (b *ServerBuilder) Serve() (*RunningServer, error) {
	o := newOwner()
	rs := &RunningServer{owner: o}
	type result struct {
		nodeID NodeID
		port   uint16
		err    error
	}
	ready := make(chan result, 1)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		defer close(o.done)

		server, err := b.Build()
		if err != nil {
			ready <- result{err: err}
			return
		}
		// the dht borrows the server, so both live in this owner frame NW110000.
		var dht *Dht
		if b.dhtSet {
			dht, err = server.AttachDht(b.dhtContacts, b.dhtSeq)
			if err != nil {
				server.Close()
				ready <- result{err: err}
				return
			}
			_ = dht.Bootstrap(nowSecs())
		}
		rs.server, rs.dht = server, dht
		nodeID, _ := server.LocalNodeID()
		ready <- result{nodeID: nodeID, port: server.LocalPort()}

		// re-announce the service address on this cadence when AnnounceAs was set.
		const announceEverySecs = 60
		var lastAnnounce uint64

		for {
			select {
			case <-o.stop:
				if dht != nil {
					dht.Close()
				}
				server.Close()
				return
			default:
			}
			o.drainCommands()
			_ = server.Tick(monoClock())
			if dht != nil {
				now := nowSecs()
				_ = dht.Tick(now)
				if b.hasAnnounce && now-lastAnnounce >= announceEverySecs {
					_ = dht.Announce(b.announce, now)
					lastAnnounce = now
				}
			}
			wait := capWait(server.NextTimeoutMs(monoClock()))
			waitReadable(server.Fd(), wait)
		}
	}()

	r := <-ready
	if r.err != nil {
		<-o.done
		return nil, r.err
	}
	rs.nodeID, rs.port = r.nodeID, r.port
	return rs, nil
}

// NodeID returns the server's node_id.
func (rs *RunningServer) NodeID() NodeID { return rs.nodeID }

// LocalPort returns the bound udp port.
func (rs *RunningServer) LocalPort() uint16 { return rs.port }

// Do runs fn against the server on the owner goroutine, the safe cross-goroutine call.
//
// use it to call any Server method (Notify, a deferred Respond, Metrics) from a
// goroutine other than the owner. fn runs to completion on the owner before Do
// returns. returns false when the server has already shut down.
func (rs *RunningServer) Do(fn func(*Server)) bool {
	return rs.owner.run(func() { fn(rs.server) })
}

// Notify pushes a server-initiated NOTIFY from any goroutine NW060500.
func (rs *RunningServer) Notify(connID uint64, event string, headers [][2]string, body []byte) error {
	var err error
	rs.Do(func(s *Server) { err = s.Notify(connID, event, headers, body) })
	return err
}

// Respond delivers a deferred response from any goroutine NW000017.
func (rs *RunningServer) Respond(connID, streamID uint64, status string, body []byte) error {
	var err error
	rs.Do(func(s *Server) { err = s.Respond(connID, streamID, status, body) })
	return err
}

// Metrics returns a counter snapshot, scraped on the owner goroutine.
func (rs *RunningServer) Metrics() (ServerMetrics, error) {
	var m ServerMetrics
	var err error
	rs.Do(func(s *Server) { m, err = s.Metrics() })
	return m, err
}

// Resolve resolves a peer's node_id to an address through the managed dht NW110800.
//
// it runs an iterative lookup on the owner goroutine while the loop ticks the dht,
// so the resolution makes progress without the caller running anything. checks the
// local store first, then starts a lookup and polls it until it resolves or timeout
// elapses. requires the server was built with ServerBuilder.Dht.
//
// returns the resolved Address.
// errors with a config error when no managed dht is attached, and IdentityNotFound
// when the lookup times out with no record.
func (rs *RunningServer) Resolve(target NodeID, timeoutMs int) (Address, error) {
	if rs.dht == nil {
		return Address{}, ErrConfigMissing
	}
	deadline := monoClock() + int64(timeoutMs)
	started := false
	for {
		var addr Address
		var found bool
		var rerr error
		ok := rs.owner.run(func() {
			rec, has, err := rs.dht.LookupResult(target)
			if err != nil {
				rerr = err
				return
			}
			if has {
				addr, found = rec.Address, true
				return
			}
			if !started {
				rerr = rs.dht.StartLookup(target, nowSecs())
				started = true
			}
		})
		if !ok {
			return Address{}, ErrNetworkClosed
		}
		if rerr != nil {
			return Address{}, rerr
		}
		if found {
			return addr, nil
		}
		if monoClock() >= deadline {
			return Address{}, ErrIdentityNotFound
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// DhtMetrics returns a snapshot of the managed dht's counters, scraped on the owner.
//
// errors with a config error when no managed dht is attached.
func (rs *RunningServer) DhtMetrics() (DhtMetrics, error) {
	if rs.dht == nil {
		return DhtMetrics{}, ErrConfigMissing
	}
	var m DhtMetrics
	var err error
	rs.Do(func(*Server) { m, err = rs.dht.Metrics() })
	return m, err
}

// Shutdown stops the loop and waits for the owner goroutine to finish.
//
// the server is closed on the owner goroutine, so its teardown never races the
// loop. safe to call more than once.
func (rs *RunningServer) Shutdown() {
	select {
	case <-rs.owner.done:
		return
	default:
	}
	close(rs.owner.stop)
	<-rs.owner.done
}

// capWait caps a next-timeout (ms, negative for none) to the owner's max wait.
func capWait(timeoutMs int) time.Duration {
	if timeoutMs < 0 || timeoutMs > maxWaitMs {
		timeoutMs = maxWaitMs
	}
	return time.Duration(timeoutMs) * time.Millisecond
}
