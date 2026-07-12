// the layer 0 server calls, the listening half of a node NW070000.

package sys

/*
#include <stdlib.h>
#include <nwep.h>
*/
import "C"

import "unsafe"

// ServerMetrics mirrors nwep_server_metrics, a pull-model snapshot of counters.
type ServerMetrics struct {
	ConnectionsActive   uint64
	ConnectionsAccepted uint64
	ConnectionsRefused  uint64
	ConnectionsClosed   uint64
	BytesReceived       uint64
	BytesSent           uint64
	DatagramsReceived   uint64
	DatagramsSent       uint64
	RequestsDispatched  uint64
	RequestsShed        uint64
	ParkedActive        uint64
	Load                int32
}

// ServerListen binds a udp socket and allocates a server (nwep_server_listen).
func ServerListen(identity unsafe.Pointer, addr Address) (unsafe.Pointer, int) {
	var out *C.nwep_server
	rc := int(C.nwep_server_listen(&out, (*C.nwep_keypair)(identity), addr.c()))
	return unsafe.Pointer(out), rc
}

// ServerListenReuseport binds with SO_REUSEPORT for multi-socket sharding (nwep_server_listen_reuseport).
func ServerListenReuseport(identity unsafe.Pointer, addr Address) (unsafe.Pointer, int) {
	var out *C.nwep_server
	rc := int(C.nwep_server_listen_reuseport(&out, (*C.nwep_keypair)(identity), addr.c()))
	return unsafe.Pointer(out), rc
}

// ServerListenFd adopts a caller-owned udp socket (nwep_server_listen_fd).
func ServerListenFd(identity unsafe.Pointer, fd uintptr) (unsafe.Pointer, int) {
	var out *C.nwep_server
	rc := int(C.nwep_server_listen_fd(&out, (*C.nwep_keypair)(identity), C.uintptr_t(fd)))
	return unsafe.Pointer(out), rc
}

// ServerListenFdSharded adopts a socket as one shard of a reuseport set (nwep_server_listen_fd_sharded).
func ServerListenFdSharded(identity unsafe.Pointer, fd uintptr, shardID uint16) (unsafe.Pointer, int) {
	var out *C.nwep_server
	rc := int(C.nwep_server_listen_fd_sharded(&out, (*C.nwep_keypair)(identity), C.uintptr_t(fd), C.uint16_t(shardID)))
	return unsafe.Pointer(out), rc
}

// CidShardID returns which shard a connection id routes to (nwep_cid_shard_id).
func CidShardID(cid []byte) int {
	return int(C.nwep_cid_shard_id(bytePtr(cid), C.size_t(len(cid))))
}

// ReusePortSupported reports whether the platform supports SO_REUSEPORT (nwep_reuse_port_supported).
func ReusePortSupported() bool {
	return C.nwep_reuse_port_supported() != 0
}

// ServerSetOverloaded toggles the front-door shed switch (nwep_server_set_overloaded).
func ServerSetOverloaded(server unsafe.Pointer, on bool) {
	v := C.int(0)
	if on {
		v = 1
	}
	C.nwep_server_set_overloaded((*C.nwep_server)(server), v)
}

// ServerSetMaxParked caps the number of deferred responses outstanding (nwep_server_set_max_parked).
func ServerSetMaxParked(server unsafe.Pointer, maxParked int) {
	C.nwep_server_set_max_parked((*C.nwep_server)(server), C.size_t(maxParked))
}

// ServerLoad returns the current load gauge, 0 to 100 (nwep_server_load).
func ServerLoad(server unsafe.Pointer) int {
	return int(C.nwep_server_load((*C.nwep_server)(server)))
}

