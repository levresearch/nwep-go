// the layer 0 read side of a decoded message, headers, status, and body NW050000 NW060000.

package sys

/*
#include <stdlib.h>
#include <nwep.h>
*/
import "C"

import "unsafe"

// MessageGetHeader returns the value of header name, or empty when absent (nwep_message_get_header).
func MessageGetHeader(msg unsafe.Pointer, name string) string {
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	v := C.nwep_message_get_header((*C.nwep_message)(msg), cn)
	if v == nil {
		return ""
	}
	return C.GoString(v)
}

// MessageHeaderCount returns how many headers the message carries (nwep_message_header_count).
func MessageHeaderCount(msg unsafe.Pointer) int {
	return int(C.nwep_message_header_count((*C.nwep_message)(msg)))
}

// MessageHeaderAt returns the name and value of header i (nwep_message_header_at).
func MessageHeaderAt(msg unsafe.Pointer, i int) (name, value string, rc int) {
	var cn, cv *C.char
	rc = int(C.nwep_message_header_at((*C.nwep_message)(msg), C.size_t(i), &cn, &cv))
	if rc == 0 {
		name = C.GoString(cn)
		value = C.GoString(cv)
	}
	return
}

// MessageGetStatus returns the response status string, or empty for a request (nwep_message_get_status).
func MessageGetStatus(msg unsafe.Pointer) string {
	s := C.nwep_message_get_status((*C.nwep_message)(msg))
	if s == nil {
		return ""
	}
	return C.GoString(s)
}

// MessageGetBody copies out the decoded body bytes (nwep_message_get_body).
func MessageGetBody(msg unsafe.Pointer) []byte {
	var n C.size_t
	p := C.nwep_message_get_body((*C.nwep_message)(msg), &n)
	if p == nil || n == 0 {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(p), C.int(n))
}

// MessageFree frees a message handle owned by the caller (nwep_message_free).
func MessageFree(msg unsafe.Pointer) {
	C.nwep_message_free((*C.nwep_message)(msg))
}
