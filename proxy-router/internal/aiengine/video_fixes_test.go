package aiengine

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/stretchr/testify/require"
)

// Renderful's Seedance rows have different limits: the base spec follows the
// catalog so a quote is never sold for a clip the vendor refuses.
func TestRenderfulBaseVideoSpecPerSeedanceModel(t *testing.T) {
	cases := []struct {
		model       string
		min, max    int
		resolutions []string
		audio       bool
	}{
		{"seedance-2.5-t2v", 4, 30, []string{"480p", "720p"}, true},
		{"seedance-2.0-t2v", 4, 15, []string{"480p", "720p"}, true},
		{"seedance-2.0-fast-t2v", 4, 15, []string{"480p", "720p"}, true},
		{"seedance-1.5-pro", 4, 12, []string{"480p", "720p"}, true},
		{"seedance-1.0-pro", 2, 12, []string{"720p", "1080p"}, false},
		{"kling-v2-master", 4, 30, []string{"480p", "720p"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.model, func(t *testing.T) {
			spec := NewRenderfulEngine(tc.model, "", "", &lib.LoggerMock{}, nil).BaseVideoSpec()
			require.Equal(t, tc.min, spec.MinDuration)
			require.Equal(t, tc.max, spec.MaxDuration)
			require.Equal(t, tc.resolutions, spec.Resolutions)
			require.Equal(t, tc.audio, spec.Audio)
			require.NoError(t, spec.Validate())
		})
	}
}

// A Seedance first-frame job on Renderful keeps the frame's ratio, and the
// task says so for the receipt.
func TestRenderfulSeedanceFirstFrameIsAdaptive(t *testing.T) {
	vendor := newVideoVendor(t, videoJSONReply(http.StatusAccepted, `{"id":"gen_1","status":"queued"}`))
	engine := NewRenderfulEngine("seedance-2.5-t2v", vendor.URL, "k", &lib.LoggerMock{}, nil)

	res := videoTestResolved()
	res.AspectRatio = "16:9"
	res.HasImage = true
	task, err := engine.CreateVideoJob(context.Background(), &VideoJobInput{Resolved: res, ImageURL: "https://img.example.com/a.png", Audio: true})
	require.NoError(t, err)
	require.Equal(t, "adaptive", task.AspectRatio)

	body := decodeVideoBody(t, vendor.Calls()[0].Body)
	require.Equal(t, "image-to-video", body["type"])
	require.Equal(t, "seedance-2.5-i2v", body["model"])
	require.Equal(t, "adaptive", body["aspect_ratio"])

	// Text to video keeps the asked ratio and reports no change.
	task, err = engine.CreateVideoJob(context.Background(), &VideoJobInput{Resolved: videoTestResolved(), Audio: true})
	require.NoError(t, err)
	require.Empty(t, task.AspectRatio)
}

func TestVideoTransportErrorHidesTheURL(t *testing.T) {
	for _, engine := range []VideoJobEngine{
		NewRenderfulEngine("seedance-2.5-t2v", "http://10.255.255.1:1", "k", &lib.LoggerMock{}, nil),
		NewBytePlusVideoEngine("dreamina-seedance-2-5-260628", "http://10.255.255.1:1", "k", &lib.LoggerMock{}, nil),
	} {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := engine.CreateVideoJob(ctx, &VideoJobInput{Resolved: videoTestResolved()})
		require.True(t, errors.Is(err, ErrVideoVendorUnreachable), "%v", err)
		require.NotContains(t, err.Error(), "10.255.255.1")

		_, err = engine.FetchVideoStatus(ctx, "task-1")
		require.True(t, errors.Is(err, ErrVideoVendorUnreachable), "%v", err)
		require.NotContains(t, err.Error(), "10.255.255.1")

		err = engine.ProbeVideoVendor(ctx)
		require.True(t, errors.Is(err, ErrVideoVendorUnreachable), "%v", err)
		require.NotContains(t, err.Error(), "10.255.255.1")
	}
}

// The BytePlus probe lists one task of the account: a GET that starts no
// render, answered 200 for a working key.
func TestBytePlusProbeVideoVendor(t *testing.T) {
	vendor := newVideoVendor(t, videoJSONReply(http.StatusOK, `{"total":0,"items":[]}`))
	engine := NewBytePlusVideoEngine("dreamina-seedance-2-5-260628", vendor.URL+"/api/v3", "ark-key", &lib.LoggerMock{}, nil)
	require.NoError(t, engine.ProbeVideoVendor(context.Background()))

	calls := vendor.Calls()
	require.Len(t, calls, 1)
	require.Equal(t, http.MethodGet, calls[0].Method)
	require.Equal(t, "/api/v3/contents/generations/tasks", calls[0].Path)
	require.Equal(t, "/api/v3/contents/generations/tasks?page_num=1&page_size=1", calls[0].RequestURI)
	require.Equal(t, "Bearer ark-key", calls[0].Auth)

	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests, http.StatusInternalServerError} {
		vendor := newVideoVendor(t, videoJSONReply(status, `{"error":{"code":"AuthenticationError","message":"the API key is invalid"}}`))
		engine := NewBytePlusVideoEngine("dreamina-seedance-2-5-260628", vendor.URL+"/api/v3", "bad", &lib.LoggerMock{}, nil)
		err := engine.ProbeVideoVendor(context.Background())
		var vendorErr *VideoVendorError
		require.ErrorAs(t, err, &vendorErr, "%d", status)
		require.Equal(t, status, vendorErr.StatusCode)
	}

	// A wrong apiUrl that answers 200 with something else is not BytePlus.
	page := newVideoVendor(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html>welcome</html>"))
	})
	engine = NewBytePlusVideoEngine("dreamina-seedance-2-5-260628", page.URL+"/api/v3", "ark-key", &lib.LoggerMock{}, nil)
	var vendorErr *VideoVendorError
	require.ErrorAs(t, engine.ProbeVideoVendor(context.Background()), &vendorErr)
	require.Equal(t, http.StatusOK, vendorErr.StatusCode)
}

