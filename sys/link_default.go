//go:build !nwep_custom_link && !nwep_dev

// the default link configuration, for users who installed libnwep via the nwep
// installer. the installer writes nwep.pc to {prefix}/lib/pkgconfig and pkg-config
// resolves the right -L and -l flags automatically.
//
// linux — system installs (/usr/local) are found automatically. user installs
// (~/.local) require PKG_CONFIG_PATH:
//
//	PKG_CONFIG_PATH=~/.local/lib/pkgconfig go build ./...
//
// windows — the installer writes nwep.pc to %NWEP_LIB_DIR%\pkgconfig but does
// not update PKG_CONFIG_PATH, so set it before building (use the MSYS2/mingw64
// shell where pkg-config is available):
//
//	PKG_CONFIG_PATH="$NWEP_LIB_DIR/pkgconfig" go build ./...
//
// if you are building from the nwep source tree, use the nwep_dev tag instead:
//
//	go build -tags nwep_dev ./...
//
// for a cross build or a self-contained static link, use nwep_custom_link and
// drive linking entirely through CGO_LDFLAGS.

package sys

// #cgo pkg-config: nwep
import "C"
