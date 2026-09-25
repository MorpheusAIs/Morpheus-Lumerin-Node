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

// The one JSON-number reader every probe shares: whole float64 values only,
// within the range a float64 represents exactly; zero and negatives pass so
// the callers that trace "out of range; ignored" still see the value.
func TestJSONWholeNumber(t *testing.T) {
	for _, v := range []any{2.5, float64(1 << 60), -float64(1 << 60), 1e30, 128000, "1", true, nil, []any{1}} {
		_, ok := jsonWholeNumber(v)
		require.False(t, ok, "%v", v)
	}
	for _, v := range []float64{0, -1, 1, 128000, 1 << 31, 1 << 53, -(1 << 53)} {
		n, ok := jsonWholeNumber(v)
		require.True(t, ok, "%v", v)
		require.Equal(t, int64(v), n)
	}
}

// jsonPositiveInt is jsonWholeNumber inside the capability bound (0, 1<<31).
func TestJSONPositiveInt(t *testing.T) {
	cases := []struct {
		v    any
		want int64
		ok   bool
	}{
		{float64(40960), 40960, true},
		{float64(1), 1, true},
		{float64(1<<31 - 1), 1<<31 - 1, true},
		{float64(1 << 31), 0, false},
		{float64(4294967296), 0, false},
		{float64(0), 0, false},
		{float64(-5), 0, false},
		{1.5, 0, false},
		{1e30, 0, false},
		{"7", 0, false},
		{nil, 0, false},
	}
	for _, c := range cases {
		got, ok := jsonPositiveInt(c.v)
		require.Equal(t, c.ok, ok, "%v", c.v)
		require.Equal(t, c.want, got, "%v", c.v)
	}
}

func TestMatchModelEntryBy(t *testing.T) {
	data := []any{
		map[string]any{"key": "org/model-a", "id": "ignored"},
		map[string]any{"key": "Model-B"},
		"not an object",
	}
	require.Equal(t, "Model-B", matchModelEntryBy(data, "model-b", "key")["key"], "exact match is case-insensitive")
	require.Equal(t, "org/model-a", matchModelEntryBy(data, "model-a:tee=1", "key")["key"], "bare id, then a prefix/ suffix match")
	require.Nil(t, matchModelEntryBy(data, "model-c", "key"))
	require.Nil(t, matchModelEntryBy(data, "model-b", "id"), "the identifier field is the caller's")
	require.Equal(t, "Model-B", matchModelEntryBy(data[1:2], "anything", "key")["key"], "a lone entry matches any name")
}
