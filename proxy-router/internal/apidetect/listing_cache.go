package apidetect

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
)

// A registry host serves one catalogue for every model configured against
// it; 64 distinct base URL + key pairs is far beyond one node's hosts.
const listingCacheMax = 64

type listingEntry struct {
	list map[string]any
	at   time.Time
}

// cachedListing is GET {base}/models memoised per base URL and key for the
// detector TTL, so N models on one registry host cost one listing per sweep.
// A failed read (transport error, non-2xx, over maxRegistryBody, not JSON)
// is never cached. The returned map is shared between callers and must be
// read only.
func (d *Detector) cachedListing(ctx context.Context, base, apiKey string) map[string]any {
	key := listingKey(base, apiKey)
	d.mu.Lock()
	entry, ok := d.listings[key]
	d.mu.Unlock()
	if ok && time.Since(entry.at) < d.ttl {
		tracef(ctx, "GET %s/models -> served from the listing cache", base)
		return entry.list
	}

	list := d.requestJSON(ctx, http.MethodGet, base+"/models", apiKey, nil, maxRegistryBody)
	if list == nil {
		return nil
	}

	d.mu.Lock()
	if len(d.listings) >= listingCacheMax {
		sweepCache(d.listings, listingCacheMax, d.ttl, func(e listingEntry) time.Time { return e.at })
	}
	d.listings[key] = listingEntry{list: list, at: time.Now()}
	d.mu.Unlock()
	return list
}

// The key is hashed so the cache map never holds a bearer in the clear.
func listingKey(base, apiKey string) string {
	sum := sha256.Sum256([]byte(apiKey))
	return base + "|" + hex.EncodeToString(sum[:])
}
