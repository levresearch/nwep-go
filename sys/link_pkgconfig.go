//go:build !nwep_custom_link && nwep_pkgconfig

// the pkg-config link configuration, for Linux installs where the nwep installer
// wrote nwep.pc. build with this tag when PKG_CONFIG_PATH includes the pkgconfig
// directory of your nwep install (e.g. ~/.local/lib/pkgconfig for a user install):
//
//   PKG_CONFIG_PATH=~/.local/lib/pkgconfig go build -tags nwep_pkgconfig ./...
//
// for a system install (/usr/local/lib) pkg-config is found automatically. run
// `go run nwep/cmd/checklib` to see which command applies to your setup.

package sys

// #cgo pkg-config: nwep
import "C"
