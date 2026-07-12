// Package sys is layer 0 of the nwep go binding, the raw cgo surface NWG0200.
//
// it is a 1:1, unsafe, total translation of nwep.h and nwep_trust.h. every
// exported c symbol gets a thin wrapper here and nothing more, so the safe
// layers above can never become a feature ceiling NWG1000. c struct types
// cannot cross a cgo package boundary, so opaque handles are returned as
// unsafe.Pointer and the fixed value types as go arrays. secret key material is
// kept in c memory and never copied into the go heap so it can be wiped with
// Zeroize NWG0700. nobody is meant to enjoy this layer, reach for the nwep
// package instead.
package sys

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <stdlib.h>
#include <nwep.h>
*/
import "C"

import "unsafe"

// the fixed byte sizes of the value types, taken from the header NW040200.
const (
	NodeIDSize  = C.NWEP_NODEID_SIZE
	PubKeySize  = C.NWEP_PUBKEY_SIZE
	PrivKeySize = C.NWEP_PRIVKEY_SIZE
)

// Version returns the library version string, for example 0.1.0 (nwep_version).
func Version() string {
	return C.GoString(C.nwep_version())
}

// Strerror returns the human message for an nwep error code (nwep_strerror).
func Strerror(code int) string {
	return C.GoString(C.nwep_strerror(C.int(code)))
}

// Zeroize wipes length bytes at ptr, for clearing secret material (nwep_zeroize).
//
// # safety
//
// ptr must point at length valid writable bytes. used by the safe layer to wipe
// c-owned secrets before free, callers below it should not need this directly.
func Zeroize(ptr unsafe.Pointer, length int) {
	C.nwep_zeroize(ptr, C.size_t(length))
}
