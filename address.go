// the address type, an opaque ipv6 socket address, web/1 is ipv6 only NW110900.

package nwep

import "nwep/sys"

// Address is an ipv6 udp socket address a server binds or a client dials.
//
// web/1 is ipv6 only, use IPv4Mapped to reach an ipv4 peer through the
// ::ffff:a.b.c.d mapped form (rfc 4291).
type Address struct {
	raw sys.Address
}

// Loopback returns the ::1 loopback address at port.
func Loopback(port uint16) Address { return Address{sys.AddressLoopback(port)} }

// Wildcard returns the :: wildcard address at port, for binding every interface.
func Wildcard(port uint16) Address { return Address{sys.AddressWildcard(port)} }

// IPv4Mapped returns the ::ffff:a.b.c.d ipv4-mapped address at port.
func IPv4Mapped(a, b, c, d byte, port uint16) Address {
	return Address{sys.AddressIPv4Mapped(a, b, c, d, port)}
}

// AddressFromBytes returns the address for a raw 16-byte ipv6 address at port.
func AddressFromBytes(addr [16]byte, port uint16) Address {
	return Address{sys.AddressFromBytes(addr, port)}
}

// Port returns the udp port held in the address.
func (a Address) Port() uint16 { return sys.AddressGetPort(a.raw) }

// sysAddr exposes the raw sys address for the binding's own internal calls.
func (a Address) sysAddr() sys.Address { return a.raw }
