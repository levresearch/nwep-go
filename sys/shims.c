// the c callback trampolines, in a real .c file so they are compiled exactly once.
//
// each forwards the library's callback to its go export (declared in the generated
// _cgo_export.h). the go exports take non-const pointers because cgo //export
// drops const, so the shims cast const away from the library's const arguments.
// see callbacks.go for the go side and the rationale (guide 9).

#include <nwep.h>
#include "_cgo_export.h"

int nwepServerHandlerShim(nwep_server* s, uint64_t c, uint64_t st, const nwep_message* r, nwep_buf* b, void* u) {
    return nwepGoServerHandler(s, c, st, (nwep_message*)r, b, u);
}

void nwepRequestDoneShim(nwep_client* c, uint64_t id, int status, nwep_message* m, void* u) {
    nwepGoRequestDone(c, id, status, m, u);
}

void nwepLogAppendShim(void* ctx, const uint8_t* e, size_t n, uint64_t idx) {
    nwepGoLogAppend(ctx, (uint8_t*)e, n, idx);
}
