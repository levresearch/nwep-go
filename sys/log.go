// the layer 0 merkle log and log server, the transparency log half NW120014.

package sys

/*
#include <stdlib.h>
#include <nwep.h>
*/
import "C"

import "unsafe"

// LogCreate allocates an empty append-only merkle log (nwep_log_create).
func LogCreate() unsafe.Pointer {
	return unsafe.Pointer(C.nwep_log_create())
}

// LogFree frees a log (nwep_log_free).
func LogFree(log unsafe.Pointer) {
	C.nwep_log_free((*C.nwep_log)(log))
}

// LogAppend appends an entry and returns its index, or a negative code (nwep_log_append).
func LogAppend(log unsafe.Pointer, bytes []byte) int64 {
	return int64(C.nwep_log_append((*C.nwep_log)(log), bytePtr(bytes), C.size_t(len(bytes))))
}

// LogSize returns the number of entries in the log (nwep_log_size).
func LogSize(log unsafe.Pointer) uint64 {
	return uint64(C.nwep_log_size((*C.nwep_log)(log)))
}

// LogRoot copies out the current merkle root, rfc-6962 (nwep_log_root).
func LogRoot(log unsafe.Pointer) (root [32]byte, rc int) {
	rc = int(C.nwep_log_root((*C.nwep_log)(log), (*C.uint8_t)(unsafe.Pointer(&root[0]))))
	return
}

// LogServerCreate wraps a log behind a server that answers log queries (nwep_log_server_create).
func LogServerCreate(identity unsafe.Pointer, log unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(C.nwep_log_server_create((*C.nwep_keypair)(identity), (*C.nwep_log)(log)))
}

// LogServerFree frees a log server and retires its on-append hook (nwep_log_server_free).
func LogServerFree(ls unsafe.Pointer) {
	cbClear(ls)
	C.nwep_log_server_free((*C.nwep_log_server)(ls))
}

// LogServerSetOnAppend registers the accepted-entry persistence hook, nil to clear (nwep_log_server_set_on_append).
func LogServerSetOnAppend(ls unsafe.Pointer, fn LogAppendFunc) {
	ctx := cbSet(ls, fn)
	var cb C.nwep_log_append_fn
	if ctx != nil {
		cb = logAppendCb
	}
	C.nwep_log_server_set_on_append((*C.nwep_log_server)(ls), cb, ctx)
}

// LogServerDispatch answers a log request into buf, called from a handler (nwep_log_server_dispatch).
func LogServerDispatch(ls unsafe.Pointer, connID uint64, req, buf unsafe.Pointer, nowSecs int64) int {
	return int(C.nwep_log_server_dispatch((*C.nwep_log_server)(ls), C.uint64_t(connID), (*C.nwep_message)(req), (*C.nwep_buf)(buf), C.int64_t(nowSecs)))
}

// LogEntryType returns the entry-type tag of an encoded log entry (nwep_log_entry_type).
func LogEntryType(bytes []byte) int {
	return int(C.nwep_log_entry_type(bytePtr(bytes), C.size_t(len(bytes))))
}
