// the layer 0 client calls, the connecting half of a node NW070000.

package sys

/*
#include <stdlib.h>
#include <nwep.h>
*/
import "C"

import "unsafe"

// ClientMetrics mirrors nwep_client_metrics, a pull-model snapshot of one connection.
type ClientMetrics struct {
	RequestsInflight  uint64
	RequestsCompleted uint64
	RequestsFailed    uint64
	SmoothedRTTus     uint64
	Alive             int32
}

// nodeIDPtr returns a c pointer to a temporary copy of nodeID for a borrowed arg.
func nodeIDPtr(nodeID [NodeIDSize]byte) *C.nwep_node_id {
	// the copy lives in the caller frame for the duration of the c call.
	id := nodeIDToC(nodeID)
	return &id
}

// ClientConnect dials target_addr and runs the web/1 handshake, blocking (nwep_client_connect).
func ClientConnect(identity unsafe.Pointer, target [NodeIDSize]byte, addr Address) (unsafe.Pointer, int) {
	var out *C.nwep_client
	id := nodeIDToC(target)
	rc := int(C.nwep_client_connect(&out, (*C.nwep_keypair)(identity), &id, addr.c()))
	return unsafe.Pointer(out), rc
}

// ClientConnectFd dials over a caller-owned udp socket, blocking (nwep_client_connect_fd).
func ClientConnectFd(identity unsafe.Pointer, target [NodeIDSize]byte, addr Address, fd uintptr) (unsafe.Pointer, int) {
	var out *C.nwep_client
	id := nodeIDToC(target)
	rc := int(C.nwep_client_connect_fd(&out, (*C.nwep_keypair)(identity), &id, addr.c(), C.uintptr_t(fd)))
	return unsafe.Pointer(out), rc
}

// ClientConnectAsync starts a non-blocking connect, driven with ClientConnectPoll (nwep_client_connect_async).
func ClientConnectAsync(identity unsafe.Pointer, target [NodeIDSize]byte, addr Address) (unsafe.Pointer, int) {
	var out *C.nwep_client
	id := nodeIDToC(target)
	rc := int(C.nwep_client_connect_async(&out, (*C.nwep_keypair)(identity), &id, addr.c()))
	return unsafe.Pointer(out), rc
}

// ClientConnectFdAsync starts a non-blocking connect over a caller socket (nwep_client_connect_fd_async).
func ClientConnectFdAsync(identity unsafe.Pointer, target [NodeIDSize]byte, addr Address, fd uintptr) (unsafe.Pointer, int) {
	var out *C.nwep_client
	id := nodeIDToC(target)
	rc := int(C.nwep_client_connect_fd_async(&out, (*C.nwep_keypair)(identity), &id, addr.c(), C.uintptr_t(fd)))
	return unsafe.Pointer(out), rc
}

// ClientConnectPoll advances an async connect, returning would-block until ready (nwep_client_connect_poll).
func ClientConnectPoll(client unsafe.Pointer) int {
	return int(C.nwep_client_connect_poll((*C.nwep_client)(client)))
}

// ClientConnectByNodeid resolves the target through the dht then connects (nwep_client_connect_by_nodeid).
func ClientConnectByNodeid(identity unsafe.Pointer, target [NodeIDSize]byte, dht unsafe.Pointer, lookupTimeoutMs uint32) (unsafe.Pointer, int) {
	var out *C.nwep_client
	id := nodeIDToC(target)
	rc := int(C.nwep_client_connect_by_nodeid(&out, (*C.nwep_keypair)(identity), &id, (*C.nwep_dht)(dht), C.uint32_t(lookupTimeoutMs)))
	return unsafe.Pointer(out), rc
}

