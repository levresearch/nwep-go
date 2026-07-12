// Package nwep is the go binding for libnwep, the web/1 protocol over quic.
//
// this package is the safe api, both the driven layer (you own the event loop,
// folding Fd, Tick, and NextTimeout into your own reactor) and the managed layer
// (a Runtime owns a goroutine that runs the loop for you). the raw cgo surface is
// the nwep/sys package, reachable from any handle here through a Raw method when
// you need an operation the safe api does not wrap (the no-cliffs rule, NWG0200).
//
// linking libnwep (the complete build) brings the trust layer with it. clients
// that do not need anchor verification can link libnwep_core instead.
package nwep
