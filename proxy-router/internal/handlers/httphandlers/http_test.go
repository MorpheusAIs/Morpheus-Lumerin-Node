package httphandlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

func TestCORSAllowsMorpheusHistoryHeader(t *testing.T) {
	t.Setenv("PROXY_CORS_ALLOWED_ORIGINS", "")
	router := CreateHTTPServer(&lib.LoggerMock{}, system.HTTPAuthConfig{})

	req := httptest.NewRequest(http.MethodOptions, "/v1/chat/completions", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "x-morpheus-history")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusNoContent, w.Body.String())
	}
	if !strings.Contains(strings.ToLower(w.Header().Get("Access-Control-Allow-Headers")), "x-morpheus-history") {
		t.Fatalf("Access-Control-Allow-Headers = %q, expected x-morpheus-history", w.Header().Get("Access-Control-Allow-Headers"))
	}
}

// MyProvider (https://myprovider.mor.org) is a browser SPA that calls a
// provider's :8082 API directly with Basic Auth. It must keep working on a
// stock node with no PROXY_CORS_ALLOWED_ORIGINS set, while an arbitrary
// third-party site is still refused.
func TestCORSDefaultOrigins(t *testing.T) {
	t.Setenv("PROXY_CORS_ALLOWED_ORIGINS", "")
	router := CreateHTTPServer(&lib.LoggerMock{}, system.HTTPAuthConfig{})

	preflight := func(origin string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodOptions, "/blockchain/providers", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		req.Header.Set("Access-Control-Request-Headers", "authorization")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	for _, origin := range []string{"https://myprovider.mor.org", "https://MyProvider.mor.org", "http://127.0.0.1:8082"} {
		w := preflight(origin)
		if w.Code != http.StatusNoContent {
			t.Fatalf("origin %s: status = %d, want %d", origin, w.Code, http.StatusNoContent)
		}
		if got := w.Header().Get("Access-Control-Allow-Origin"); !strings.EqualFold(got, origin) {
			t.Fatalf("origin %s: Access-Control-Allow-Origin = %q", origin, got)
		}
	}

	for _, origin := range []string{"https://evil.example.com", "https://myprovider.mor.org.evil.example.com", "http://myprovider.mor.org"} {
		w := preflight(origin)
		if w.Code == http.StatusNoContent || w.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatalf("origin %s must be refused; status = %d, Access-Control-Allow-Origin = %q", origin, w.Code, w.Header().Get("Access-Control-Allow-Origin"))
		}
	}
}

func TestCORSAllowedOriginsEnvExtends(t *testing.T) {
	t.Setenv("PROXY_CORS_ALLOWED_ORIGINS", "https://ops.example.com/, ")
	router := CreateHTTPServer(&lib.LoggerMock{}, system.HTTPAuthConfig{})

	for _, origin := range []string{"https://ops.example.com", "https://myprovider.mor.org"} {
		req := httptest.NewRequest(http.MethodOptions, "/healthcheck", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", http.MethodGet)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNoContent {
			t.Fatalf("origin %s: status = %d, want %d", origin, w.Code, http.StatusNoContent)
		}
	}
}
