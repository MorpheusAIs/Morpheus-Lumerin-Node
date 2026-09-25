package apidetect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/stretchr/testify/require"
)

// Shape: https://github.com/huggingface/text-generation-inference/blob/v3.3.7/router/src/lib.rs
func tgiInfo() map[string]any {
	return map[string]any{
		"model_id": "mistralai/Mistral-7B-Instruct-v0.3", "model_sha": "abc", "model_pipeline_tag": "text-generation",
		"max_concurrent_requests": 128, "max_best_of": 2, "max_stop_sequences": 4, "max_input_tokens": 4095,
		"max_total_tokens": 4096, "validation_workers": 2, "max_client_batch_size": 4, "router": "text-generation-router",
		"version": "3.3.7",
	}
}

func probeTGIAt(t *testing.T, srvURL, apiKey string) (apispec.Evidence, []string) {
	t.Helper()
	ctx, tr := withTrace(context.Background())
	var ev apispec.Evidence
	require.True(t, newTestDetector().probeTGI(ctx, srvURL, apiKey, &ev))
	return ev, tr.lines
}

func TestProbeTGIFacts(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/info", jsonHandler(tgiInfo()))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ev, trace := probeTGIAt(t, srv.URL, "")
	require.Equal(t, "tgi", ev.Stack)
	require.Equal(t, "mistralai/Mistral-7B-Instruct-v0.3", ev.ServedModelID)
	c := ev.Caps
	require.Equal(t, int64(4096), *c.Context)
	require.Nil(t, c.MaxOutput, "max_total_tokens - max_input_tokens is a guaranteed minimum, not a maximum")
	require.Nil(t, c.ContextTrained)
	require.Equal(t, map[string]string{apispec.CapFieldContext: "tgi:/info.max_total_tokens"}, c.Sources)
	require.Contains(t, trace, "tgi /info max_total_tokens 4096 -> context")
	for _, line := range trace {
		require.NotContains(t, line, "maxOutput")
	}
}

// max_input_tokens, whatever its value, never becomes maxOutput (contract
// amendment a): the difference to max_total_tokens is the completion budget
// at the longest prompt, a floor.
func TestProbeTGINeverPublishesMaxOutput(t *testing.T) {
	cases := map[string]map[string]any{
		"input absent":       {},
		"input zero":         {"max_input_tokens": 0},
		"input negative":     {"max_input_tokens": -1},
		"input string":       {"max_input_tokens": "4095"},
		"input equals total": {"max_input_tokens": 4096},
		"input above total":  {"max_input_tokens": 8192},
		"input fractional":   {"max_input_tokens": 4095.5},
		"input far below":    {"max_input_tokens": 1024},
	}
	for name, patch := range cases {
		t.Run(name, func(t *testing.T) {
			info := tgiInfo()
			delete(info, "max_input_tokens")
			for k, v := range patch {
				info[k] = v
			}
			mux := http.NewServeMux()
			mux.HandleFunc("/info", jsonHandler(info))
			srv := httptest.NewServer(mux)
			defer srv.Close()

			ev, _ := probeTGIAt(t, srv.URL, "")
			require.Equal(t, int64(4096), *ev.Caps.Context)
			require.Nil(t, ev.Caps.MaxOutput)
		})
	}
}

// Without a valid window nothing is published: maxOutput has no operand.
func TestProbeTGIWithoutTotalSetsNothing(t *testing.T) {
	cases := map[string]map[string]any{
		"total absent":       {},
		"total string":       {"max_total_tokens": "4096"},
		"total fractional":   {"max_total_tokens": 4096.5},
		"total zero":         {"max_total_tokens": 0},
		"total negative":     {"max_total_tokens": -4096},
		"total out of range": {"max_total_tokens": 1 << 31},
	}
	for name, patch := range cases {
		t.Run(name, func(t *testing.T) {
			info := tgiInfo()
			delete(info, "max_total_tokens")
			for k, v := range patch {
				info[k] = v
			}
			mux := http.NewServeMux()
			mux.HandleFunc("/info", jsonHandler(info))
			srv := httptest.NewServer(mux)
			defer srv.Close()

			ev, _ := probeTGIAt(t, srv.URL, "")
			require.Equal(t, "tgi", ev.Stack)
			require.True(t, ev.Caps.Empty())
		})
	}
}

func TestProbeTGIKeyFreeFirst(t *testing.T) {
	t.Run("no key sent to the open /info", func(t *testing.T) {
		var auth []string
		mux := http.NewServeMux()
		mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
			auth = append(auth, r.Header.Get("Authorization"))
			jsonHandler(tgiInfo())(w, r)
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		probeTGIAt(t, srv.URL, "secret")
		require.Equal(t, []string{""}, auth)
	})
	t.Run("key sent only after 403", func(t *testing.T) {
		var auth []string
		mux := http.NewServeMux()
		mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
			auth = append(auth, r.Header.Get("Authorization"))
			if r.Header.Get("Authorization") != "Bearer secret" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			jsonHandler(tgiInfo())(w, r)
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		ev, _ := probeTGIAt(t, srv.URL, "secret")
		require.Equal(t, []string{"", "Bearer secret"}, auth)
		require.Equal(t, int64(4096), *ev.Caps.Context)
	})
}

func TestDetectTGIPublishesCapabilities(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/info", jsonHandler(tgiInfo()))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "mistral-7b")
	require.NotNil(t, api)
	require.Equal(t, "tgi", api.Stack)
	require.NotNil(t, api.Capabilities)
	require.Equal(t, int64(4096), *api.Capabilities.Context)
	require.Nil(t, api.Capabilities.MaxOutput)
	require.Equal(t, "tgi:/info.max_total_tokens", api.Evidence["capabilities.context"])
	require.NotContains(t, api.Evidence, "capabilities.maxOutput")
}
