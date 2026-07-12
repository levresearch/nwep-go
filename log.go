// the merkle log and log server, the transparency log half NW120014.
//
// Log is an append-only rfc-6962 merkle tree of key-management entries. LogServer
// wraps one behind a node that answers log queries, dispatched from a handler.

package nwep

import (
	"unsafe"

	"github.com/levresearch/nwep-go/sys"
)

// Log is an append-only merkle log of key-management entries NW120000.
type Log struct {
	ptr unsafe.Pointer
}

// NewLog allocates an empty merkle log (nwep_log_create).
func NewLog() *Log { return &Log{ptr: sys.LogCreate()} }

// Append adds an entry and returns its index (nwep_log_append).
//
// returns the zero-based index the entry landed at.
// errors when the entry is malformed or the log rejects it.
func (l *Log) Append(entry []byte) (uint64, error) {
	rc := sys.LogAppend(l.ptr, entry)
	if rc < 0 {
		return 0, newError(int(rc))
	}
	return uint64(rc), nil
}

// Size returns the number of entries in the log.
func (l *Log) Size() uint64 { return sys.LogSize(l.ptr) }

// Root returns the current rfc-6962 merkle root (nwep_log_root).
func (l *Log) Root() ([32]byte, error) {
	root, rc := sys.LogRoot(l.ptr)
	return root, check(rc)
}

// Close frees the log (nwep_log_free).
func (l *Log) Close() {
	if l.ptr != nil {
		sys.LogFree(l.ptr)
		l.ptr = nil
	}
}

// Raw returns the underlying sys log pointer, the no-cliffs escape to L0 NWG0200.
func (l *Log) Raw() unsafe.Pointer { return l.ptr }

// EntryType returns the type tag of an encoded log entry (nwep_log_entry_type).
func EntryType(entry []byte) (int, error) {
	t := sys.LogEntryType(entry)
	if t < 0 {
		return 0, newError(t)
	}
	return t, nil
}

// LogServer answers log queries over a wrapped Log, dispatched from a handler NW120014.
type LogServer struct {
	ptr unsafe.Pointer
}

// NewLogServer wraps a log behind a query-answering server with identity (nwep_log_server_create).
func NewLogServer(identity *Identity, log *Log) *LogServer {
	return &LogServer{ptr: sys.LogServerCreate(identity.Raw(), log.ptr)}
}

// OnAppend registers the accepted-entry persistence hook, or clears it with nil.
func (ls *LogServer) OnAppend(fn func(entry []byte, index uint64)) {
	if fn == nil {
		sys.LogServerSetOnAppend(ls.ptr, nil)
		return
	}
	sys.LogServerSetOnAppend(ls.ptr, sys.LogAppendFunc(fn))
}

// Dispatch answers a log request into res, called from a server handler (nwep_log_server_dispatch).
func (ls *LogServer) Dispatch(connID uint64, req *Message, res *Responder, nowSecs int64) error {
	return check(sys.LogServerDispatch(ls.ptr, connID, req.ptr, res.buf, nowSecs))
}

// Close frees the log server and retires its on-append hook (nwep_log_server_free).
func (ls *LogServer) Close() {
	if ls.ptr != nil {
		sys.LogServerFree(ls.ptr)
		ls.ptr = nil
	}
}

// Raw returns the underlying sys log server pointer, the no-cliffs escape NWG0200.
func (ls *LogServer) Raw() unsafe.Pointer { return ls.ptr }
