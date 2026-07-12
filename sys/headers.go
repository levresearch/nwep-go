// the layer 0 header-array marshalling shared by send, notify, and stream open.

package sys

/*
#include <stdlib.h>
#include <nwep.h>
*/
import "C"

import "unsafe"

// cHeaders builds a null-terminated nwep_header array from go name/value pairs.
//
// returns a pointer to the first element (nil for no headers) and a release
// function the caller must defer to free the c strings. the trailing {nil, nil}
// sentinel the library expects is the zero-valued last element.
func cHeaders(headers [][2]string) (*C.nwep_header, func()) {
	if len(headers) == 0 {
		return nil, func() {}
	}
	arr := make([]C.nwep_header, len(headers)+1)
	frees := make([]unsafe.Pointer, 0, len(headers)*2)
	for i, h := range headers {
		name := C.CString(h[0])
		value := C.CString(h[1])
		arr[i].name = name
		arr[i].value = value
		frees = append(frees, unsafe.Pointer(name), unsafe.Pointer(value))
	}
	return &arr[0], func() {
		for _, p := range frees {
			C.free(p)
		}
	}
}
