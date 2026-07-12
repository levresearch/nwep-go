// the layer 0 shamir secret sharing, splitting a recovery key into shares NW150400.

package sys

/*
#include <nwep.h>
*/
import "C"

import "unsafe"

// bytePtr returns a c pointer to the first byte of b, or nil for an empty slice.
func bytePtr(b []byte) *C.uint8_t {
	if len(b) == 0 {
		return nil
	}
	return (*C.uint8_t)(unsafe.Pointer(&b[0]))
}

// ShamirSplit splits secret into n shares, any t of which recombine it (nwep_shamir_split).
//
// each share is 1 + len(secret) bytes (a 1-based index byte then data), so the
// output size is exact and needs no sizing probe. returns the concatenated share
// bytes and the c return code.
func ShamirSplit(secret []byte, t, n int) ([]byte, int) {
	shareLen := 1 + len(secret)
	out := make([]byte, n*shareLen)
	outlen := C.size_t(len(out))
	rc := int(C.nwep_shamir_split(bytePtr(secret), C.size_t(len(secret)), C.size_t(t), C.size_t(n), bytePtr(out), &outlen))
	if rc != 0 {
		return nil, rc
	}
	return out[:outlen], rc
}

// ShamirCombine recombines n_shares concatenated shares back into the secret (nwep_shamir_combine).
//
// the secret is shareLen-1 bytes, so a shareLen buffer is always ample.
func ShamirCombine(shares []byte, nShares, shareLen int) ([]byte, int) {
	out := make([]byte, shareLen)
	outlen := C.size_t(shareLen)
	rc := int(C.nwep_shamir_combine(bytePtr(shares), C.size_t(nShares), C.size_t(shareLen), bytePtr(out), &outlen))
	if rc != 0 {
		return nil, rc
	}
	return out[:outlen], rc
}