// ClientSend sends a request and blocks for the response (nwep_client_send).
func ClientSend(client unsafe.Pointer, method int, path string, headers [][2]string, body []byte) (unsafe.Pointer, int) {
	cp := C.CString(path)
	defer C.free(unsafe.Pointer(cp))
	hp, freeHeaders := cHeaders(headers)
	defer freeHeaders()
	var out *C.nwep_message
	rc := int(C.nwep_client_send((*C.nwep_client)(client), C.int(method), cp, hp, bytePtr(body), C.size_t(len(body)), &out))
	return unsafe.Pointer(out), rc
}

// ClientFd returns the client's udp socket handle (nwep_client_fd).
func ClientFd(client unsafe.Pointer) uintptr {
	return uintptr(C.nwep_client_fd((*C.nwep_client)(client)))
}

// ClientTick advances the client state machine at now_ms monotonic (nwep_client_tick).
func ClientTick(client unsafe.Pointer, nowMs int64) int {
	return int(C.nwep_client_tick((*C.nwep_client)(client), C.int64_t(nowMs)))
}

// ClientNextTimeoutMs returns ms until the next timer, or negative for none (nwep_client_next_timeout_ms).
func ClientNextTimeoutMs(client unsafe.Pointer, nowMs int64) int {
	return int(C.nwep_client_next_timeout_ms((*C.nwep_client)(client), C.int64_t(nowMs)))
}

// ClientIsAlive reports whether the connection is still usable (nwep_client_is_alive).
func ClientIsAlive(client unsafe.Pointer) bool {
	return C.nwep_client_is_alive((*C.nwep_client)(client)) != 0
}

// ClientMetricsGet fills a metrics snapshot from the client (nwep_client_metrics_get).
func ClientMetricsGet(client unsafe.Pointer) (ClientMetrics, int) {
	var m C.nwep_client_metrics
	rc := int(C.nwep_client_metrics_get((*C.nwep_client)(client), &m))
	return ClientMetrics{
		RequestsInflight:  uint64(m.requests_inflight),
		RequestsCompleted: uint64(m.requests_completed),
		RequestsFailed:    uint64(m.requests_failed),
		SmoothedRTTus:     uint64(m.smoothed_rtt_us),
		Alive:             int32(m.alive),
	}, rc
}

// ClientRequestSubmit queues a non-blocking request, returning its id (nwep_client_request_submit).
func ClientRequestSubmit(client unsafe.Pointer, method int, path string, headers [][2]string, body []byte) (uint64, int) {
	cp := C.CString(path)
	defer C.free(unsafe.Pointer(cp))
	hp, freeHeaders := cHeaders(headers)
	defer freeHeaders()
	var id C.nwep_request_id
	rc := int(C.nwep_client_request_submit((*C.nwep_client)(client), C.int(method), cp, hp, bytePtr(body), C.size_t(len(body)), &id))
	return uint64(id), rc
}

// ClientRequestPoll checks a submitted request, returning the response when ready (nwep_client_request_poll).
func ClientRequestPoll(client unsafe.Pointer, id uint64) (unsafe.Pointer, int) {
	var out *C.nwep_message
	rc := int(C.nwep_client_request_poll((*C.nwep_client)(client), C.nwep_request_id(id), &out))
	return unsafe.Pointer(out), rc
}

// ClientRequestCancel abandons a submitted request by id (nwep_client_request_cancel).
func ClientRequestCancel(client unsafe.Pointer, id uint64) {
	C.nwep_client_request_cancel((*C.nwep_client)(client), C.nwep_request_id(id))
}

// ClientSetRequestDone registers the per-client request completion hook, nil to clear (nwep_client_set_request_done).
func ClientSetRequestDone(client unsafe.Pointer, fn RequestDoneFunc) int {
	ud := cbSet(client, fn)
	var cb C.nwep_request_done_fn
	if ud != nil {
		cb = requestDoneCb
	}
	return int(C.nwep_client_set_request_done((*C.nwep_client)(client), cb, ud))
}

