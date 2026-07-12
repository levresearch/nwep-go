//go:build !nwep_custom_link && !nwep_dev

// the default link configuration, for users who installed libnwep via the nwep
// installer. the installer writes nwep.pc to {prefix}/lib/pkgconfig and pkg-config
// resolves the right -L and -l flags automatically.
//
// linux, system install (/usr/local) — found automatically, just build:
//
//	go build ./...
//
// linux, user install (~/.local) — set PKG_CONFIG_PATH first:
//
//	PKG_CONFIG_PATH=~/.local/lib/pkgconfig go build ./...
//
// windows — pkg-config in MSYS2/mingw64 does not reliably handle the
// Windows-native paths the installer writes into nwep.pc. use nwep_custom_link
// and point CGO_LDFLAGS at the lib dir the installer set in NWEP_LIB_DIR:
//
//	CGO_LDFLAGS="-L$NWEP_LIB_DIR" go build -tags nwep_custom_link ./...
//
// if you are building from the nwep source tree, use the nwep_dev tag instead:
//
//	go build -tags nwep_dev ./...

package sys

// #cgo pkg-config: nwep
import "C"
