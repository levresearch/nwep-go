// the server, the listening half of a node NW070000.
//
// build one with the builder. Build returns a driven Server you tick yourself
// (this file). Serve hands it to a Runtime that owns the loop (runtime.go), the
// two-terminal rule of NWG0300. the request handler runs synchronously inside
// Tick and must not block, defer with Responder.Defer for slow work NWG0600.

package nwep

import (
	"fmt"
	"unsafe"

	"nwep/sys"
)

// Handler answers one decoded request by writing into res NW060000.
//
// it runs on the owning thread synchronously inside Tick and must not block. for
// a slow answer call res.Defer and deliver it later with Server.Respond.
type Handler func(req *Message, res *Responder)

// ServerMetrics is a pull-model snapshot of a server's counters.
type ServerMetrics = sys.ServerMetrics

// Server is a bound, listening node you drive with Tick (the driven layer).
type Server struct {
	ptr     unsafe.Pointer
	handler Handler
}

// ServerBuilder collects the options for a server, terminated by Build or Serve.
type ServerBuilder struct {
	identity  *Identity
	addr      Address
	hasAddr   bool
	handler   Handler
	reusePort bool
	fd        uintptr
	useFd     bool

	// the managed-dht options, used only by Serve (the managed terminal). when
	// dhtSet is true Serve attaches a dht on the owner goroutine and ticks it.
	dhtSet      bool
	dhtContacts []BootstrapEntry
	dhtSeq      uint64
	announce    Address
	hasAnnounce bool
}

// NewServer starts building a server NW070000.
func NewServer() *ServerBuilder { return &ServerBuilder{} }

// Identity sets the keypair the server proves ownership of in the handshake NW090000.
func (b *ServerBuilder) Identity(id *Identity) *ServerBuilder { b.identity = id; return b }

// Bind sets the udp address to listen on.
func (b *ServerBuilder) Bind(addr Address) *ServerBuilder {
	b.addr = addr
	b.hasAddr = true
	return b
}

// ReusePort binds with SO_REUSEPORT so several servers can share the port NW070000.
func (b *ServerBuilder) ReusePort() *ServerBuilder { b.reusePort = true; return b }

// AdoptSocket listens on a caller-owned udp socket instead of binding one.
func (b *ServerBuilder) AdoptSocket(fd uintptr) *ServerBuilder {
	b.fd = fd
	b.useFd = true
	return b
}

// OnRequest sets the request handler NW060000.
func (b *ServerBuilder) OnRequest(h Handler) *ServerBuilder { b.handler = h; return b }

// Dht attaches a managed dht seeded with contacts, a Serve-only option NW110000.
//
// when set, Serve attaches a dht on the owner goroutine, joins the network, and
// ticks it alongside the server, so RunningServer.Resolve can resolve a node_id to
// an address without the caller running any loop. ignored by Build (the driven
// terminal), attach a Dht yourself there with Server.AttachDht.
func (b *ServerBuilder) Dht(contacts []BootstrapEntry) *ServerBuilder {
	b.dhtSet = true
	b.dhtContacts = contacts
	return b
}

// DhtInitialSeq sets the managed dht's initial announcement sequence NW110300.
func (b *ServerBuilder) DhtInitialSeq(seq uint64) *ServerBuilder { b.dhtSeq = seq; return b }

// AnnounceAs makes the managed dht re-announce this service address periodically NW110300.
func (b *ServerBuilder) AnnounceAs(addr Address) *ServerBuilder {
	b.announce = addr
	b.hasAnnounce = true
	return b
}

// Build constructs a driven Server, the L1 terminal, you own the loop NWG0300.
//
// returns a Server ready to tick. close it with Close.
// errors with a config error when neither Bind nor AdoptSocket was set, and any
// listen error from the transport.
func (b *ServerBuilder) Build() (*Server, error) {
	if b.identity == nil {
		return nil, fmt.Errorf("nwep: server requires an Identity")
	}
	var ptr unsafe.Pointer
	var rc int
	switch {
	case b.useFd:
		ptr, rc = sys.ServerListenFd(b.identity.Raw(), b.fd)
	case b.reusePort && b.hasAddr:
		ptr, rc = sys.ServerListenReuseport(b.identity.Raw(), b.addr.sysAddr())
	case b.hasAddr:
		ptr, rc = sys.ServerListen(b.identity.Raw(), b.addr.sysAddr())
	default:
		return nil, fmt.Errorf("nwep: server requires Bind or AdoptSocket")
	}
	if err := check(rc); err != nil {
		return nil, err
	}
	s := &Server{ptr: ptr, handler: b.handler}
	if b.handler != nil {
		if rc := sys.ServerSetHandler(ptr, s.dispatch); rc != 0 {
			sys.ServerClose(ptr)
			return nil, check(rc)
		}
	}
	return s, nil
}

// dispatch is the sys handler trampoline, turning a c callback into a Handler call.
func (s *Server) dispatch(server unsafe.Pointer, connID, streamID uint64, req, buf unsafe.Pointer) int {
	r := &Message{ptr: req, connID: connID, streamID: streamID}
	res := &Responder{buf: buf, server: server, connID: connID, streamID: streamID}
	s.handler(r, res)
	if res.deferred {
		return sys.DeferSentinel
	}
	if res.err != nil {
		if e, ok := res.err.(*Error); ok {
			return e.Code
		}
		return sys.ErrInternal
	}
	return 0
}

