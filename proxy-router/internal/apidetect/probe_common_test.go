package apidetect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/stretchr/testify/require"
)

// An auth proxy in front of an engine may answer a keyless request with a
// 404, a redirect to a login page on another host or an HTML page. Every
// fingerprint request carries the model's key, so the proxy lets the first
// one in and the engine is identified.
func TestFingerprintThroughAnAuthProxy(t *testing.T) {
	type probeFunc func(d *Detector, ctx context.Context, base string, ev *apispec.Evidence) bool
	engines := []struct {
		stack string
		path  string
		body  any
		probe probeFunc
	}{
		{"vllm", "/version", map[string]any{"version": "0.30.0"}, func(d *Detector, ctx context.Context, base string, ev *apispec.Evidence) bool {
			return d.probeVLLM(ctx, base, "m", "secret", ev)
		}},
		{"ollama", "/api/tags", map[string]any{"models": []any{}}, func(d *Detector, ctx context.Context, base string, ev *apispec.Evidence) bool {
			return d.probeOllama(ctx, base, "m", "secret", ev)
		}},
		{"tgi", "/info", map[string]any{"model_id": "m"}, func(d *Detector, ctx context.Context, base string, ev *apispec.Evidence) bool {
			return d.probeTGI(ctx, base, "secret", ev)
		}},
		{"koboldcpp", "/api/extra/version", map[string]any{"result": "KoboldCpp"}, func(d *Detector, ctx context.Context, base string, ev *apispec.Evidence) bool {
			return d.probeKoboldCpp(ctx, base, "secret", ev)
		}},
		{"lmstudio", "/api/v1/models", map[string]any{"models": []any{}}, func(d *Detector, ctx context.Context, base string, ev *apispec.Evidence) bool {
			return d.probeLMStudio(ctx, base, "m", "secret", ev)
		}},
	}
	walls := []struct {
		name   string
		refuse http.HandlerFunc
	}{
		{"404", http.NotFound},
		{"a redirect to a login page", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://login.example.invalid/", http.StatusFound)
		}},
		{"an HTML page", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html>sign in</html>"))
		}},
	}
	for _, e := range engines {
		for _, wall := range walls {
			t.Run(e.stack+" behind "+wall.name, func(t *testing.T) {
				var mu sync.Mutex
				var auth []string
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					mu.Lock()
					auth = append(auth, r.Header.Get("Authorization"))
					mu.Unlock()
					switch {
					case r.Header.Get("Authorization") != "Bearer secret":
						wall.refuse(w, r)
					case r.URL.Path == e.path:
						jsonHandler(e.body)(w, r)
					default:
						http.NotFound(w, r)
					}
				}))
				defer srv.Close()

				var ev apispec.Evidence
				require.True(t, e.probe(newTestDetector(), context.Background(), srv.URL, &ev))
				require.Equal(t, e.stack, ev.Stack)
				mu.Lock()
				defer mu.Unlock()
				for i, a := range auth {
					require.Equal(t, "Bearer secret", a, "request %d", i)
				}
			})
		}
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
	require.Equal(t, "Model-B", entryBy(data, "model-b", "key")["key"], "exact match is case-insensitive")
	require.Equal(t, "org/model-a", entryBy(data, "model-a:tee=1", "key")["key"], "bare id, then a prefix/ suffix match")
	require.Nil(t, entryBy(data, "model-c", "key"))
	require.Nil(t, entryBy(data, "model-b", "id"), "the identifier field is the caller's")
	require.Equal(t, "Model-B", entryBy(data[1:2], "anything", "key")["key"], "a lone entry matches any name")
}

// entryBy and matchedEntry keep only the entry of findModelEntryBy.
func entryBy(data []any, modelName, field string) map[string]any {
	entry, _ := findModelEntryBy(data, modelName, field)
	return entry
}

func matchedEntry(data []any, modelName string) map[string]any {
	entry, _ := findModelEntry(data, modelName)
	return entry
}
