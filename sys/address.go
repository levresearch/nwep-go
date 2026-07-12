// the layer 0 address calls, an opaque ipv6 sockaddr NW110900.

package sys

/*
#include <nwep.h>
*/
import "C"

import "unsafe"

// Address is the raw 32-byte opaque nwep_address storage, never inspected directly.
//
// web/1 is ipv6 only, ipv4 callers embed a v4 address with AddressIPv4Mapped. the
// size matches the header's opaque[32], asserted in the tests.
type Address [32]byte

// c reinterprets the go storage as the c struct, identical layout.
func (a *Address) c() *C.nwep_address { return (*C.nwep_address)(unsafe.Pointer(a)) }

// AddressLoopback returns the ::1 loopback at port (nwep_address_loopback).
func AddressLoopback(port uint16) Address {
	var a Address
	C.nwep_address_loopback(a.c(), C.uint16_t(port))
	return a
}

// AddressWildcard returns the :: wildcard at port, for binding any interface (nwep_address_wildcard).
func AddressWildcard(port uint16) Address {
	var a Address
	C.nwep_address_wildcard(a.c(), C.uint16_t(port))
	return a
}

// AddressIPv4Mapped returns the ::ffff:a.b.c.d ipv4-mapped form at port (nwep_address_ipv4_mapped).
func AddressIPv4Mapped(a, b, c, d byte, port uint16) Address {
	var out Address
	C.nwep_address_ipv4_mapped(out.c(), C.uint8_t(a), C.uint8_t(b), C.uint8_t(c), C.uint8_t(d), C.uint16_t(port))
	return out
}

// AddressFromBytes returns the address for a raw 16-byte ipv6 address at port (nwep_address_from_bytes).
func AddressFromBytes(addr [16]byte, port uint16) Address {
	var out Address
	C.nwep_address_from_bytes(out.c(), (*C.uint8_t)(unsafe.Pointer(&addr[0])), C.uint16_t(port))
	return out
}

// AddressGetPort returns the udp port held in an address (nwep_address_get_port).
func AddressGetPort(a Address) uint16 {
	return uint16(C.nwep_address_get_port(a.c()))
}