// Tick advances the server at nowMs monotonic, the heart of the driven loop NW070000.
func (s *Server) Tick(nowMs int64) error { return check(sys.ServerTick(s.ptr, nowMs)) }

// Fd returns the udp socket handle to fold into your reactor NWG0600 NWG1200.
//
// it is a uintptr to hold a posix fd or a windows SOCKET, poll it for readability.
func (s *Server) Fd() uintptr { return sys.ServerFd(s.ptr) }

// NextTimeoutMs returns ms until the next timer fires, or negative for none NW070000.
func (s *Server) NextTimeoutMs(nowMs int64) int { return sys.ServerNextTimeoutMs(s.ptr, nowMs) }

// LocalPort returns the bound udp port, useful when binding port 0.
func (s *Server) LocalPort() uint16 { return sys.ServerLocalPort(s.ptr) }

// LocalNodeID returns the server's own node_id.
func (s *Server) LocalNodeID() (NodeID, error) {
	id, rc := sys.ServerLocalNodeid(s.ptr)
	if err := check(rc); err != nil {
		return NodeID{}, err
	}
	return NodeID(id), nil
}

// PeerNodeID returns the verified node_id of a connection's peer NW090000.
func (s *Server) PeerNodeID(connID uint64) (NodeID, error) {
	id, rc := sys.ServerGetPeerNodeid(s.ptr, connID)
	if err := check(rc); err != nil {
		return NodeID{}, err
	}
	return NodeID(id), nil
}

// Metrics returns a snapshot of the server's counters.
func (s *Server) Metrics() (ServerMetrics, error) {
	m, rc := sys.ServerMetricsGet(s.ptr)
	return m, check(rc)
}

// Load returns the current load gauge, 0 to 100.
func (s *Server) Load() int { return sys.ServerLoad(s.ptr) }

// SetOverloaded toggles the front-door shed switch NW060000.
func (s *Server) SetOverloaded(on bool) { sys.ServerSetOverloaded(s.ptr, on) }

// SetMaxParked caps how many deferred responses may be outstanding.
func (s *Server) SetMaxParked(n int) { sys.ServerSetMaxParked(s.ptr, n) }

// Drain begins a graceful drain, refusing new connections.
func (s *Server) Drain() error { return check(sys.ServerDrain(s.ptr)) }

// IsDrained reports whether a drain has completed.
func (s *Server) IsDrained() bool { return sys.ServerIsDrained(s.ptr) }

// LastHandshakeError returns the most recent handshake failure code, for diagnostics.
func (s *Server) LastHandshakeError() int { return sys.ServerLastHandshakeError(s.ptr) }

// Notify pushes a server-initiated NOTIFY on a connection NW060500.
func (s *Server) Notify(connID uint64, event string, headers [][2]string, body []byte) error {
	return check(sys.ServerNotify(s.ptr, connID, event, headers, body))
}

// Respond delivers a deferred response after a handler called Responder.Defer NW000017.
func (s *Server) Respond(connID, streamID uint64, status string, body []byte) error {
	return check(sys.ServerRespond(s.ptr, connID, streamID, status, body))
}

// RespondHeader appends a header to a deferred response before delivering it.
func (s *Server) RespondHeader(connID, streamID uint64, name, value string) error {
	return check(sys.ServerRespondHeader(s.ptr, connID, streamID, name, value))
}

// Relay delivers an upstream response to a deferred request verbatim.
func (s *Server) Relay(connID, streamID uint64, origin *Message) error {
	return check(sys.ServerRelay(s.ptr, connID, streamID, origin.ptr))
}

// BeginStream opens a server-pushed streamed response NW060200.
func (s *Server) BeginStream(connID, streamID uint64, path, status string, headers [][2]string) error {
	return check(sys.ServerBeginStream(s.ptr, connID, streamID, path, status, headers))
}

// StreamSend sends as much of body as flow control allows, returning the count NW060200.
//
// the body may be accepted in part. returns the number of bytes taken (0 when the
// stream is flow-blocked, retry on a later tick).
func (s *Server) StreamSend(connID, streamID uint64, body []byte) (int, error) {
	rc := sys.ServerStreamSend(s.ptr, connID, streamID, body)
	if rc < 0 {
		return 0, newError(rc)
	}
	return rc, nil
}

// StreamEnd finishes a streamed response with a quic fin.
func (s *Server) StreamEnd(connID, streamID uint64) error {
	return check(sys.ServerStreamEnd(s.ptr, connID, streamID))
}

// Close stops the server and frees it, retiring the handler (nwep_server_close).
func (s *Server) Close() {
	if s.ptr != nil {
		sys.ServerClose(s.ptr)
		s.ptr = nil
	}
}

// Raw returns the underlying sys server pointer, the no-cliffs escape to L0 NWG0200.
func (s *Server) Raw() unsafe.Pointer { return s.ptr }
