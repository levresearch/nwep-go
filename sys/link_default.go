//go:build !nwep_custom_link && !nwep_pkgconfig

// the default link configuration, for native development on the host os.
//
// it dynamically links the libnwep built into the repo's zig-out/lib, with an
// rpath so go test and go run work with no environment setup. for a cross build or
// a self-contained static link against a per-target libnwep-full.a, build with the
// nwep_custom_link tag instead and set the link flags through CGO_LDFLAGS.

package sys

// #cgo LDFLAGS: -L${SRCDIR}/../../../zig-out/lib -lnwep -Wl,-rpath,${SRCDIR}/../../../zig-out/lib
// #cgo linux LDFLAGS: -L/usr/local/lib
import "C"
