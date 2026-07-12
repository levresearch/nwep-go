// the layer 0 response builders, writing an encoded response into a buf NW060000.

package sys

/*
#include <stdlib.h>
#include <nwep.h>
*/
import "C"

import "unsafe"

// Range is a byte range request, a half-open [Start, End) over a resource NW060800.
type Range struct {
	Start, End uint64
}

// cRanges copies a go range slice into a c array, returning its pointer and count.
func cRanges(ranges []Range) (*C.nwep_range, C.size_t) {
	if len(ranges) == 0 {
		return nil, 0
	}
	out := make([]C.nwep_range, len(ranges))
	for i, r := range ranges {
		out[i].start = C.uint64_t(r.Start)
		out[i].end = C.uint64_t(r.End)
	}
	return &out[0], C.size_t(len(ranges))
}

// ResponseOk writes a 200 ok response with body into buf (nwep_response_ok).
func ResponseOk(buf unsafe.Pointer, body []byte) int {
	return int(C.nwep_response_ok((*C.nwep_buf)(buf), bytePtr(body), C.size_t(len(body))))
}

// ResponseStatus writes a response with an explicit status and body into buf (nwep_response_status).
func ResponseStatus(buf unsafe.Pointer, status string, body []byte) int {
	cs := C.CString(status)
	defer C.free(unsafe.Pointer(cs))
	return int(C.nwep_response_status((*C.nwep_buf)(buf), cs, bytePtr(body), C.size_t(len(body))))
}

// ResponseNotModified writes a 304 not-modified carrying etag into buf (nwep_response_not_modified).
func ResponseNotModified(buf unsafe.Pointer, etag string) int {
	cs := C.CString(etag)
	defer C.free(unsafe.Pointer(cs))
	return int(C.nwep_response_not_modified((*C.nwep_buf)(buf), cs))
}

// ResponsePartial writes a 206 partial response for the given ranges into buf (nwep_response_partial).
func ResponsePartial(buf unsafe.Pointer, body []byte, ranges []Range, contentType string) int {
	rp, n := cRanges(ranges)
	cs := C.CString(contentType)
	defer C.free(unsafe.Pointer(cs))
	return int(C.nwep_response_partial((*C.nwep_buf)(buf), bytePtr(body), C.size_t(len(body)), rp, n, cs))
}

// ResponseRangeNotSatisfiable writes a 416 with the resource length into buf (nwep_response_range_not_satisfiable).
func ResponseRangeNotSatisfiable(buf unsafe.Pointer, totalLen uint64) int {
	return int(C.nwep_response_range_not_satisfiable((*C.nwep_buf)(buf), C.uint64_t(totalLen)))
}

// ResponseHeader appends one header to a response being built in buf (nwep_response_header).
func ResponseHeader(buf unsafe.Pointer, name, value string) int {
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	cv := C.CString(value)
	defer C.free(unsafe.Pointer(cv))
	return int(C.nwep_response_header((*C.nwep_buf)(buf), cn, cv))
}

// ResponseRelay writes an upstream response back out verbatim into buf (nwep_response_relay).
func ResponseRelay(buf unsafe.Pointer, origin unsafe.Pointer) int {
	return int(C.nwep_response_relay((*C.nwep_buf)(buf), (*C.nwep_message)(origin)))
}

// ResponseCapture copies the encoded response frame out of buf (nwep_response_capture).
//
// returns the captured frame bytes, sized by a first probe call, and the code.
func ResponseCapture(buf unsafe.Pointer) ([]byte, int) {
	var outlen C.size_t
	rc := int(C.nwep_response_capture((*C.nwep_buf)(buf), nil, 0, &outlen))
	if rc != 0 {
		return nil, rc
	}
	out := make([]byte, outlen)
	rc = int(C.nwep_response_capture((*C.nwep_buf)(buf), bytePtr(out), outlen, &outlen))
	if rc != 0 {
		return nil, rc
	}
	return out[:outlen], rc
}

// ResponseBlit writes a pre-encoded response frame straight into buf (nwep_response_blit).
func ResponseBlit(buf unsafe.Pointer, frame []byte) int {
	return int(C.nwep_response_blit((*C.nwep_buf)(buf), bytePtr(frame), C.size_t(len(frame))))
}

// ResponseVerify checks a response signature against pubkey for path at now (nwep_response_verify).
func ResponseVerify(resp unsafe.Pointer, pubkey [PubKeySize]byte, path string, nowSecs uint64) int {
	cp := C.CString(path)
	defer C.free(unsafe.Pointer(cp))
	return int(C.nwep_response_verify((*C.nwep_message)(resp), (*C.uint8_t)(unsafe.Pointer(&pubkey[0])), cp, C.uint64_t(nowSecs)))
}
