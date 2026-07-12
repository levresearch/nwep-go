//go:build !unix

// the readability wait for the managed loop on non-unix targets (windows), a
// bounded sleep. it degrades gracefully rather than breaking NWG1200, the
// managed loop caps the wait so inbound packets are still serviced promptly, and
// the driven layer remains the fully portable floor for callers who want a real
// reactor with WSAPoll.

package nwep

import "time"

// waitReadable waits up to timeout before the next tick, the degraded form.
func waitReadable(_ uintptr, timeout time.Duration) {
	if timeout > 0 {
		time.Sleep(timeout)
	}
}
