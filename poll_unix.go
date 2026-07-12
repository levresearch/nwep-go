//go:build unix

// the readability wait for the managed loop on unix, a select on the socket fd.

package nwep

import (
	"syscall"
	"time"
	"unsafe"
)

// waitReadable waits up to timeout for fd to become readable, best effort NWG1200.
//
// a wakeup, a timeout, or an interrupted select all simply return, the next tick
// reads whatever is ready. this is the portable poll primitive the managed loop
// uses (select), no eventfd or epoll, so it works on every unix target.
func waitReadable(fd uintptr, timeout time.Duration) {
	if timeout <= 0 {
		return
	}
	var set syscall.FdSet
	// fd_set is a little-endian bitmap, set bit fd through a byte view so this
	// stays correct whatever width the FdSet.Bits field has on this arch NWG1200.
	bytes := (*[unsafe.Sizeof(set)]byte)(unsafe.Pointer(&set))
	bytes[fd/8] |= 1 << (fd % 8)
	tv := syscall.NsecToTimeval(int64(timeout))
	// errors (including EINTR or a racing close) are ignored, the loop re-ticks.
	_, _ = syscall.Select(int(fd)+1, &set, nil, nil, &tv)
}
