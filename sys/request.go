// the layer 0 request helpers, conditional freshness and range parsing NW060700 NW060800.

package sys

/*
#include <stdlib.h>
#include <nwep.h>
*/
import "C"

import "unsafe"

// cStrOrNil returns a freshly allocated c string for s, or nil for the empty string.
//
// the caller must free a non-nil result. an empty string maps to nil so an absent
// optional argument reads as a c null rather than a zero-length string.
func cStrOrNil(s string) *C.char {
	if s == "" {
		return nil
	}
	return C.CString(s)
}

// RequestIsFresh reports whether a conditional request still matches etag (nwep_request_is_fresh).
//
// returns 1 when the cached representation is fresh (a 304 may be sent), 0 when
// stale, or a negative code.
func RequestIsFresh(req unsafe.Pointer, etag string) int {
	ce := cStrOrNil(etag)
	if ce != nil {
		defer C.free(unsafe.Pointer(ce))
	}
	return int(C.nwep_request_is_fresh((*C.nwep_message)(req), ce))
}

// RequestRange parses the request range header against totalLen (nwep_request_range).
//
// returns the satisfiable ranges (up to maxOut) and the c return code.
func RequestRange(req unsafe.Pointer, totalLen uint64, etag string, maxOut int) ([]Range, int) {
	ce := cStrOrNil(etag)
	if ce != nil {
		defer C.free(unsafe.Pointer(ce))
	}
	out := make([]C.nwep_range, maxOut)
	var op *C.nwep_range
	if maxOut > 0 {
		op = &out[0]
	}
	var count C.size_t
	rc := int(C.nwep_request_range((*C.nwep_message)(req), C.uint64_t(totalLen), ce, op, C.size_t(maxOut), &count))
	if rc != 0 {
		return nil, rc
	}
	res := make([]Range, count)
	for i := 0; i < int(count); i++ {
		res[i] = Range{Start: uint64(out[i].start), End: uint64(out[i].end)}
	}
	return res, rc
}
