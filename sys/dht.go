// the layer 0 dht calls, the kademlia node-to-address resolver NW110000.

package sys

/*
#include <stdlib.h>
#include <nwep.h>
*/
import "C"

import "unsafe"

// BootstrapEntry is a known node to seed the routing table, a node_id and address.
type BootstrapEntry struct {
	NodeID [NodeIDSize]byte
	Addr   Address
}

// DhtRecord is a resolved announcement, what a lookup yields NW110300.
type DhtRecord struct {
	NodeID    [NodeIDSize]byte
	Addr      Address
	Pubkey    [PubKeySize]byte
	Seq       uint64
	Timestamp uint64
}

// DhtMetrics mirrors nwep_dht_metrics, the dht's own datagram counters.
type DhtMetrics struct {
	DatagramsSent     uint64
	DatagramsReceived uint64
	BytesSent         uint64
	BytesReceived     uint64
}

// DhtParseBootstrap parses a "<nodeid>@[<ipv6>]:<port>" entry (nwep_dht_parse_bootstrap).
func DhtParseBootstrap(input string) (BootstrapEntry, int) {
	cs := C.CString(input)
	defer C.free(unsafe.Pointer(cs))
	var out C.nwep_bootstrap_entry
	rc := int(C.nwep_dht_parse_bootstrap(&out, cs, C.size_t(len(input))))
	var e BootstrapEntry
	if rc == 0 {
		e.NodeID = *(*[NodeIDSize]byte)(unsafe.Pointer(&out.node_id.bytes[0]))
		e.Addr = *(*Address)(unsafe.Pointer(&out.addr))
	}
	return e, rc
}

// DhtAttach attaches a dht to a server, seeding it with bootstrap nodes (nwep_dht_attach).
func DhtAttach(server unsafe.Pointer, bootstrap []BootstrapEntry, initialSeq uint64) (unsafe.Pointer, int) {
	var arr *C.nwep_bootstrap_entry
	if len(bootstrap) > 0 {
		c := make([]C.nwep_bootstrap_entry, len(bootstrap))
		for i, e := range bootstrap {
			c[i].node_id = nodeIDToC(e.NodeID)
			c[i].addr = *(*C.nwep_address)(unsafe.Pointer(&e.Addr))
		}
		arr = &c[0]
	}
	var out *C.nwep_dht
	rc := int(C.nwep_dht_attach(&out, (*C.nwep_server)(server), arr, C.size_t(len(bootstrap)), C.uint64_t(initialSeq)))
	return unsafe.Pointer(out), rc
}

// DhtBootstrap kicks off the initial self-lookup to populate routing (nwep_dht_bootstrap).
func DhtBootstrap(dht unsafe.Pointer, nowSecs uint64) int {
	return int(C.nwep_dht_bootstrap((*C.nwep_dht)(dht), C.uint64_t(nowSecs)))
}

// DhtAnnounce publishes this node's service address to the dht (nwep_dht_announce).
func DhtAnnounce(dht unsafe.Pointer, serviceAddr Address, nowSecs uint64) int {
	return int(C.nwep_dht_announce((*C.nwep_dht)(dht), serviceAddr.c(), C.uint64_t(nowSecs)))
}

// DhtStartLookup begins an iterative find_value for a node_id (nwep_dht_start_lookup).
func DhtStartLookup(dht unsafe.Pointer, target [NodeIDSize]byte, nowSecs uint64) int {
	id := nodeIDToC(target)
	return int(C.nwep_dht_start_lookup((*C.nwep_dht)(dht), &id, C.uint64_t(nowSecs)))
}

// DhtLookupResult fetches a completed lookup's record, if resolved (nwep_dht_lookup_result).
func DhtLookupResult(dht unsafe.Pointer, target [NodeIDSize]byte) (DhtRecord, int) {
	id := nodeIDToC(target)
	var out C.nwep_dht_record
	rc := int(C.nwep_dht_lookup_result((*C.nwep_dht)(dht), &id, &out))
	var r DhtRecord
	if rc == 0 {
		r.NodeID = *(*[NodeIDSize]byte)(unsafe.Pointer(&out.node_id.bytes[0]))
		r.Addr = *(*Address)(unsafe.Pointer(&out.addr))
		r.Pubkey = *(*[PubKeySize]byte)(unsafe.Pointer(&out.pubkey[0]))
		r.Seq = uint64(out.seq)
		r.Timestamp = uint64(out.timestamp)
	}
	return r, rc
}

// DhtTick advances the dht state machine at now_secs unix time (nwep_dht_tick).
func DhtTick(dht unsafe.Pointer, nowSecs uint64) int {
	return int(C.nwep_dht_tick((*C.nwep_dht)(dht), C.uint64_t(nowSecs)))
}

// DhtNextTimeoutMs returns ms until the next dht timer, or negative for none (nwep_dht_next_timeout_ms).
func DhtNextTimeoutMs(dht unsafe.Pointer, nowSecs uint64) int {
	return int(C.nwep_dht_next_timeout_ms((*C.nwep_dht)(dht), C.uint64_t(nowSecs)))
}

// DhtMetricsGet fills a metrics snapshot from the dht (nwep_dht_metrics_get).
func DhtMetricsGet(dht unsafe.Pointer) (DhtMetrics, int) {
	var m C.nwep_dht_metrics
	rc := int(C.nwep_dht_metrics_get((*C.nwep_dht)(dht), &m))
	return DhtMetrics{
		DatagramsSent:     uint64(m.datagrams_sent),
		DatagramsReceived: uint64(m.datagrams_received),
		BytesSent:         uint64(m.bytes_sent),
		BytesReceived:     uint64(m.bytes_received),
	}, rc
}

// DhtClose detaches and frees the dht (nwep_dht_close).
func DhtClose(dht unsafe.Pointer) {
	C.nwep_dht_close((*C.nwep_dht)(dht))
}
