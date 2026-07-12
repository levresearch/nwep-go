// the layer 0 response cache, a signed-response store for clients NW060900.

package sys

/*
#include <stdlib.h>
#include <nwep.h>
*/
import "C"

import "unsafe"

// CacheStats is a snapshot of a cache's hit, miss, store, and eviction counters.
type CacheStats struct {
	Hits, Misses, Stores, Evictions uint64
}

// CacheCreate allocates a response cache bounded by bytes and entries (nwep_cache_create).
func CacheCreate(maxBytes, maxEntries int) unsafe.Pointer {
	return unsafe.Pointer(C.nwep_cache_create(C.size_t(maxBytes), C.size_t(maxEntries)))
}

// CacheFree frees a cache (nwep_cache_free).
func CacheFree(cache unsafe.Pointer) {
	C.nwep_cache_free((*C.nwep_cache)(cache))
}

// CacheClear drops every entry but keeps the cache allocated (nwep_cache_clear).
func CacheClear(cache unsafe.Pointer) {
	C.nwep_cache_clear((*C.nwep_cache)(cache))
}

// CacheStatsGet reads the cache's counters (nwep_cache_stats).
func CacheStatsGet(cache unsafe.Pointer) CacheStats {
	var hits, misses, stores, evictions C.uint64_t
	C.nwep_cache_stats((*C.nwep_cache)(cache), &hits, &misses, &stores, &evictions)
	return CacheStats{uint64(hits), uint64(misses), uint64(stores), uint64(evictions)}
}

// CachePutSigned stores a verified response keyed by method, path, and origin (nwep_cache_put_signed).
func CachePutSigned(cache unsafe.Pointer, method, path string, resp unsafe.Pointer, originPubkey [PubKeySize]byte, nowSecs uint64) int {
	cm := C.CString(method)
	defer C.free(unsafe.Pointer(cm))
	cp := C.CString(path)
	defer C.free(unsafe.Pointer(cp))
	return int(C.nwep_cache_put_signed((*C.nwep_cache)(cache), cm, cp, (*C.nwep_message)(resp), (*C.uint8_t)(unsafe.Pointer(&originPubkey[0])), C.uint64_t(nowSecs)))
}

// CacheGetSigned fetches a still-fresh cached response, if present (nwep_cache_get_signed).
func CacheGetSigned(cache unsafe.Pointer, method, path string, originPubkey [PubKeySize]byte, nowSecs uint64) (unsafe.Pointer, int) {
	cm := C.CString(method)
	defer C.free(unsafe.Pointer(cm))
	cp := C.CString(path)
	defer C.free(unsafe.Pointer(cp))
	var out *C.nwep_message
	rc := int(C.nwep_cache_get_signed((*C.nwep_cache)(cache), cm, cp, (*C.uint8_t)(unsafe.Pointer(&originPubkey[0])), C.uint64_t(nowSecs), &out))
	return unsafe.Pointer(out), rc
}
