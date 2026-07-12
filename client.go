// the client, the connecting half of a node NW070000.
//
// build one with the builder, terminated by Connect (a blocking connect that runs
// the handshake and returns a driven Client). after connecting, Send is blocking,
// or submit non-blocking requests and drive them with Tick NWG0600.

package nwep

import (
	"fmt"
	"unsafe"

	"nwep/sys"
)

// Method is a web/1 request method NW040400.
type Method int

// the request methods NW040400.
const (
	Read      Method = Method(sys.MethodRead)
	Write     Method = Method(sys.MethodWrite)
	Update    Method = Method(sys.MethodUpdate)
	Delete    Method = Method(sys.MethodDelete)
	Heartbeat Method = Method(sys.MethodHeartbeat)
	Head      Method = Method(sys.MethodHead)
)

// ClientMetrics is a pull-model snapshot of one connection's counters.
type ClientMetrics = sys.ClientMetrics

// Client is a connected peer you send requests to (the driven layer).
type Client struct {
	ptr unsafe.Pointer
}

// ClientBuilder collects the options for a connection, terminated by Connect.
type ClientBuilder struct {
	identity  *Identity
	target    NodeID
	hasTarget bool
	addr      Address
	hasAddr   bool
	dht       *Dht
	timeoutMs uint32
}

// NewClient starts building a connection NW070000.
func NewClient() *ClientBuilder { return &ClientBuilder{timeoutMs: 5000} }

// Identity sets the keypair the client proves ownership of in the handshake NW090000.
func (b *ClientBuilder) Identity(id *Identity) *ClientBuilder { b.identity = id; return b }

// Target sets the node_id to connect to, verified during the handshake NW090000.
func (b *ClientBuilder) Target(nodeID NodeID) *ClientBuilder {
	b.target = nodeID
	b.hasTarget = true
	return b
}

// Address sets the peer's udp address, skipping dht resolution.
func (b *ClientBuilder) Address(addr Address) *ClientBuilder {
	b.addr = addr
	b.hasAddr = true
	return b
}

// ResolveWith sets a dht to resolve the target's address when no Address is set NW110000.
func (b *ClientBuilder) ResolveWith(dht *Dht, lookupTimeoutMs uint32) *ClientBuilder {
	b.dht = dht
	b.timeoutMs = lookupTimeoutMs
	return b
}

// Connect runs the handshake and returns a driven Client, blocking (nwep_client_connect).
//
// when an Address is set it dials directly, otherwise it resolves the target
// through the dht set with ResolveWith. returns a connected Client, close it with
// Close.
// errors with a config error when the target or destination is missing, the fatal
// handshake codes on identity or signature failure, and a network error on timeout.
func (b *ClientBuilder) Connect() (*Client, error) {
	if b.identity == nil || !b.hasTarget {
		return nil, fmt.Errorf("nwep: client requires an Identity and a Target")
	}
	var ptr unsafe.Pointer
	var rc int
	switch {
	case b.hasAddr:
		ptr, rc = sys.ClientConnect(b.identity.Raw(), b.target, b.addr.sysAddr())
	case b.dht != nil:
		ptr, rc = sys.ClientConnectByNodeid(b.identity.Raw(), b.target, b.dht.ptr, b.timeoutMs)
	default:
		return nil, fmt.Errorf("nwep: client requires an Address or ResolveWith a dht")
	}
	if err := check(rc); err != nil {
		return nil, err
	}
	return &Client{ptr: ptr}, nil
}

// Send sends a request and blocks for the response (nwep_client_send).
//
// returns an owned response Message, close it with Close.
// errors with the protocol and network codes, fatal on a signature failure.
func (c *Client) Send(method Method, path string, headers [][2]string, body []byte) (*Message, error) {
	ptr, rc := sys.ClientSend(c.ptr, int(method), path, headers, body)
	if err := check(rc); err != nil {
		return nil, err
	}
	return &Message{ptr: ptr, owned: true}, nil
}

// Tick advances the connection at nowMs monotonic, for the async request path NW070000.
func (c *Client) Tick(nowMs int64) error { return check(sys.ClientTick(c.ptr, nowMs)) }

// Fd returns the udp socket handle to fold into your reactor NWG0600 NWG1200.
func (c *Client) Fd() uintptr { return sys.ClientFd(c.ptr) }

