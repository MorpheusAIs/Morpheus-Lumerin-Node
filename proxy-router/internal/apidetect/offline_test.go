package apidetect

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// No test in this package reaches the network. Every hosted vendor now has a
// probe (probe_vendors.go, probe_registry.go), so a fixture naming a real
// vendor host (api.anthropic.com, api.groq.com, ...) would send a live HTTPS
// request; such fixtures point at stubVendorHost instead and assert how many
// requests the stub saw. The invariant is checked by running the package
// with every outbound proxy pointed at a closed port, so any real request
// fails fast and the run must still be green:
//
//	HTTPS_PROXY=http://127.0.0.1:9 HTTP_PROXY=http://127.0.0.1:9 NO_PROXY=127.0.0.1,localhost \
//	  go test -race -count=1 ./internal/apidetect
//
// (net/http reads the proxy variables once per process; httptest servers are
// on 127.0.0.1 and excluded by NO_PROXY.) Vendor probes tolerate a failed
// lookup silently, so that run proves no request leaves the host; the
// request-count assertions on the stubs are what fail when a fixture is
// pointed back at a real vendor.

// stubVendorHost stands in for a hosted vendor's real host: an httptest
// server that counts requests and answers 404 (the silent miss every vendor
// probe tolerates), mapped to vendor for the test's lifetime. The returned
// URL is the stub's origin; the counter is the number of requests it saw.
func stubVendorHost(t *testing.T, vendor string) (string, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	mapHostToVendor(t, srv.URL, vendor)
	return srv.URL, &hits
}
