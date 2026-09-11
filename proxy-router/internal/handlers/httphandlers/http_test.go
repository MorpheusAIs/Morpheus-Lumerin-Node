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
