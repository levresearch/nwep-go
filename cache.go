// the response cache, a client-side store of verified signed responses NW060900.

package nwep

import (
	"unsafe"

	"nwep/sys"
)

// CacheStats is a snapshot of a cache's hit, miss, store, and eviction counters.
type CacheStats = sys.CacheStats

// Cache is a bounded store of signed responses, attached to a client NW060900.
type Cache struct {
	ptr unsafe.Pointer
}

// NewCache allocates a cache bounded by total bytes and entry count.
func NewCache(maxBytes, maxEntries int) *Cache {
	return &Cache{ptr: sys.CacheCreate(maxBytes, maxEntries)}
}

// Clear drops every entry but keeps the cache allocated.
func (c *Cache) Clear() { sys.CacheClear(c.ptr) }

// Stats returns the cache's counters.
func (c *Cache) Stats() CacheStats { return sys.CacheStatsGet(c.ptr) }

// PutSigned stores a verified response keyed by method, path, and origin NW060900.
func (c *Cache) PutSigned(method, path string, resp *Message, originPubkey [32]byte, nowSecs uint64) error {
	return check(sys.CachePutSigned(c.ptr, method, path, resp.ptr, originPubkey, nowSecs))
}

// GetSigned fetches a still-fresh cached response, if present NW060900.
//
// returns the response and true on a hit, a nil message and false on a miss.
func (c *Cache) GetSigned(method, path string, originPubkey [32]byte, nowSecs uint64) (*Message, bool, error) {
	ptr, rc := sys.CacheGetSigned(c.ptr, method, path, originPubkey, nowSecs)
	if rc == sys.ErrAppNotFound || ptr == nil {
		return nil, false, nil
	}
	if err := check(rc); err != nil {
		return nil, false, err
	}
	return &Message{ptr: ptr, owned: true}, true, nil
}

// Close frees the cache (nwep_cache_free).
func (c *Cache) Close() {
	if c.ptr != nil {
		sys.CacheFree(c.ptr)
		c.ptr = nil
	}
}

// Raw returns the underlying sys cache pointer, the no-cliffs escape to L0 NWG0200.
func (c *Cache) Raw() unsafe.Pointer { return c.ptr }
