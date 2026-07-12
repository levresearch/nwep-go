// the layer 0 uri parse, web://nodeid[:port]/path NW110900.

package sys

/*
#include <stdlib.h>
#include <nwep.h>
*/
import "C"

import "unsafe"

// URIParse parses a web://nodeid_base58[:port]/path uri (nwep_uri_parse).
//
// returns the target node_id, the port (0 when absent), and the path. the c path
// borrows the input buffer, so it is copied out into an owned go string here.
func URIParse(input string) (nodeID [NodeIDSize]byte, port uint16, path string, rc int) {
	cs := C.CString(input)
	defer C.free(unsafe.Pointer(cs))
	var u C.nwep_uri
	rc = int(C.nwep_uri_parse(&u, cs, C.size_t(len(input))))
	if rc != 0 {
		return
	}
	nodeID = *(*[NodeIDSize]byte)(unsafe.Pointer(&u.node_id.bytes[0]))
	port = uint16(u.port)
	if u.path != nil && u.path_len > 0 {
		path = C.GoStringN(u.path, C.int(u.path_len))
	}
	return
}
