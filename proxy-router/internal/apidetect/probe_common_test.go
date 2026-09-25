package apidetect

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetJSONKeyFree(t *testing.T) {
	cases := []struct {
		name        string
		firstStatus int // answer to a request without Authorization
		apiKey      string
		wantAuth    []string // Authorization header of each request, in order
		wantObject  bool
		wantRetry   bool
	}{
		{"open endpoint answers without a key", http.StatusOK, "k", []string{""}, true, false},
		{"401 is retried with the key", http.StatusUnauthorized, "k", []string{"", "Bearer k"}, true, true},
		{"403 is retried with the key", http.StatusForbidden, "k", []string{"", "Bearer k"}, true, true},
		{"404 is final", http.StatusNotFound, "k", []string{""}, false, false},
		{"500 is final", http.StatusInternalServerError, "k", []string{""}, false, false},
		{"401 without a configured key is final", http.StatusUnauthorized, "", []string{""}, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var seen []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				auth := r.Header.Get("Authorization")
				seen = append(seen, auth)
				if auth == "" && c.firstStatus != http.StatusOK {
					w.WriteHeader(c.firstStatus)
					return
				}
				jsonHandler(map[string]any{"version": "1"})(w, r)
			}))
			defer srv.Close()

			ctx, trace := withTrace(context.Background())
			out := newTestDetector().getJSONKeyFree(ctx, srv.URL+"/version", c.apiKey)
			require.Equal(t, c.wantAuth, seen)
			require.Equal(t, c.wantObject, out != nil)

			retryLine := fmt.Sprintf("GET %s/version -> HTTP %d without a key; retrying with the configured key", srv.URL, c.firstStatus)
			if c.wantRetry {
				require.Contains(t, trace.lines, retryLine)
			} else {
				require.NotContains(t, trace.lines, retryLine)
			}
		})
	}
}
