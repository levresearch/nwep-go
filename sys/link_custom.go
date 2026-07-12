//go:build nwep_custom_link

// the custom link configuration, for cross builds and self-contained static links.
//
// it adds no link flags of its own, so the caller controls linking entirely
// through CGO_LDFLAGS. point it at the per-target libnwep-full.a (the
// self-contained archive the core emits for every target) plus the c++ runtime and
// the system libraries, for example:
//
//	CGO_LDFLAGS="/path/linux-aarch64/libnwep-full.a -lc++ -lm -lpthread -ldl" \
//	CC="zig cc -target aarch64-linux-gnu" GOOS=linux GOARCH=arm64 \
//	CGO_ENABLED=1 go build -tags nwep_custom_link ./...

package sys