// ServerMetricsGet fills a metrics snapshot from the server (nwep_server_metrics_get).
func ServerMetricsGet(server unsafe.Pointer) (ServerMetrics, int) {
	var m C.nwep_server_metrics
	rc := int(C.nwep_server_metrics_get((*C.nwep_server)(server), &m))
	return ServerMetrics{
		ConnectionsActive:   uint64(m.connections_active),
		ConnectionsAccepted: uint64(m.connections_accepted),
		ConnectionsRefused:  uint64(m.connections_refused),
		ConnectionsClosed:   uint64(m.connections_closed),
		BytesReceived:       uint64(m.bytes_received),
		BytesSent:           uint64(m.bytes_sent),
		DatagramsReceived:   uint64(m.datagrams_received),
		DatagramsSent:       uint64(m.datagrams_sent),
		RequestsDispatched:  uint64(m.requests_dispatched),
		RequestsShed:        uint64(m.requests_shed),
		ParkedActive:        uint64(m.parked_active),
		Load:                int32(m.load),
	}, rc
}

// ServerDrain begins a graceful drain, refusing new connections (nwep_server_drain).
func ServerDrain(server unsafe.Pointer) int {
	return int(C.nwep_server_drain((*C.nwep_server)(server)))
}

// ServerIsDrained reports whether the drain has completed (nwep_server_is_drained).
func ServerIsDrained(server unsafe.Pointer) bool {
	return C.nwep_server_is_drained((*C.nwep_server)(server)) != 0
}

// ServerSetHandler registers the request handler, or clears it with a nil fn (nwep_server_set_handler).
func ServerSetHandler(server unsafe.Pointer, fn HandlerFunc) int {
	ud := cbSet(server, fn)
	var cb C.nwep_handler_fn
	if ud != nil {
		cb = serverHandlerCb
	}
	return int(C.nwep_server_set_handler((*C.nwep_server)(server), cb, ud))
}

// ServerTick advances the server state machine at now_ms monotonic (nwep_server_tick).
func ServerTick(server unsafe.Pointer, nowMs int64) int {
	return int(C.nwep_server_tick((*C.nwep_server)(server), C.int64_t(nowMs)))
}

// ServerLocalPort returns the bound udp port (nwep_server_local_port).
func ServerLocalPort(server unsafe.Pointer) uint16 {
	return uint16(C.nwep_server_local_port((*C.nwep_server)(server)))
}

// ServerFd returns the udp socket handle to fold into a reactor (nwep_server_fd).
//
// returned as uintptr to hold a posix int or a windows SOCKET NWG1200.
func ServerFd(server unsafe.Pointer) uintptr {
	return uintptr(C.nwep_server_fd((*C.nwep_server)(server)))
}

// ServerNextTimeoutMs returns ms until the next timer, or negative for none (nwep_server_next_timeout_ms).
func ServerNextTimeoutMs(server unsafe.Pointer, nowMs int64) int {
	return int(C.nwep_server_next_timeout_ms((*C.nwep_server)(server), C.int64_t(nowMs)))
}

// ServerGetPeerNodeid returns the verified node_id of a connection's peer (nwep_server_get_peer_nodeid).
func ServerGetPeerNodeid(server unsafe.Pointer, connID uint64) (nodeID [NodeIDSize]byte, rc int) {
	var id C.nwep_node_id
	rc = int(C.nwep_server_get_peer_nodeid((*C.nwep_server)(server), C.uint64_t(connID), &id))
	if rc == 0 {
		nodeID = *(*[NodeIDSize]byte)(unsafe.Pointer(&id.bytes[0]))
	}
	return
}

// ServerLocalNodeid returns the server's own node_id (nwep_server_local_nodeid).
func ServerLocalNodeid(server unsafe.Pointer) (nodeID [NodeIDSize]byte, rc int) {
	var id C.nwep_node_id
	rc = int(C.nwep_server_local_nodeid((*C.nwep_server)(server), &id))
	if rc == 0 {
		nodeID = *(*[NodeIDSize]byte)(unsafe.Pointer(&id.bytes[0]))
	}
	return
}

// ServerLastHandshakeError returns the last handshake failure code, for diagnostics (nwep_server_last_handshake_error).
func ServerLastHandshakeError(server unsafe.Pointer) int {
	return int(C.nwep_server_last_handshake_error((*C.nwep_server)(server)))
}