// NextTimeoutMs returns ms until the next timer, or negative for none.
func (c *Client) NextTimeoutMs(nowMs int64) int { return sys.ClientNextTimeoutMs(c.ptr, nowMs) }

// IsAlive reports whether the connection is still usable.
func (c *Client) IsAlive() bool { return sys.ClientIsAlive(c.ptr) }

// Metrics returns a snapshot of the connection's counters.
func (c *Client) Metrics() (ClientMetrics, error) {
	m, rc := sys.ClientMetricsGet(c.ptr)
	return m, check(rc)
}

// PeerPubkey returns the verified ed25519 public key of the peer NW090000.
func (c *Client) PeerPubkey() ([32]byte, error) {
	pk, rc := sys.ClientPeerPubkey(c.ptr)
	return pk, check(rc)
}

// Compression reports whether the connection negotiated body compression.
func (c *Client) Compression() bool { return sys.ClientCompression(c.ptr) > 0 }

// VerifyResponse checks a response signature against the peer for path NW060900.
func (c *Client) VerifyResponse(resp *Message, path string, nowSecs uint64) error {
	return check(sys.ClientVerifyResponse(c.ptr, resp.ptr, path, nowSecs))
}

// PollNotify returns the next buffered server NOTIFY, or nil NW060500.
func (c *Client) PollNotify() *Message {
	ptr := sys.ClientPollNotify(c.ptr)
	if ptr == nil {
		return nil
	}
	return &Message{ptr: ptr, owned: true}
}

// Submit queues a non-blocking request and returns its id (nwep_client_request_submit).
func (c *Client) Submit(method Method, path string, headers [][2]string, body []byte) (uint64, error) {
	id, rc := sys.ClientRequestSubmit(c.ptr, int(method), path, headers, body)
	return id, check(rc)
}

// Poll checks a submitted request, returning the response when ready (nwep_client_request_poll).
//
// returns the response and true when ready, a nil message and false on would-block.
func (c *Client) Poll(id uint64) (*Message, bool, error) {
	ptr, rc := sys.ClientRequestPoll(c.ptr, id)
	if rc == sys.ErrWouldBlock {
		return nil, false, nil
	}
	if err := check(rc); err != nil {
		return nil, false, err
	}
	return &Message{ptr: ptr, owned: true}, true, nil
}

// Cancel abandons a submitted request by id.
func (c *Client) Cancel(id uint64) { sys.ClientRequestCancel(c.ptr, id) }

// SetCache attaches a response cache, or detaches it with nil NW060900.
func (c *Client) SetCache(cache *Cache) error {
	var p unsafe.Pointer
	if cache != nil {
		p = cache.ptr
	}
	return check(sys.ClientSetCache(c.ptr, p))
}

// OpenStream opens a streamed read, returning the stream id NW060200.
func (c *Client) OpenStream(method Method, path string, headers [][2]string) (uint64, error) {
	id, rc := sys.ClientOpenStream(c.ptr, int(method), path, headers)
	return id, check(rc)
}

// StreamResponse fetches the response headers of an opened stream.
func (c *Client) StreamResponse(streamID uint64) (*Message, error) {
	ptr, rc := sys.ClientStreamResponse(c.ptr, streamID)
	if err := check(rc); err != nil {
		return nil, err
	}
	return &Message{ptr: ptr, owned: true}, nil
}

// StreamRecv reads the next chunk of a stream into buf, reporting end of stream.
func (c *Client) StreamRecv(streamID uint64, buf []byte) (n int, ended bool, err error) {
	n, ended, rc := sys.ClientStreamRecv(c.ptr, streamID, buf)
	return n, ended, check(rc)
}

// StreamVerify checks the running trailer signature over a stream against pubkey NW060900.
func (c *Client) StreamVerify(streamID uint64, pubkey [32]byte) error {
	return check(sys.ClientStreamVerify(c.ptr, streamID, pubkey))
}

// StreamClose releases a stream's client-side state.
func (c *Client) StreamClose(streamID uint64) { sys.ClientStreamClose(c.ptr, streamID) }

// Close tears down the connection and frees it (nwep_client_close).
func (c *Client) Close() {
	if c.ptr != nil {
		sys.ClientClose(c.ptr)
		c.ptr = nil
	}
}

// Raw returns the underlying sys client pointer, the no-cliffs escape to L0 NWG0200.
func (c *Client) Raw() unsafe.Pointer { return c.ptr }
