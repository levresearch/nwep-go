//go:build nwep_dev && !nwep_custom_link

// the dev link configuration, for building against the nwep source tree.
//
// it dynamically links the libnwep built into the repo's zig-out/lib, with an
// rpath so go test and go run work with no environment setup. run zig build from
// the repo root first, then:
//
//	go build -tags nwep_dev ./...
//	go test -tags nwep_dev ./...

package sys

// #cgo LDFLAGS: -L${SRCDIR}/../../../zig-out/lib -lnwep -Wl,-rpath,${SRCDIR}/../../../zig-out/lib
// #cgo linux LDFLAGS: -L/usr/local/lib
import "C"