// ServerNotify pushes a server-initiated NOTIFY on a connection (nwep_server_notify).
func ServerNotify(server unsafe.Pointer, connID uint64, event string, headers [][2]string, body []byte) int {
	ce := C.CString(event)
	defer C.free(unsafe.Pointer(ce))
	hp, freeHeaders := cHeaders(headers)
	defer freeHeaders()
	return int(C.nwep_server_notify((*C.nwep_server)(server), C.uint64_t(connID), ce, hp, bytePtr(body), C.size_t(len(body))))
}

// ServerBeginStream opens a server-pushed streamed response (nwep_server_begin_stream).
func ServerBeginStream(server unsafe.Pointer, connID, streamID uint64, path, status string, headers [][2]string) int {
	cp := C.CString(path)
	defer C.free(unsafe.Pointer(cp))
	cs := C.CString(status)
	defer C.free(unsafe.Pointer(cs))
	hp, freeHeaders := cHeaders(headers)
	defer freeHeaders()
	return int(C.nwep_server_begin_stream((*C.nwep_server)(server), C.uint64_t(connID), C.uint64_t(streamID), cp, cs, hp))
}

// ServerStreamSend sends a chunk of a streamed response body (nwep_server_stream_send).
func ServerStreamSend(server unsafe.Pointer, connID, streamID uint64, body []byte) int {
	return int(C.nwep_server_stream_send((*C.nwep_server)(server), C.uint64_t(connID), C.uint64_t(streamID), bytePtr(body), C.size_t(len(body))))
}

// ServerStreamEnd finishes a streamed response with a quic fin (nwep_server_stream_end).
func ServerStreamEnd(server unsafe.Pointer, connID, streamID uint64) int {
	return int(C.nwep_server_stream_end((*C.nwep_server)(server), C.uint64_t(connID), C.uint64_t(streamID)))
}

// ServerRespondHeader appends a header to an out-of-band deferred response (nwep_server_respond_header).
func ServerRespondHeader(server unsafe.Pointer, connID, streamID uint64, name, value string) int {
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	cv := C.CString(value)
	defer C.free(unsafe.Pointer(cv))
	return int(C.nwep_server_respond_header((*C.nwep_server)(server), C.uint64_t(connID), C.uint64_t(streamID), cn, cv))
}

// ServerRespond delivers a deferred response out of band after a handler deferred (nwep_server_respond).
func ServerRespond(server unsafe.Pointer, connID, streamID uint64, status string, body []byte) int {
	cs := C.CString(status)
	defer C.free(unsafe.Pointer(cs))
	return int(C.nwep_server_respond((*C.nwep_server)(server), C.uint64_t(connID), C.uint64_t(streamID), cs, bytePtr(body), C.size_t(len(body))))
}

// ServerRelay relays an upstream response back to a deferred request (nwep_server_relay).
func ServerRelay(server unsafe.Pointer, connID, streamID uint64, originResp unsafe.Pointer) int {
	return int(C.nwep_server_relay((*C.nwep_server)(server), C.uint64_t(connID), C.uint64_t(streamID), (*C.nwep_message)(originResp)))
}

// ServerRespondBlit delivers a pre-encoded frame as a deferred response (nwep_server_respond_blit).
func ServerRespondBlit(server unsafe.Pointer, connID, streamID uint64, frame []byte) int {
	return int(C.nwep_server_respond_blit((*C.nwep_server)(server), C.uint64_t(connID), C.uint64_t(streamID), bytePtr(frame), C.size_t(len(frame))))
}

// ServerConnCompression reports whether a connection negotiated body compression (nwep_server_conn_compression).
func ServerConnCompression(server unsafe.Pointer, connID uint64) int {
	return int(C.nwep_server_conn_compression((*C.nwep_server)(server), C.uint64_t(connID)))
}

// ServerClose stops the server, frees it, and retires its handler (nwep_server_close).
func ServerClose(server unsafe.Pointer) {
	cbClear(server)
	C.nwep_server_close((*C.nwep_server)(server))
}
