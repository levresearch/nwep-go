// the dht, the kademlia node-to-address resolver attached to a server NW110000.
//
// a Dht shares the server's udp socket (first-byte demux) and is driven with its
// own Tick on a unix-seconds clock, distinct from the transport's monotonic tick.

package nwep

import (
	"unsafe"

	"nwep/sys"
)

// BootstrapEntry is a known node to seed the routing table, a node and address.
type BootstrapEntry struct {
	NodeID  NodeID
	Address Address
}

// DhtRecord is a resolved announcement, what a lookup yields NW110300.
type DhtRecord struct {
	NodeID    NodeID
	Address   Address
	Pubkey    [32]byte
	Seq       uint64
	Timestamp uint64
}

// DhtMetrics mirrors the dht's own datagram counters.
type DhtMetrics = sys.DhtMetrics

// Dht is a routing table sharing a server's socket, driven on unix time NW110000.
type Dht struct {
	ptr unsafe.Pointer
}

// ParseBootstrap parses a "<nodeid>@[<ipv6>]:<port>" bootstrap entry NW110900.
func ParseBootstrap(input string) (BootstrapEntry, error) {
	e, rc := sys.DhtParseBootstrap(input)
	if err := check(rc); err != nil {
		return BootstrapEntry{}, err
	}
	return BootstrapEntry{NodeID: NodeID(e.NodeID), Address: Address{}.fromSys(e.Addr)}, nil
}

// fromSys wraps a raw sys address, a small adapter for the dht conversions.
func (Address) fromSys(a sys.Address) Address { return Address{raw: a} }

// AttachDht attaches a dht to the server, seeding it with bootstrap nodes NW110000.
//
// returns the dht, sharing this server's socket. call Bootstrap then drive it with
// Tick. close it with Close before the server.
func (s *Server) AttachDht(bootstrap []BootstrapEntry, initialSeq uint64) (*Dht, error) {
	sysBoot := make([]sys.BootstrapEntry, len(bootstrap))
	for i, e := range bootstrap {
		sysBoot[i] = sys.BootstrapEntry{NodeID: e.NodeID, Addr: e.Address.sysAddr()}
	}
	ptr, rc := sys.DhtAttach(s.ptr, sysBoot, initialSeq)
	if err := check(rc); err != nil {
		return nil, err
	}
	return &Dht{ptr: ptr}, nil
}

// Bootstrap kicks off the initial self-lookup to populate routing NW110000.
func (d *Dht) Bootstrap(nowSecs uint64) error { return check(sys.DhtBootstrap(d.ptr, nowSecs)) }

// Announce publishes this node's service address to the dht NW110300.
func (d *Dht) Announce(serviceAddr Address, nowSecs uint64) error {
	return check(sys.DhtAnnounce(d.ptr, serviceAddr.sysAddr(), nowSecs))
}

// StartLookup begins an iterative find_value for a node_id NW110000.
func (d *Dht) StartLookup(target NodeID, nowSecs uint64) error {
	return check(sys.DhtStartLookup(d.ptr, target, nowSecs))
}

// LookupResult fetches a completed lookup's record, if resolved NW110300.
//
// it is a poll, so any non-success return means the lookup has not resolved yet (no
// record cached, or the find_value is still in flight), reported as not-found
// rather than an error. returns the record and true when resolved, a zero record
// and false while still looking.
func (d *Dht) LookupResult(target NodeID) (DhtRecord, bool, error) {
	r, rc := sys.DhtLookupResult(d.ptr, target)
	if rc != 0 {
		return DhtRecord{}, false, nil
	}
	return DhtRecord{
		NodeID:    NodeID(r.NodeID),
		Address:   Address{raw: r.Addr},
		Pubkey:    r.Pubkey,
		Seq:       r.Seq,
		Timestamp: r.Timestamp,
	}, true, nil
}

// Tick advances the dht at nowSecs unix time NW110000.
func (d *Dht) Tick(nowSecs uint64) error { return check(sys.DhtTick(d.ptr, nowSecs)) }

// NextTimeoutMs returns ms until the next dht timer, or negative for none.
func (d *Dht) NextTimeoutMs(nowSecs uint64) int { return sys.DhtNextTimeoutMs(d.ptr, nowSecs) }

// Metrics returns a snapshot of the dht's datagram counters.
func (d *Dht) Metrics() (DhtMetrics, error) {
	m, rc := sys.DhtMetricsGet(d.ptr)
	return m, check(rc)
}

// Close detaches and frees the dht (nwep_dht_close).
func (d *Dht) Close() {
	if d.ptr != nil {
		sys.DhtClose(d.ptr)
		d.ptr = nil
	}
}

// Raw returns the underlying sys dht pointer, the no-cliffs escape to L0 NWG0200.
func (d *Dht) Raw() unsafe.Pointer { return d.ptr }
