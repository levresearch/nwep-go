#!/usr/bin/env bash
# cross-compiles the go binding test binary for each target and runs it under
# emulation where that is reliable (guide 12). zig is the cross c compiler, and the
# per-target self-contained libnwep-full.a is the link input. run from bindings/go.
#
# the archives are expected under DIST (the published per-target libs). override
# DIST to point at wherever the core's per-target -full.a archives live.
set -u

DIST="${DIST:-$(cd "$(dirname "$0")/../../.." && pwd)/distrib/public/libs/0.1.0}"
GOPKG="./"
RUN='TestIdentityVertical|TestLoopbackRequestResponse|TestManagedServe|TestShamir|TestBLSRoundTrip'

# target rows: name goarch-env  zig-target  dist-dir  docker-platform (empty = no run)
targets=(
  "linux-aarch64|GOOS=linux GOARCH=arm64|aarch64-linux-gnu|linux-aarch64|linux/arm64"
  "linux-x86|GOOS=linux GOARCH=386|x86-linux-gnu|linux-x86|linux/386"
  "linux-armv7|GOOS=linux GOARCH=arm GOARM=7|arm-linux-gnueabihf|linux-armv7|linux/arm/v7"
)

pass=0 fail=0
for row in "${targets[@]}"; do
  IFS='|' read -r name goenv zigtarget distdir platform <<<"$row"
  archive="$DIST/$distdir/libnwep-full.a"
  out="/tmp/nwep-$name.test"
  echo "=== $name ==="
  if [ ! -f "$archive" ]; then echo "  skip, no archive at $archive"; continue; fi

  if env $goenv CGO_ENABLED=1 \
      CC="zig cc -target $zigtarget" CXX="zig c++ -target $zigtarget" \
      CGO_LDFLAGS="$archive -lc++ -lm -lpthread -ldl" \
      go test -tags nwep_custom_link -c -o "$out" "$GOPKG" 2>/tmp/nwep-$name.build.log; then
    echo "  compile ok"
  else
    echo "  COMPILE FAILED, see /tmp/nwep-$name.build.log"; fail=$((fail+1)); continue
  fi

  if [ -n "$platform" ] && command -v docker >/dev/null 2>&1; then
    img="ubuntu:22.04"; [ "$platform" = "linux/386" ] && img="i386/debian:12"
    if docker run --rm --platform "$platform" -v "$out:/t:ro" "$img" /t -test.run "$RUN" >/tmp/nwep-$name.run.log 2>&1; then
      echo "  run ok (emulated)"; pass=$((pass+1))
    else
      echo "  RUN FAILED (emulation may be unreliable for this arch), see /tmp/nwep-$name.run.log"; fail=$((fail+1))
    fi
  else
    echo "  run skipped (no emulation configured)"
  fi
done

echo
echo "verified-run: $pass, failed: $fail"
