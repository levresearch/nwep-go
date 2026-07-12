// the layer 0 callback bridge, handing c real function pointers that forward to go.
//
// the library invokes a handler, a request-done hook, and a log-append hook. cgo
// cannot pass a go function as a c function pointer, so each is a static c shim
// (below) that forwards to an exported go trampoline. the go closure is found
// through a runtime/cgo.Handle carried in the userdata void pointer, and every
// trampoline recovers panics so nothing unwinds across the c frame NWG0900.

package sys

/*
#include <stdlib.h>
#include <nwep.h>

// the shim bodies live in shims.c so they are compiled exactly once. cgo copies
// this preamble into every generated translation unit, so a definition here would
// multiply-define, and a static definition's address would not survive the link.
// these extern declarations let go take each shim's address (the c function
// pointer handed to the library, never a go function pointer NWG0900).
extern int nwepServerHandlerShim(nwep_server*, uint64_t, uint64_t, const nwep_message*, nwep_buf*, void*);
extern void nwepRequestDoneShim(nwep_client*, uint64_t, int, nwep_message*, void*);
extern void nwepLogAppendShim(void*, const uint8_t*, size_t, uint64_t);
*/
import "C"

import (
	"runtime/cgo"
	"sync"
	"unsafe"
)

// HandlerFunc is the go form of nwep_handler_fn, run synchronously inside a tick.
//
// it returns 0 for a synchronous answer (after writing buf), a negative code for
// a generic error response, or DeferSentinel to answer out of band later.
type HandlerFunc func(server unsafe.Pointer, connID, streamID uint64, req, buf unsafe.Pointer) int

// RequestDoneFunc is the go form of nwep_request_done_fn, one per client.
//
// on status 0 resp owns a message the callback must free with MessageFree, on a
// negative status resp is nil.
type RequestDoneFunc func(client unsafe.Pointer, id uint64, status int, resp unsafe.Pointer)

// LogAppendFunc is the go form of nwep_log_append_fn, the accepted-entry hook.
//
// entry is copied into an owned slice before the call, so it stays valid.
type LogAppendFunc func(entry []byte, index uint64)

// DeferSentinel is NWEP_DEFER, the handler return that answers out of band later.
const DeferSentinel = int(C.NWEP_DEFER)

// the c callback function pointers, computed in this file because the shims are
// only declared in this file's cgo preamble. other files in the package hand
// these to the library (a c function pointer, never a go one NWG0900).
var (
	serverHandlerCb = C.nwep_handler_fn(C.nwepServerHandlerShim)
	requestDoneCb   = C.nwep_request_done_fn(C.nwepRequestDoneShim)
	logAppendCb     = C.nwep_log_append_fn(C.nwepLogAppendShim)
)

// cbEntry is one owner's live callback, a cgo.Handle pinning the go closure and
// the c cell that carries the handle to the library as a void pointer.
type cbEntry struct {
	handle cgo.Handle
	cell   unsafe.Pointer
}

// the registry pins each owner's active callback so the closure survives until the
// owner is closed or the callback is replaced.
var (
	cbMu  sync.Mutex
	cbReg = make(map[unsafe.Pointer]cbEntry)
)

// cbSet stores fn for owner, retires any previous callback, and returns the
// userdata void pointer to hand the library (nil when fn is nil, to clear).
//
// the handle is carried to c through a heap cell holding its value, rather than by
// casting the handle integer to a pointer, so the conversion is a real c pointer
// the runtime pointer checker accepts.
func cbSet(owner unsafe.Pointer, fn any) unsafe.Pointer {
	cbMu.Lock()
	defer cbMu.Unlock()
	if old, ok := cbReg[owner]; ok {
		old.handle.Delete()
		C.free(old.cell)
		delete(cbReg, owner)
	}
	if fn == nil {
		return nil
	}
	h := cgo.NewHandle(fn)
	cell := C.malloc(C.size_t(unsafe.Sizeof(h)))
	*(*cgo.Handle)(cell) = h
	cbReg[owner] = cbEntry{handle: h, cell: cell}
	return cell
}

// cbClear retires owner's callback, called when the owner is freed.
func cbClear(owner unsafe.Pointer) {
	cbMu.Lock()
	defer cbMu.Unlock()
	if e, ok := cbReg[owner]; ok {
		e.handle.Delete()
		C.free(e.cell)
		delete(cbReg, owner)
	}
}

// cbValue recovers the go closure a userdata void pointer (a handle cell) points at.
func cbValue(ud unsafe.Pointer) any {
	if ud == nil {
		return nil
	}
	return (*(*cgo.Handle)(ud)).Value()
}

//export nwepGoServerHandler
func nwepGoServerHandler(server *C.nwep_server, connID, streamID C.uint64_t, req *C.nwep_message, buf *C.nwep_buf, ud unsafe.Pointer) (rc C.int) {
	fn, _ := cbValue(ud).(HandlerFunc)
	if fn == nil {
		return C.int(ErrInternal)
	}
	// recover so a panicking handler becomes a generic error, never unwinds into c.
	defer func() {
		if recover() != nil {
			rc = C.int(ErrInternal)
		}
	}()
	return C.int(fn(unsafe.Pointer(server), uint64(connID), uint64(streamID), unsafe.Pointer(req), unsafe.Pointer(buf)))
}

//export nwepGoRequestDone
func nwepGoRequestDone(client *C.nwep_client, id C.uint64_t, status C.int, resp *C.nwep_message, ud unsafe.Pointer) {
	fn, _ := cbValue(ud).(RequestDoneFunc)
	if fn == nil {
		return
	}
	defer func() { _ = recover() }()
	fn(unsafe.Pointer(client), uint64(id), int(status), unsafe.Pointer(resp))
}

//export nwepGoLogAppend
func nwepGoLogAppend(ctx unsafe.Pointer, entry *C.uint8_t, length C.size_t, index C.uint64_t) {
	fn, _ := cbValue(ctx).(LogAppendFunc)
	if fn == nil {
		return
	}
	defer func() { _ = recover() }()
	var b []byte
	if entry != nil && length > 0 {
		b = C.GoBytes(unsafe.Pointer(entry), C.int(length))
	}
	fn(b, uint64(index))
}