// renderfulLike answers as the real Renderful does on its API path: 401
// without a key or with a wrong one, 404 with its JSON error for a working
// key and an unknown generation.
func renderfulLike(key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("Authorization") {
		case "":
			videoJSONReply(http.StatusUnauthorized, `{"error":"Unauthorized","message":"API key is required. Use Authorization: Bearer <api_key>"}`)(w, r)
		case "Bearer " + key:
			videoJSONReply(http.StatusNotFound, `{"error":"Not found","message":"Generation task not found"}`)(w, r)
		default:
			videoJSONReply(http.StatusUnauthorized, `{"error":"Unauthorized","message":"API key not found or disabled"}`)(w, r)
		}
	}
}

// The Renderful probe looks up a generation that does not exist: a working
// key gets the JSON 404, and the same URL without the key must get 401, which
// proves the path is behind Renderful's key check.
func TestRenderfulProbeVideoVendor(t *testing.T) {
	vendor := newVideoVendor(t, renderfulLike("rk-key"))
	engine := NewRenderfulEngine("seedance-2.5-t2v", vendor.URL+"/api/v1", "rk-key", &lib.LoggerMock{}, nil)
	require.NoError(t, engine.ProbeVideoVendor(context.Background()))
	calls := vendor.Calls()
	require.Len(t, calls, 2)
	for _, c := range calls {
		require.Equal(t, http.MethodGet, c.Method)
		require.Equal(t, "/api/v1/generations/"+renderfulProbeID, c.Path)
	}
	require.Equal(t, "Bearer rk-key", calls[0].Auth)
	require.Empty(t, calls[1].Auth, "the control request carries no key")

	// A wrong or disabled key is refused before the id is looked at.
	engine = NewRenderfulEngine("seedance-2.5-t2v", vendor.URL+"/api/v1", "sk_disabled", &lib.LoggerMock{}, nil)
	var vendorErr *VideoVendorError
	require.ErrorAs(t, engine.ProbeVideoVendor(context.Background()), &vendorErr)
	require.Equal(t, http.StatusUnauthorized, vendorErr.StatusCode)

	// Another JSON API that 404s unknown ids whatever the key: the control
	// request is not refused, so the probe fails.
	foreign := newVideoVendor(t, videoJSONReply(http.StatusNotFound, `{"error":"Not Found","message":"Route GET:/generations/x not found"}`))
	engine = NewRenderfulEngine("seedance-2.5-t2v", foreign.URL, "rk-key", &lib.LoggerMock{}, nil)
	require.ErrorAs(t, engine.ProbeVideoVendor(context.Background()), &vendorErr)
	require.Equal(t, http.StatusNotFound, vendorErr.StatusCode)
	require.Len(t, foreign.Calls(), 2)

	// Anything else is not proof of a working key: a wrong apiUrl that 404s
	// with a page (what api.renderful.ai answers off its API path), a
	// catch-all 200, and statuses Renderful was never seen to answer a
	// working key with.
	for name, reply := range map[string]http.HandlerFunc{
		"404 page": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("<html>not here</html>"))
		},
		"200 catch-all": videoJSONReply(http.StatusOK, `{"ok":true}`),
		"400":           videoJSONReply(http.StatusBadRequest, `{"error":"Bad request","message":"bad id"}`),
		"422":           videoJSONReply(http.StatusUnprocessableEntity, `{"error":"Unprocessable","message":"bad id"}`),
	} {
		vendor := newVideoVendor(t, reply)
		engine := NewRenderfulEngine("seedance-2.5-t2v", vendor.URL+"/api/v1", "rk-key", &lib.LoggerMock{}, nil)
		var vendorErr *VideoVendorError
		require.ErrorAs(t, engine.ProbeVideoVendor(context.Background()), &vendorErr, name)
	}

	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusPaymentRequired, http.StatusTooManyRequests, http.StatusBadGateway} {
		vendor := newVideoVendor(t, videoJSONReply(status, `{"error":"Unauthorized","message":"Invalid API key format"}`))
		engine := NewRenderfulEngine("seedance-2.5-t2v", vendor.URL+"/api/v1", "bad", &lib.LoggerMock{}, nil)
		err := engine.ProbeVideoVendor(context.Background())
		var vendorErr *VideoVendorError
		require.ErrorAs(t, err, &vendorErr, "%d", status)
		require.Equal(t, status, vendorErr.StatusCode)
	}
}

// A vendor that does not answer in time is unreachable, and the error says
// it timed out, without naming the vendor's host.
func TestVideoProbeTimeoutIsMarked(t *testing.T) {
	hang := newVideoVendor(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	})
	for _, engine := range []VideoJobEngine{
		NewBytePlusVideoEngine("dreamina-seedance-2-5-260628", hang.URL+"/api/v3", "k", &lib.LoggerMock{}, nil),
		NewRenderfulEngine("seedance-2.5-t2v", hang.URL+"/api/v1", "k", &lib.LoggerMock{}, nil),
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		err := engine.ProbeVideoVendor(ctx)
		cancel()
		require.ErrorIs(t, err, ErrVideoVendorUnreachable)
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.NotContains(t, err.Error(), strings.TrimPrefix(hang.URL, "http://"))
	}
}
