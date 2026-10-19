package apidetect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/stretchr/testify/require"
)

// Shape: https://github.com/LostRuins/koboldcpp/blob/v1.121/koboldcpp.py
func koboldVersion() map[string]any {
	return map[string]any{
		"result": "KoboldCpp", "version": "1.121", "protected": false, "txt2img": false, "vision": true, "audio": false,
		"transcribe": false, "multiplayer": false, "websearch": true, "embeddings": false, "tts": false, "savedata": false,
		"admin": false, "guidance": false, "jinja": true, "llm": true,
	}
}

func probeKoboldAt(t *testing.T, srvURL, apiKey string) (apispec.Evidence, []string) {
	t.Helper()
	ctx, tr := withTrace(context.Background())
	var ev apispec.Evidence
	require.True(t, newTestDetector().probeKoboldCpp(ctx, srvURL, apiKey, &ev))
	return ev, tr.lines
}

func TestProbeKoboldCppFacts(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/extra/version", jsonHandler(koboldVersion()))
	mux.HandleFunc("/api/extra/true_max_context_length", jsonHandler(map[string]any{"value": 8192}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ev, trace := probeKoboldAt(t, srv.URL, "")
	require.Equal(t, "koboldcpp", ev.Stack)
	c := ev.Caps
	require.Equal(t, int64(8192), *c.Context)
	require.Equal(t, []string{"image"}, c.Modalities.Input, "audio false adds nothing")
	require.True(t, *c.WebSearch)
	require.False(t, *c.Embeddings)
	require.Nil(t, c.Tools)
	require.Nil(t, c.Reasoning)
	require.Equal(t, map[string]string{
		apispec.CapFieldContext:         "koboldcpp:/api/extra/true_max_context_length.value",
		apispec.CapFieldModalitiesInput: "koboldcpp:/api/extra/version.vision",
		apispec.CapFieldWebSearch:       "koboldcpp:/api/extra/version.websearch",
		apispec.CapFieldEmbeddings:      "koboldcpp:/api/extra/version.embeddings",
	}, c.Sources)
	require.Contains(t, trace, "identified koboldcpp by /api/extra/version")
	require.Contains(t, trace, "koboldcpp /api/extra/version vision -> modalities.input image")
	require.Contains(t, trace, "koboldcpp /api/extra/version websearch true -> webSearch")
	require.Contains(t, trace, "koboldcpp /api/extra/version embeddings false -> embeddings")
	require.Contains(t, trace, "koboldcpp /api/extra/true_max_context_length value 8192 -> context")
}

func TestProbeKoboldCppAudioAndMissingContext(t *testing.T) {
	ver := koboldVersion()
	ver["vision"], ver["audio"] = false, true
	delete(ver, "websearch")
	delete(ver, "embeddings")
	mux := http.NewServeMux()
	mux.HandleFunc("/api/extra/version", jsonHandler(ver))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ev, _ := probeKoboldAt(t, srv.URL, "")
	require.Equal(t, "koboldcpp", ev.Stack)
	require.Equal(t, []string{"audio"}, ev.Caps.Modalities.Input)
	require.Nil(t, ev.Caps.Context, "true_max_context_length unreachable: no served window")
	require.Nil(t, ev.Caps.WebSearch)
	require.Nil(t, ev.Caps.Embeddings)
}

func TestProbeKoboldCppRejectsInvalidContext(t *testing.T) {
	cases := map[string]any{
		"zero":         0,
		"negative":     -8192,
		"string":       "8192",
		"fractional":   4096.5,
		"out of range": 1 << 31,
		"absent":       nil,
	}
	for name, value := range cases {
		t.Run(name, func(t *testing.T) {
			body := map[string]any{}
			if value != nil {
				body["value"] = value
			}
			mux := http.NewServeMux()
			mux.HandleFunc("/api/extra/version", jsonHandler(koboldVersion()))
			mux.HandleFunc("/api/extra/true_max_context_length", jsonHandler(body))
			srv := httptest.NewServer(mux)
			defer srv.Close()

			ev, _ := probeKoboldAt(t, srv.URL, "")
			require.Nil(t, ev.Caps.Context)
			require.NotContains(t, ev.Caps.Sources, apispec.CapFieldContext)
		})
	}
}

// Flags that are not booleans are not reports; only the served window remains.
func TestProbeKoboldCppIgnoresWrongTypedFlags(t *testing.T) {
	ver := koboldVersion()
	ver["vision"], ver["audio"], ver["websearch"], ver["embeddings"] = "true", 1, "false", nil
	mux := http.NewServeMux()
	mux.HandleFunc("/api/extra/version", jsonHandler(ver))
	mux.HandleFunc("/api/extra/true_max_context_length", jsonHandler(map[string]any{"value": 8192}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ev, _ := probeKoboldAt(t, srv.URL, "")
	require.Equal(t, "koboldcpp", ev.Stack)
	require.Equal(t, int64(8192), *ev.Caps.Context)
	require.Nil(t, ev.Caps.Modalities)
	require.Nil(t, ev.Caps.WebSearch)
	require.Nil(t, ev.Caps.Embeddings)
	require.Equal(t, map[string]string{apispec.CapFieldContext: "koboldcpp:/api/extra/true_max_context_length.value"}, ev.Caps.Sources)
}

func TestProbeKoboldCppSendsTheKey(t *testing.T) {
	var auth []string
	record := func(body any) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			auth = append(auth, r.Header.Get("Authorization"))
			jsonHandler(body)(w, r)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/extra/version", record(koboldVersion()))
	mux.HandleFunc("/api/extra/true_max_context_length", record(map[string]any{"value": 4096}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ev, _ := probeKoboldAt(t, srv.URL, "secret")
	require.Equal(t, []string{"Bearer secret", "Bearer secret"}, auth)
	require.Equal(t, int64(4096), *ev.Caps.Context)
}

func TestDetectKoboldCppPublishesCapabilities(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/extra/version", jsonHandler(koboldVersion()))
	mux.HandleFunc("/api/extra/true_max_context_length", jsonHandler(map[string]any{"value": 8192}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	api := detect(t, srv.URL+"/v1/chat/completions", "openai", "llama-3.3-70b")
	require.NotNil(t, api)
	require.Equal(t, "koboldcpp", api.Stack)
	require.NotNil(t, api.Capabilities)
	require.Equal(t, int64(8192), *api.Capabilities.Context)
	require.True(t, *api.Capabilities.WebSearch)
	require.Contains(t, api.Capabilities.Modalities.Input, "image")
	require.Equal(t, "koboldcpp:/api/extra/true_max_context_length.value", api.Evidence["capabilities.context"])
	require.Equal(t, "koboldcpp:/api/extra/version.websearch", api.Evidence["capabilities.webSearch"])
}
