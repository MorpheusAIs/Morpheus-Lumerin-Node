package apidetect

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/stretchr/testify/require"
)

// OpenRouter-shaped listing (https://openrouter.ai/docs/api/api-reference/models/get-models).
func openRouterListing(ids ...string) map[string]any {
	data := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		data = append(data, map[string]any{"id": id, "context_length": 131072, "supported_parameters": []string{"tools", "reasoning"}})
	}
	return map[string]any{"data": data}
}

func listingServer(t *testing.T, calls *int32, status *int32) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(calls, 1)
		if code := atomic.LoadInt32(status); code != 0 {
			w.WriteHeader(int(code))
			return
		}
		jsonHandler(openRouterListing("deepseek/deepseek-chat-v3.1", "qwen/qwen3-235b-a22b"))(w, r)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestCachedListingServesTwoModelsFromOneRequest(t *testing.T) {
	var calls, status int32
	srv := listingServer(t, &calls, &status)
	d := newTestDetector()
	base := srv.URL + "/api/v1"

	const key = "sk-secret-listing-key"
	first := d.cachedListing(context.Background(), base, key)
	require.NotNil(t, first)
	require.NotNil(t, matchModelEntry(first["data"].([]any), "deepseek/deepseek-chat-v3.1"))

	ctx, trace := withTrace(context.Background())
	second := d.cachedListing(ctx, base, key)
	require.NotNil(t, matchModelEntry(second["data"].([]any), "qwen/qwen3-235b-a22b"))
	require.Equal(t, int32(1), atomic.LoadInt32(&calls), "one listing per host and key")
	require.Contains(t, strings.Join(trace.lines, "\n"), "served from the listing cache")
	require.NotContains(t, strings.Join(trace.lines, "\n"), key, "the trace never carries the key")
}

func TestCachedListingIsPerBaseAndKey(t *testing.T) {
	var calls, status int32
	srv := listingServer(t, &calls, &status)
	d := newTestDetector()
	base := srv.URL + "/api/v1"

	require.NotNil(t, d.cachedListing(context.Background(), base, "k1"))
	require.NotNil(t, d.cachedListing(context.Background(), base, "k2"))
	require.Equal(t, int32(2), atomic.LoadInt32(&calls), "a rotated key re-reads the listing")

	other := listingServer(t, &calls, &status)
	require.NotNil(t, d.cachedListing(context.Background(), other.URL+"/api/v1", "k1"))
	require.Equal(t, int32(3), atomic.LoadInt32(&calls), "another host is another listing")

	d.mu.Lock()
	for key := range d.listings {
		require.NotContains(t, key, "k1")
		require.NotContains(t, key, "k2")
	}
	d.mu.Unlock()
}

func TestCachedListingNeverCachesAFailedRead(t *testing.T) {
	var calls, status int32
	srv := listingServer(t, &calls, &status)
	d := newTestDetector()
	base := srv.URL + "/api/v1"

	atomic.StoreInt32(&status, http.StatusServiceUnavailable)
	require.Nil(t, d.cachedListing(context.Background(), base, "k"))
	require.Nil(t, d.cachedListing(context.Background(), base, "k"))
	require.Equal(t, int32(2), atomic.LoadInt32(&calls), "a failure is retried, not remembered")

	atomic.StoreInt32(&status, 0)
	require.NotNil(t, d.cachedListing(context.Background(), base, "k"))
	require.NotNil(t, d.cachedListing(context.Background(), base, "k"))
	require.Equal(t, int32(3), atomic.LoadInt32(&calls))
}

func TestCachedListingOverLimitIsNotCached(t *testing.T) {
	prev := maxRegistryBody
	maxRegistryBody = 256
	t.Cleanup(func() { maxRegistryBody = prev })

	var calls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/models", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"x/y","pad":"` + strings.Repeat("p", 1024) + `"}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	d := newTestDetector()
	require.Nil(t, d.cachedListing(context.Background(), srv.URL+"/api/v1", ""))
	require.Nil(t, d.cachedListing(context.Background(), srv.URL+"/api/v1", ""))
	require.Equal(t, int32(2), atomic.LoadInt32(&calls))
	d.mu.Lock()
	require.Empty(t, d.listings)
	d.mu.Unlock()
}

func TestCachedListingExpiresWithTheDetectorTTL(t *testing.T) {
	var calls, status int32
	srv := listingServer(t, &calls, &status)
	d := newTestDetector()
	base := srv.URL + "/api/v1"

	require.NotNil(t, d.cachedListing(context.Background(), base, "k"))
	d.mu.Lock()
	for key, e := range d.listings {
		e.at = time.Now().Add(-d.ttl)
		d.listings[key] = e
	}
	d.mu.Unlock()
	require.NotNil(t, d.cachedListing(context.Background(), base, "k"))
	require.Equal(t, int32(2), atomic.LoadInt32(&calls), "an expired listing is re-read")
}

func TestCachedListingBoundedUnderFreshKeyChurn(t *testing.T) {
	var calls, status int32
	srv := listingServer(t, &calls, &status)
	d := newTestDetector()

	seedAt := time.Now().Add(-time.Minute)
	d.mu.Lock()
	for i := 0; i < listingCacheMax+10; i++ {
		d.listings[fmt.Sprintf("h%03d", i)] = listingEntry{list: map[string]any{}, at: seedAt.Add(time.Duration(i) * time.Millisecond)}
	}
	d.mu.Unlock()

	base := srv.URL + "/api/v1"
	require.NotNil(t, d.cachedListing(context.Background(), base, "k"))

	d.mu.Lock()
	defer d.mu.Unlock()
	require.Len(t, d.listings, listingCacheMax/2+1, "evicted down to half the bound, then the new entry")
	require.Contains(t, d.listings, listingKey(base, "k"))
	require.NotContains(t, d.listings, "h000", "the oldest entries go first")
}

// Two models configured against one registry host cost one GET /models per
// detector TTL through the probe itself, not only through cachedListing.
func TestProbeRegistryShapeSharesTheListingBetweenModels(t *testing.T) {
	var calls, status int32
	srv := listingServer(t, &calls, &status)
	d := newTestDetector()
	bases := []string{srv.URL + "/api/v1"}
	const key = "sk-secret-listing-key"

	for i, model := range []string{"deepseek/deepseek-chat-v3.1", "qwen/qwen3-235b-a22b"} {
		ctx, trace := withTrace(context.Background())
		ev := apispec.Evidence{ModelName: model}
		d.probeRegistryShape(ctx, bases, model, key, &ev)
		require.Equal(t, "openrouter", ev.Stack, model)
		require.EqualValues(t, 131072, *ev.Caps.Context, model)
		require.Equal(t, true, *ev.Caps.Tools, model)
		require.Equal(t, int32(1), atomic.LoadInt32(&calls), "one listing request after model %d", i+1)
		joined := strings.Join(trace.lines, "\n")
		require.Equal(t, i == 1, strings.Contains(joined, "served from the listing cache"), model)
		require.NotContains(t, joined, key)
	}
}
