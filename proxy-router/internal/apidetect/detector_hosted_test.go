package apidetect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/stretchr/testify/require"
)

func TestHostedProbeDispatch(t *testing.T) {
	t.Run("registered vendor is probed", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/models", jsonHandler(map[string]any{
			"data": []map[string]any{{"id": "x/y", "supported_parameters": []string{"reasoning", "tools"}}},
		}))
		srv := httptest.NewServer(mux)
		defer srv.Close()
		mapHostToVendor(t, srv.URL, "openrouter")

		api, trace := newTestDetector().DetectWithTrace(context.Background(), config.ModelConfig{ModelName: "x/y", ApiType: "openai", ApiURL: srv.URL + "/chat/completions"})
		require.NotNil(t, api)
		require.Equal(t, "openrouter", api.Stack)
		require.Contains(t, trace, "host recognized as openrouter by hostname; importing its model listing")
		require.Contains(t, trace, "model listing at "+srv.URL+"/models has supported_parameters -> openrouter-style registry")
	})

	t.Run("unregistered hosted vendor is not probed", func(t *testing.T) {
		var hits atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits.Add(1)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()
		mapHostToVendor(t, srv.URL, "openai")
		_, dispatched := hostedProbes["openai"]
		require.False(t, dispatched, "this case needs a vendor without a hosted probe")

		api, trace := newTestDetector().DetectWithTrace(context.Background(), config.ModelConfig{ModelName: "gpt-x", ApiType: config.StackTransport["openai"], ApiURL: srv.URL + "/v1/chat/completions"})
		require.NotNil(t, api)
		require.Equal(t, "openai", api.Stack)
		require.Contains(t, trace, `host recognized as hosted vendor "openai" by hostname; no probing needed`)
		require.Zero(t, hits.Load())
	})

	t.Run("duplicate registration panics", func(t *testing.T) {
		require.PanicsWithValue(t, "apidetect: hosted probe for openrouter registered twice", func() {
			registerHostedProbe("openrouter", probeHostedRegistry)
		})
	})
}