// ClientOpenStream opens a streamed read, returning the stream id (nwep_client_open_stream).
func ClientOpenStream(client unsafe.Pointer, method int, path string, headers [][2]string) (uint64, int) {
	cp := C.CString(path)
	defer C.free(unsafe.Pointer(cp))
	hp, freeHeaders := cHeaders(headers)
	defer freeHeaders()
	var sid C.uint64_t
	rc := int(C.nwep_client_open_stream((*C.nwep_client)(client), C.int(method), cp, hp, &sid))
	return uint64(sid), rc
}

// ClientStreamResponse fetches the response headers of an opened stream (nwep_client_stream_response).
func ClientStreamResponse(client unsafe.Pointer, streamID uint64) (unsafe.Pointer, int) {
	var out *C.nwep_message
	rc := int(C.nwep_client_stream_response((*C.nwep_client)(client), C.uint64_t(streamID), &out))
	return unsafe.Pointer(out), rc
}

// ClientStreamRecv reads the next chunk of a stream body into buf (nwep_client_stream_recv).
//
// returns the bytes read, whether the stream ended (quic fin), and the c code.
func ClientStreamRecv(client unsafe.Pointer, streamID uint64, buf []byte) (n int, ended bool, rc int) {
	var outLen C.size_t
	var outEnded C.int
	rc = int(C.nwep_client_stream_recv((*C.nwep_client)(client), C.uint64_t(streamID), bytePtr(buf), C.size_t(len(buf)), &outLen, &outEnded))
	return int(outLen), outEnded != 0, rc
}

// ClientStreamVerify checks the running signature over a stream against pubkey (nwep_client_stream_verify).
func ClientStreamVerify(client unsafe.Pointer, streamID uint64, pubkey [PubKeySize]byte) int {
	return int(C.nwep_client_stream_verify((*C.nwep_client)(client), C.uint64_t(streamID), (*C.uint8_t)(unsafe.Pointer(&pubkey[0]))))
}

// ClientStreamClose releases a stream's client-side state (nwep_client_stream_close).
func ClientStreamClose(client unsafe.Pointer, streamID uint64) {
	C.nwep_client_stream_close((*C.nwep_client)(client), C.uint64_t(streamID))
}

// ClientSetCache attaches a response cache to the client, nil to detach (nwep_client_set_cache).
func ClientSetCache(client unsafe.Pointer, cache unsafe.Pointer) int {
	return int(C.nwep_client_set_cache((*C.nwep_client)(client), (*C.nwep_cache)(cache)))
}

// ClientCompression reports whether the connection negotiated body compression (nwep_client_compression).
func ClientCompression(client unsafe.Pointer) int {
	return int(C.nwep_client_compression((*C.nwep_client)(client)))
}

// ClientPeerPubkey copies out the verified ed25519 public key of the peer (nwep_client_peer_pubkey).
func ClientPeerPubkey(client unsafe.Pointer) (pubkey [PubKeySize]byte, rc int) {
	rc = int(C.nwep_client_peer_pubkey((*C.nwep_client)(client), (*C.uint8_t)(unsafe.Pointer(&pubkey[0]))))
	return
}

// ClientVerifyResponse checks a response signature against the peer for path (nwep_client_verify_response).
func ClientVerifyResponse(client unsafe.Pointer, resp unsafe.Pointer, path string, nowSecs uint64) int {
	cp := C.CString(path)
	defer C.free(unsafe.Pointer(cp))
	return int(C.nwep_client_verify_response((*C.nwep_client)(client), (*C.nwep_message)(resp), cp, C.uint64_t(nowSecs)))
}

// ClientPollNotify returns the next buffered server NOTIFY, or nil (nwep_client_poll_notify).
func ClientPollNotify(client unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.nwep_client_poll_notify((*C.nwep_client)(client)))
}

// ClientClose tears down the connection, frees it, and retires its hooks (nwep_client_close).
func ClientClose(client unsafe.Pointer) {
	cbClear(client)
	C.nwep_client_close((*C.nwep_client)(client))
}
