package aiengine

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"sync"
	"testing"
	"time"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/videojob"
	"github.com/stretchr/testify/require"
)

// videoVendorCall is one request a fake vendor received.
type videoVendorCall struct {
	Method      string
	Path        string
	RequestURI  string
	Auth        string
	Accept      string
	ContentType string
	Body        []byte
}

// videoVendor is an httptest vendor that records every request before
// handing it to handler.
type videoVendor struct {
	*httptest.Server
	mu    sync.Mutex
	calls []videoVendorCall
}

func newVideoVendor(t *testing.T, handler http.HandlerFunc) *videoVendor {
	t.Helper()
	v := &videoVendor{}
	v.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		v.mu.Lock()
		v.calls = append(v.calls, videoVendorCall{
			Method:      r.Method,
			Path:        r.URL.Path,
			RequestURI:  r.RequestURI,
			Auth:        r.Header.Get("Authorization"),
			Accept:      r.Header.Get("Accept"),
			ContentType: r.Header.Get("Content-Type"),
			Body:        body,
		})
		v.mu.Unlock()
		if handler != nil {
			handler(w, r)
		}
	}))
	t.Cleanup(v.Close)
	return v
}

func (v *videoVendor) Calls() []videoVendorCall {
	v.mu.Lock()
	defer v.mu.Unlock()
	return slices.Clone(v.calls)
}

func videoJSONReply(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func decodeVideoBody(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(raw, &m), string(raw))
	return m
}

func videoTestResolved() videojob.Resolved {
	return videojob.Resolved{
		Prompt:          "a harbour at dusk",
		DurationSeconds: 6,
		AspectRatio:     "9:16",
		Resolution:      "480p",
		GenerateAudio:   true,
	}
}

func videoBoolPtr(v bool) *bool { return &v }
func videoIntPtr(v int) *int    { return &v }

// pngBytes is enough of a PNG for DetectContentType.
var videoTestPNG = append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 16)...)

// --- Renderful -------------------------------------------------------------

func TestRenderfulCreateVideoJobTextToVideo(t *testing.T) {
	vendor := newVideoVendor(t, videoJSONReply(http.StatusOK, `{"id":"gen/1 2","status":"queued","poll_url":"/api/v1/generations/gen%2F1%202"}`))
	// A trailing slash on apiUrl is trimmed.
	engine := NewRenderfulEngine("seedance-2.5-t2v", vendor.URL+"/api/v1/", "rk-test", &lib.LoggerMock{}, nil)

	task, err := engine.CreateVideoJob(context.Background(), &VideoJobInput{Resolved: videoTestResolved(), Audio: engine.BaseVideoSpec().Audio})
	require.NoError(t, err)
	require.Equal(t, &VideoTask{
		Vendor:    videojob.VendorRenderful,
		TaskID:    "gen/1 2",
		StatusURL: vendor.URL + "/api/v1/generations/gen%2F1%202",
	}, task)

	calls := vendor.Calls()
	require.Len(t, calls, 1)
	call := calls[0]
	require.Equal(t, http.MethodPost, call.Method)
	require.Equal(t, "/api/v1/generations", call.Path)
	require.Equal(t, "Bearer rk-test", call.Auth)
	require.Equal(t, "application/json", call.ContentType)
	require.Equal(t, "application/json", call.Accept)
	require.Equal(t, map[string]any{
		"type":           "text-to-video",
		"model":          "seedance-2.5-t2v",
		"prompt":         "a harbour at dusk",
		"aspect_ratio":   "9:16",
		"resolution":     "480p",
		"duration":       float64(6),
		"generate_audio": true,
	}, decodeVideoBody(t, call.Body))
}

func TestRenderfulCreateVideoJobAudioSwitchOnlyForSeedance(t *testing.T) {
	vendor := newVideoVendor(t, videoJSONReply(http.StatusOK, `{"id":"g1"}`))

	// Seedance: the switch is always sent, false included.
	seedance := NewRenderfulEngine("seedance-2.5-t2v", vendor.URL, "k", &lib.LoggerMock{}, nil)
	res := videoTestResolved()
	res.GenerateAudio = false
	_, err := seedance.CreateVideoJob(context.Background(), &VideoJobInput{Resolved: res, Audio: seedance.BaseVideoSpec().Audio})
	require.NoError(t, err)

	// Other catalog models name the switch differently: never sent.
	kling := NewRenderfulEngine("kling-v2-master", vendor.URL, "k", &lib.LoggerMock{}, nil)
	_, err = kling.CreateVideoJob(context.Background(), &VideoJobInput{Resolved: videoTestResolved(), Audio: kling.BaseVideoSpec().Audio})
	require.NoError(t, err)

	calls := vendor.Calls()
	require.Len(t, calls, 2)
	first := decodeVideoBody(t, calls[0].Body)
	require.Contains(t, first, "generate_audio")
	require.Equal(t, false, first["generate_audio"])

	second := decodeVideoBody(t, calls[1].Body)
	require.NotContains(t, second, "generate_audio")
	require.Equal(t, "kling-v2-master", second["model"])
	require.Equal(t, "text-to-video", second["type"])
}

func TestRenderfulCreateVideoJobImageToVideo(t *testing.T) {
	const frame = "https://img.example.com/frame.png"

	cases := []struct {
		name      string
		modelName string
		wantModel string
	}{
		{name: "seedance t2v maps to its i2v sibling", modelName: "seedance-2.5-t2v", wantModel: "seedance-2.5-i2v"},
		{name: "any t2v maps to its i2v sibling", modelName: "kling-v2-t2v", wantModel: "kling-v2-i2v"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vendor := newVideoVendor(t, videoJSONReply(http.StatusCreated, `{"id":"g-img"}`))
			engine := NewRenderfulEngine(tc.modelName, vendor.URL, "k", &lib.LoggerMock{}, nil)
			require.True(t, engine.BaseVideoSpec().Image)

			res := videoTestResolved()
			res.HasImage = true
			task, err := engine.CreateVideoJob(context.Background(), &VideoJobInput{Resolved: res, ImageURL: frame})
			require.NoError(t, err)
			require.Equal(t, "g-img", task.TaskID)

			calls := vendor.Calls()
			require.Len(t, calls, 1)
			body := decodeVideoBody(t, calls[0].Body)
			require.Equal(t, "image-to-video", body["type"])
			require.Equal(t, tc.wantModel, body["model"])
			require.Equal(t, frame, body["image_url"])
		})
	}
}

func TestRenderfulCreateVideoJobRefusesFramesItCannotTake(t *testing.T) {
	vendor := newVideoVendor(t, videoJSONReply(http.StatusOK, `{"id":"never"}`))

	// Inline bytes: Renderful takes image_url only.
	engine := NewRenderfulEngine("seedance-2.5-t2v", vendor.URL, "k", &lib.LoggerMock{}, nil)
	_, err := engine.CreateVideoJob(context.Background(), &VideoJobInput{
		Resolved: videoTestResolved(),
		Image:    &videojob.Image{MimeType: "image/png", Data: videoTestPNG},
	})
	var vErr *VideoVendorError
	require.ErrorAs(t, err, &vErr)
	require.Equal(t, http.StatusBadRequest, vErr.StatusCode)
	require.NotEmpty(t, vErr.Message)

	// A model without an image-to-video sibling takes no first frame.
	noImage := NewRenderfulEngine("wan-2.2", vendor.URL, "k", &lib.LoggerMock{}, nil)
	require.False(t, noImage.BaseVideoSpec().Image)
	_, err = noImage.CreateVideoJob(context.Background(), &VideoJobInput{Resolved: videoTestResolved(), ImageURL: "https://img.example.com/a.png"})
	require.ErrorAs(t, err, &vErr)
	require.Equal(t, http.StatusBadRequest, vErr.StatusCode)

	require.Empty(t, vendor.Calls(), "a refused frame must not reach the vendor")

	// With both set, the URL wins and the inline bytes are not an error.
	_, err = engine.CreateVideoJob(context.Background(), &VideoJobInput{
		Resolved: videoTestResolved(),
		ImageURL: "https://img.example.com/a.png",
		Image:    &videojob.Image{MimeType: "image/png", Data: videoTestPNG},
	})
	require.NoError(t, err)
	calls := vendor.Calls()
	require.Len(t, calls, 1)
	require.Equal(t, "https://img.example.com/a.png", decodeVideoBody(t, calls[0].Body)["image_url"])
}

func TestRenderfulVendorErrors(t *testing.T) {
	cases := []struct {
		name        string
		status      int
		body        string
		wantCode    string
		wantMessage string
		wantText    string
	}{
		{
			name:        "402 with error and message",
			status:      http.StatusPaymentRequired,
			body:        `{"error":"insufficient_credits","message":"Not enough credits"}`,
			wantCode:    "insufficient_credits",
			wantMessage: "Not enough credits",
			wantText:    "video vendor refused the request (402 insufficient_credits): Not enough credits",
		},
		{
			name:        "401 with error only",
			status:      http.StatusUnauthorized,
			body:        `{"error":"Unauthorized"}`,
			wantMessage: "Unauthorized",
			wantText:    "video vendor refused the request (401): Unauthorized",
		},
		{
			name:        "422 with equal error and message",
			status:      http.StatusUnprocessableEntity,
			body:        `{"error":"bad duration","message":"bad duration"}`,
			wantMessage: "bad duration",
		},
		{
			name:        "503 JSON",
			status:      http.StatusServiceUnavailable,
			body:        `{"error":"overloaded","message":"try again later"}`,
			wantCode:    "overloaded",
			wantMessage: "try again later",
		},
		{
			name:     "500 non-JSON",
			status:   http.StatusInternalServerError,
			body:     `upstream exploded`,
			wantText: "video vendor refused the request (500): Internal Server Error",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vendor := newVideoVendor(t, videoJSONReply(tc.status, tc.body))
			engine := NewRenderfulEngine("seedance-2.5-t2v", vendor.URL, "k", &lib.LoggerMock{}, nil)

			for _, call := range []func() error{
				func() error {
					_, err := engine.CreateVideoJob(context.Background(), &VideoJobInput{Resolved: videoTestResolved()})
					return err
				},
				func() error {
					_, err := engine.FetchVideoStatus(context.Background(), "g1")
					return err
				},
			} {
				err := call()
				var vErr *VideoVendorError
				require.ErrorAs(t, err, &vErr)
				require.Equal(t, tc.status, vErr.StatusCode)
				require.Equal(t, tc.wantCode, vErr.Code)
				require.Equal(t, tc.wantMessage, vErr.Message)
				if tc.wantText != "" {
					require.Equal(t, tc.wantText, vErr.Error())
				}
			}
		})
	}
}

func TestRenderfulCreateVideoJobRequiresAnID(t *testing.T) {
	for _, body := range []string{`{}`, `{"id":""}`, `not json`} {
		vendor := newVideoVendor(t, videoJSONReply(http.StatusOK, body))
		engine := NewRenderfulEngine("seedance-2.5-t2v", vendor.URL, "k", &lib.LoggerMock{}, nil)
		task, err := engine.CreateVideoJob(context.Background(), &VideoJobInput{Resolved: videoTestResolved()})
		require.Error(t, err, body)
		require.Nil(t, task, body)
		var vErr *VideoVendorError
		require.False(t, errors.As(err, &vErr), "a 2xx without an id is not a vendor refusal: %s", body)
	}
}

func TestRenderfulFetchVideoStatus(t *testing.T) {
	const status = `{"id":"gen/1 2","status":"completed","outputs":["https://cdn.example.com/clip.mp4"]}`
	vendor := newVideoVendor(t, videoJSONReply(http.StatusOK, status))
	engine := NewRenderfulEngine("seedance-2.5-t2v", vendor.URL+"/api/v1", "rk-test", &lib.LoggerMock{}, nil)

	body, err := engine.FetchVideoStatus(context.Background(), "gen/1 2")
	require.NoError(t, err)
	require.JSONEq(t, status, string(body))

	calls := vendor.Calls()
	require.Len(t, calls, 1)
	require.Equal(t, http.MethodGet, calls[0].Method)
	// The id is one escaped path segment on the wire.
	require.Equal(t, "/api/v1/generations/gen%2F1%202", calls[0].RequestURI)
	require.Equal(t, "Bearer rk-test", calls[0].Auth)
	require.Empty(t, calls[0].Body)

	update, err := videojob.ParseStatus(videojob.VendorRenderful, body)
	require.NoError(t, err)
	require.Equal(t, videojob.StatusSucceeded, update.Status)
	require.Equal(t, "https://cdn.example.com/clip.mp4", update.VideoURL)
}

func TestRenderfulFetchVideoStatusNon2xx(t *testing.T) {
	vendor := newVideoVendor(t, videoJSONReply(http.StatusNotFound, `{"error":"not_found","message":"Generation not found"}`))
	engine := NewRenderfulEngine("seedance-2.5-t2v", vendor.URL, "k", &lib.LoggerMock{}, nil)

	body, err := engine.FetchVideoStatus(context.Background(), "missing")
	require.Nil(t, body)
	var vErr *VideoVendorError
	require.ErrorAs(t, err, &vErr)
	require.Equal(t, http.StatusNotFound, vErr.StatusCode)
	require.Equal(t, "not_found", vErr.Code)
	require.Equal(t, "Generation not found", vErr.Message)
}

func TestRenderfulDefaultBaseURL(t *testing.T) {
	engine := NewRenderfulEngine("seedance-2.5-t2v", "", "k", &lib.LoggerMock{}, nil)
	require.Equal(t, RENDERFUL_DEFAULT_BASE_URL, engine.baseURL)
	require.Equal(t, RENDERFUL_DEFAULT_BASE_URL+"/generations/abc", engine.statusURL("abc"))
	require.Equal(t, API_TYPE_RENDERFUL, engine.ApiType())
}

// --- BytePlus --------------------------------------------------------------

func TestBytePlusCreateVideoJobTextOnly(t *testing.T) {
	vendor := newVideoVendor(t, videoJSONReply(http.StatusOK, `{"id":"cgt-20261028-abc"}`))
	engine := NewBytePlusVideoEngine("dreamina-seedance-2-5-260628", vendor.URL+"/api/v3/", "ark-test", &lib.LoggerMock{}, nil)

	res := videojob.Resolved{Prompt: "a harbour at dusk", DurationSeconds: 8, AspectRatio: "16:9", Resolution: "1080p", GenerateAudio: true}
	task, err := engine.CreateVideoJob(context.Background(), &VideoJobInput{Resolved: res, Audio: engine.BaseVideoSpec().Audio})
	require.NoError(t, err)
	require.Equal(t, &VideoTask{
		Vendor:    videojob.VendorBytePlus,
		TaskID:    "cgt-20261028-abc",
		StatusURL: vendor.URL + "/api/v3/contents/generations/tasks/cgt-20261028-abc",
	}, task)

	calls := vendor.Calls()
	require.Len(t, calls, 1)
	call := calls[0]
	require.Equal(t, http.MethodPost, call.Method)
	require.Equal(t, "/api/v3/contents/generations/tasks", call.Path)
	require.Equal(t, "Bearer ark-test", call.Auth)
	require.Equal(t, "application/json", call.ContentType)
	require.Equal(t, map[string]any{
		"model":          "dreamina-seedance-2-5-260628",
		"content":        []any{map[string]any{"type": "text", "text": "a harbour at dusk"}},
		"ratio":          "16:9",
		"duration":       float64(8),
		"resolution":     "1080p",
		"generate_audio": true,
		"watermark":      false,
	}, decodeVideoBody(t, call.Body))
}

// The receiver passes the served spec's Audio; with the base spec that is the
// vendor's switch.
func TestBytePlusGenerateAudioOnlyForAudioModels(t *testing.T) {
	cases := []struct {
		model     string
		wantAudio bool
	}{
		{model: "dreamina-seedance-2-5-260628", wantAudio: true},
		{model: "dreamina-seedance-2-0-fast-260128", wantAudio: true},
		{model: "seedance-1-5-pro-251215", wantAudio: true},
		{model: "seedance-1-0-pro-250528", wantAudio: false},
		{model: "ep-20261028-custom", wantAudio: false},
	}
	for _, tc := range cases {
		t.Run(tc.model, func(t *testing.T) {
			vendor := newVideoVendor(t, videoJSONReply(http.StatusOK, `{"id":"cgt-1"}`))
			engine := NewBytePlusVideoEngine(tc.model, vendor.URL, "k", &lib.LoggerMock{}, nil)
			require.Equal(t, tc.wantAudio, engine.BaseVideoSpec().Audio)

			res := videoTestResolved()
			res.GenerateAudio = false
			_, err := engine.CreateVideoJob(context.Background(), &VideoJobInput{Resolved: res, Audio: engine.BaseVideoSpec().Audio})
			require.NoError(t, err)

			body := decodeVideoBody(t, vendor.Calls()[0].Body)
			if tc.wantAudio {
				require.Contains(t, body, "generate_audio")
				require.Equal(t, false, body["generate_audio"])
			} else {
				require.NotContains(t, body, "generate_audio")
			}
		})
	}
}

func TestBytePlusCreateVideoJobFirstFrame(t *testing.T) {
	inline := &videojob.Image{MimeType: "image/png", Data: videoTestPNG}
	inlineURI := "data:image/png;base64," + base64.StdEncoding.EncodeToString(videoTestPNG)

	cases := []struct {
		name      string
		model     string
		ratio     string
		imageURL  string
		image     *videojob.Image
		wantFrame string
		wantRatio string
	}{
		{
			name:      "url on seedance 2.5 forces adaptive",
			model:     "dreamina-seedance-2-5-260628",
			ratio:     "16:9",
			imageURL:  "https://img.example.com/frame.png",
			wantFrame: "https://img.example.com/frame.png",
			wantRatio: "adaptive",
		},
		{
			name:      "inline image on seedance 2.5 forces adaptive",
			model:     "dreamina-seedance-2-5-260628",
			ratio:     "9:16",
			image:     inline,
			wantFrame: inlineURI,
			wantRatio: "adaptive",
		},
		{
			name:      "inline image as data URI is adaptive on other models too",
			model:     "seedance-1-5-pro-251215",
			ratio:     "9:16",
			image:     inline,
			wantFrame: inlineURI,
			wantRatio: "adaptive",
		},
		{
			name:      "url wins over inline bytes",
			model:     "seedance-1-0-pro-250528",
			ratio:     "1:1",
			imageURL:  "https://img.example.com/b.webp",
			image:     inline,
			wantFrame: "https://img.example.com/b.webp",
			wantRatio: "adaptive",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vendor := newVideoVendor(t, videoJSONReply(http.StatusOK, `{"id":"cgt-img"}`))
			engine := NewBytePlusVideoEngine(tc.model, vendor.URL, "k", &lib.LoggerMock{}, nil)

			res := videoTestResolved()
			res.AspectRatio = tc.ratio
			res.HasImage = true
			_, err := engine.CreateVideoJob(context.Background(), &VideoJobInput{Resolved: res, ImageURL: tc.imageURL, Image: tc.image})
			require.NoError(t, err)

			calls := vendor.Calls()
			require.Len(t, calls, 1)
			body := decodeVideoBody(t, calls[0].Body)
			require.Equal(t, tc.wantRatio, body["ratio"])
			require.Equal(t, []any{
				map[string]any{"type": "text", "text": res.Prompt},
				map[string]any{
					"type":      "image_url",
					"image_url": map[string]any{"url": tc.wantFrame},
					"role":      "first_frame",
				},
			}, body["content"])
		})
	}
}

func TestBytePlusVendorErrors(t *testing.T) {
	cases := []struct {
		name        string
		status      int
		body        string
		wantCode    string
		wantMessage string
	}{
		{
			name:        "400 invalid parameter",
			status:      http.StatusBadRequest,
			body:        `{"error":{"code":"InvalidParameter","message":"The parameter ratio specified in the request is not valid","param":"ratio","type":"BadRequest"}}`,
			wantCode:    "InvalidParameter",
			wantMessage: "The parameter ratio specified in the request is not valid",
		},
		{
			name:        "429 rate limited",
			status:      http.StatusTooManyRequests,
			body:        `{"error":{"code":"RateLimitExceeded.EndpointRPMExceeded","message":"slow down","type":"TooManyRequests"}}`,
			wantCode:    "RateLimitExceeded.EndpointRPMExceeded",
			wantMessage: "slow down",
		},
		{
			name:   "500 non-JSON",
			status: http.StatusInternalServerError,
			body:   `<html>oops</html>`,
		},
		{
			name:   "502 JSON without an error object",
			status: http.StatusBadGateway,
			body:   `{"message":"bad gateway"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vendor := newVideoVendor(t, videoJSONReply(tc.status, tc.body))
			engine := NewBytePlusVideoEngine("dreamina-seedance-2-5-260628", vendor.URL, "k", &lib.LoggerMock{}, nil)

			_, err := engine.CreateVideoJob(context.Background(), &VideoJobInput{Resolved: videoTestResolved()})
			var vErr *VideoVendorError
			require.ErrorAs(t, err, &vErr)
			require.Equal(t, tc.status, vErr.StatusCode)
			require.Equal(t, tc.wantCode, vErr.Code)
			require.Equal(t, tc.wantMessage, vErr.Message)
			if tc.wantCode != "" {
				require.Contains(t, vErr.Error(), tc.wantCode)
				require.Contains(t, vErr.Error(), tc.wantMessage)
			} else {
				require.Contains(t, vErr.Error(), http.StatusText(tc.status))
			}

			_, err = engine.FetchVideoStatus(context.Background(), "cgt-1")
			require.ErrorAs(t, err, &vErr)
			require.Equal(t, tc.status, vErr.StatusCode)
			require.Equal(t, tc.wantCode, vErr.Code)
		})
	}
}

func TestBytePlusCreateVideoJobRequiresAnID(t *testing.T) {
	for _, body := range []string{`{}`, `{"id":""}`, `[]`} {
		vendor := newVideoVendor(t, videoJSONReply(http.StatusOK, body))
		engine := NewBytePlusVideoEngine("dreamina-seedance-2-5-260628", vendor.URL, "k", &lib.LoggerMock{}, nil)
		task, err := engine.CreateVideoJob(context.Background(), &VideoJobInput{Resolved: videoTestResolved()})
		require.Error(t, err, body)
		require.Nil(t, task, body)
		var vErr *VideoVendorError
		require.False(t, errors.As(err, &vErr), body)
	}
}

func TestBytePlusFetchVideoStatus(t *testing.T) {
	const status = `{"id":"cgt-1","model":"dreamina-seedance-2-5-260628","status":"succeeded","content":{"video_url":"https://cdn.example.com/v.mp4"},"updated_at":1790000000}`
	vendor := newVideoVendor(t, videoJSONReply(http.StatusOK, status))
	engine := NewBytePlusVideoEngine("dreamina-seedance-2-5-260628", vendor.URL+"/api/v3", "ark-test", &lib.LoggerMock{}, nil)

	body, err := engine.FetchVideoStatus(context.Background(), "cgt-1")
	require.NoError(t, err)
	require.JSONEq(t, status, string(body))

	calls := vendor.Calls()
	require.Len(t, calls, 1)
	require.Equal(t, http.MethodGet, calls[0].Method)
	require.Equal(t, "/api/v3/contents/generations/tasks/cgt-1", calls[0].RequestURI)
	require.Equal(t, "Bearer ark-test", calls[0].Auth)

	update, err := videojob.ParseStatus(videojob.VendorBytePlus, body)
	require.NoError(t, err)
	require.Equal(t, videojob.StatusSucceeded, update.Status)
	require.Equal(t, "https://cdn.example.com/v.mp4", update.VideoURL)
}

func TestBytePlusDefaultBaseURL(t *testing.T) {
	engine := NewBytePlusVideoEngine("dreamina-seedance-2-5-260628", "", "k", &lib.LoggerMock{}, nil)
	require.Equal(t, BYTEPLUS_VIDEO_DEFAULT_BASE_URL, engine.baseURL)
	require.Equal(t, BYTEPLUS_VIDEO_DEFAULT_BASE_URL+"/contents/generations/tasks/a%2Fb", engine.statusURL("a/b"))
	require.Equal(t, API_TYPE_BYTEPLUS_VIDEO, engine.ApiType())
}

// --- base specs ------------------------------------------------------------

func TestBytePlusBaseVideoSpecPerModelFamily(t *testing.T) {
	sixRatios := []string{"16:9", "4:3", "1:1", "3:4", "9:16", "21:9"}
	allRatios := append(slices.Clone(sixRatios), "adaptive")
	sd := []string{"480p", "720p"}
	hd := []string{"480p", "720p", "1080p"}

	cases := []struct {
		model       string
		min, max    int
		ratios      []string
		resolutions []string
		audio       bool
	}{
		{model: "dreamina-seedance-2-5-260628", min: 4, max: 30, ratios: allRatios, resolutions: hd, audio: true},
		{model: "dreamina-seedance-2-0-260128", min: 4, max: 15, ratios: allRatios, resolutions: hd, audio: true},
		{model: "dreamina-seedance-2-0-fast-260128", min: 4, max: 15, ratios: allRatios, resolutions: sd, audio: true},
		{model: "dreamina-seedance-2-0-mini-260128", min: 4, max: 15, ratios: allRatios, resolutions: sd, audio: true},
		{model: "seedance-1-5-pro-251215", min: 4, max: 12, ratios: allRatios, resolutions: hd, audio: true},
		{model: "seedance-1-0-pro-250528", min: 2, max: 12, ratios: sixRatios, resolutions: hd, audio: false},
		{model: "seedance-1-0-pro-fast-251015", min: 2, max: 12, ratios: sixRatios, resolutions: hd, audio: false},
		// custom endpoint id / unknown model: the narrow common subset
		{model: "ep-20261028-custom", min: 4, max: 12, ratios: sixRatios, resolutions: sd, audio: false},
		{model: "", min: 4, max: 12, ratios: sixRatios, resolutions: sd, audio: false},
	}
	for _, tc := range cases {
		t.Run(tc.model, func(t *testing.T) {
			spec := NewBytePlusVideoEngine(tc.model, "", "", &lib.LoggerMock{}, nil).BaseVideoSpec()
			require.Equal(t, tc.min, spec.MinDuration)
			require.Equal(t, tc.max, spec.MaxDuration)
			require.Equal(t, tc.ratios, spec.AspectRatios)
			require.Equal(t, tc.resolutions, spec.Resolutions)
			require.Equal(t, tc.audio, spec.Audio)
			require.True(t, spec.Image)
			require.True(t, spec.ImageInline)
			require.Equal(t, videojob.Defaults{DurationSeconds: 5, AspectRatio: "16:9", Resolution: "720p"}, spec.Defaults)
			require.Equal(t, defaultVideoQuote(), spec.Quote)
			require.NoError(t, spec.Validate())
		})
	}
}

func TestBytePlusBaseVideoSpecIsFresh(t *testing.T) {
	engine := NewBytePlusVideoEngine("dreamina-seedance-2-5-260628", "", "", &lib.LoggerMock{}, nil)
	a := engine.BaseVideoSpec()
	a.AspectRatios[0] = "mutated"
	a.Quote.ResolutionWeights["720p"] = 99
	b := engine.BaseVideoSpec()
	require.Equal(t, "16:9", b.AspectRatios[0])
	require.Equal(t, 2.25, b.Quote.ResolutionWeights["720p"])
}

func TestRenderfulBaseVideoSpec(t *testing.T) {
	cases := []struct {
		name  string
		model string
		audio bool
		image bool
	}{
		{name: "seedance t2v", model: "seedance-2.5-t2v", audio: true, image: true},
		{name: "seedance without sibling", model: "seedance-2.5", audio: true, image: false},
		{name: "kling", model: "kling-v2-master", audio: false, image: false},
		{name: "kling t2v", model: "kling-v2-t2v", audio: false, image: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec := NewRenderfulEngine(tc.model, "", "", &lib.LoggerMock{}, nil).BaseVideoSpec()
			require.Equal(t, 4, spec.MinDuration)
			require.Equal(t, 30, spec.MaxDuration)
			require.Equal(t, []string{"480p", "720p"}, spec.Resolutions)
			require.Contains(t, spec.AspectRatios, "adaptive")
			require.Equal(t, tc.audio, spec.Audio)
			require.Equal(t, tc.image, spec.Image)
			require.False(t, spec.ImageInline, "renderful takes image_url only")
			require.False(t, spec.Defaults.GenerateAudio)
			require.Equal(t, defaultVideoQuote(), spec.Quote)
			require.NoError(t, spec.Validate())
		})
	}
}

func TestVideoSpecForIsBaseSpec(t *testing.T) {
	cfg := &config.ModelConfig{ModelName: "dreamina-seedance-2-5-260628", ApiType: API_TYPE_BYTEPLUS_VIDEO}
	engine := NewBytePlusVideoEngine(cfg.ModelName, "", "", &lib.LoggerMock{}, nil)

	spec, err := VideoSpecFor(cfg, engine)
	require.NoError(t, err)
	want := engine.BaseVideoSpec()
	want.ModelName = cfg.ModelName
	require.Equal(t, want, spec)
}

func TestNewVideoJobEngineRejectsNonVideoModels(t *testing.T) {
	_, err := NewVideoJobEngine(nil, &lib.LoggerMock{}, nil)
	require.Error(t, err)

	_, err = NewVideoJobEngine(&config.ModelConfig{ModelName: "m"}, &lib.LoggerMock{}, nil)
	require.Error(t, err)

	for _, apiType := range []string{API_TYPE_OPENAI, API_TYPE_CLAUDEAI, API_TYPE_PRODIA_V2, API_TYPE_DECISIONS, "byteplus", "bogus"} {
		engine, err := NewVideoJobEngine(&config.ModelConfig{ModelName: "m", ApiType: apiType, ApiURL: "https://h/v1"}, &lib.LoggerMock{}, nil)
		require.Error(t, err, apiType)
		require.Nil(t, engine, apiType)
		require.False(t, IsVideoApiType(apiType), apiType)

		_, err = VideoGuideFor(&config.ModelConfig{ModelName: "m", ApiType: apiType}, &lib.LoggerMock{})
		require.Error(t, err, apiType)
	}
	require.True(t, IsVideoApiType(API_TYPE_RENDERFUL))
	require.True(t, IsVideoApiType(API_TYPE_BYTEPLUS_VIDEO))
}

func TestNewVideoJobEngineBuildsTheAdapter(t *testing.T) {
	client := &http.Client{Timeout: time.Second}

	// A "video" key left in an old models-config is ignored.
	raw := `{"modelName":"seedance-2.5-t2v","apiType":"renderful","apiUrl":"https://renderful.example/api/v1/","apiKey":"rk",
		"video":{"maxDurationSeconds":10}}`
	var cfg config.ModelConfig
	require.NoError(t, json.Unmarshal([]byte(raw), &cfg))

	engine, err := NewVideoJobEngine(&cfg, &lib.LoggerMock{}, client)
	require.NoError(t, err)
	rf, ok := engine.(*Renderful)
	require.True(t, ok)
	require.Equal(t, "seedance-2.5-i2v", rf.imageModel)
	require.Equal(t, "https://renderful.example/api/v1", rf.baseURL)
	require.Equal(t, "rk", rf.apiKey)
	require.Same(t, client, rf.client)

	spec, err := VideoSpecFor(&cfg, engine)
	require.NoError(t, err)
	require.Equal(t, 30, spec.MaxDuration)
	require.Equal(t, defaultVideoQuote(), spec.Quote)

	bpCfg := &config.ModelConfig{ModelName: "dreamina-seedance-2-5-260628", ApiType: API_TYPE_BYTEPLUS_VIDEO, ApiKey: "ark"}
	engine, err = NewVideoJobEngine(bpCfg, &lib.LoggerMock{}, nil)
	require.NoError(t, err)
	bp, ok := engine.(*BytePlusVideo)
	require.True(t, ok)
	require.Equal(t, BYTEPLUS_VIDEO_DEFAULT_BASE_URL, bp.baseURL)
	require.Equal(t, "ark", bp.apiKey)
	require.NotNil(t, bp.client)
	require.NotNil(t, bp.client.CheckRedirect, "the default client guards the key against cross-host redirects")
}

func TestVideoGuideForResolves(t *testing.T) {
	cfg := &config.ModelConfig{
		ModelName: "seedance-2.5-t2v",
		ApiType:   API_TYPE_RENDERFUL,
		ApiURL:    "https://api.renderful.ai/api/v1",
	}
	guide, err := VideoGuideFor(cfg, &lib.LoggerMock{})
	require.NoError(t, err)
	require.Equal(t, videojob.ObjectGuide, guide.Object)
	require.Equal(t, "seedance-2.5-t2v", guide.ModelName)
	require.NotEmpty(t, guide.Summary)
	require.NotEmpty(t, guide.Instructions)

	names := make([]string, 0, len(guide.Fields))
	for _, f := range guide.Fields {
		names = append(names, f.Name)
	}
	require.Equal(t, []string{
		videojob.FieldPrompt, videojob.FieldDurationSeconds, videojob.FieldAspectRatio,
		videojob.FieldResolution, videojob.FieldGenerateAudio, videojob.FieldImageURL,
	}, names, "renderful seedance offers audio and image_url, not image_b64")

	resolved, err := guide.Resolve(&videojob.CreateRequest{Prompt: " a fox in snow "})
	require.NoError(t, err)
	require.Equal(t, videojob.Resolved{Prompt: "a fox in snow", DurationSeconds: 5, AspectRatio: "16:9", Resolution: "720p"}, resolved)

	// 5 s × 2.25 (720p) × 75 = 843.75 → 844 s.
	_, secs, err := guide.SessionSeconds(&videojob.CreateRequest{Prompt: "x"})
	require.NoError(t, err)
	require.Equal(t, int64(844), secs)

	resolved, err = guide.Resolve(&videojob.CreateRequest{Prompt: "x", DurationSeconds: videoIntPtr(30), Resolution: "480P", ImageURL: "https://img.example.com/a.png", GenerateAudio: videoBoolPtr(true)})
	require.NoError(t, err)
	require.Equal(t, videojob.Resolved{Prompt: "x", DurationSeconds: 30, AspectRatio: "16:9", Resolution: "480p", GenerateAudio: true, HasImage: true}, resolved)

	_, err = guide.Resolve(&videojob.CreateRequest{Prompt: "x", ImageB64: base64.StdEncoding.EncodeToString(videoTestPNG)})
	require.ErrorIs(t, err, videojob.ErrInvalidRequest)
	_, err = guide.Resolve(&videojob.CreateRequest{Prompt: "x", DurationSeconds: videoIntPtr(31)})
	require.ErrorIs(t, err, videojob.ErrInvalidRequest)
	_, err = guide.Resolve(&videojob.CreateRequest{Prompt: "x", Resolution: "1080p"})
	require.ErrorIs(t, err, videojob.ErrInvalidRequest)

	// The guide survives the wire unchanged.
	raw, err := json.Marshal(guide)
	require.NoError(t, err)
	var back videojob.Guide
	require.NoError(t, json.Unmarshal(raw, &back))
	again, err := back.Resolve(&videojob.CreateRequest{Prompt: " a fox in snow "})
	require.NoError(t, err)
	require.Equal(t, videojob.Resolved{Prompt: "a fox in snow", DurationSeconds: 5, AspectRatio: "16:9", Resolution: "720p"}, again)
}

// --- factory ---------------------------------------------------------------

func TestApiAdapterFactoryVideoAdapters(t *testing.T) {
	log := lib.NewTestLogger()
	ctx := context.Background()
	for _, apiType := range []string{API_TYPE_RENDERFUL, API_TYPE_BYTEPLUS_VIDEO} {
		engine, ok := ApiAdapterFactory(apiType, "m", "http://h/v1", "", nil, time.Second, log, nil)
		require.True(t, ok, apiType)
		require.NotNil(t, engine, apiType)
		require.Equal(t, apiType, engine.ApiType(), apiType)
		_, isVideo := engine.(VideoJobEngine)
		require.True(t, isVideo, apiType)

		called := false
		cb := func(ctx context.Context, completion gcs.Chunk, errResp *gcs.AiEngineErrorResponse) error {
			called = true
			return nil
		}
		require.ErrorIs(t, engine.Prompt(ctx, &gcs.OpenAICompletionRequestExtra{}, cb), ErrVideoNotSupported, apiType)
		require.ErrorIs(t, engine.AudioTranscription(ctx, &gcs.AudioTranscriptionRequest{}, cb), ErrVideoNotSupported, apiType)
		require.ErrorIs(t, engine.AudioSpeech(ctx, &gcs.AudioSpeechRequest{}, cb), ErrVideoNotSupported, apiType)
		require.ErrorIs(t, engine.Embeddings(ctx, &gcs.EmbeddingsRequest{}, cb), ErrVideoNotSupported, apiType)
		require.ErrorIs(t, engine.Decisions(ctx, &gcs.DecisionsRequest{}, cb), ErrVideoNotSupported, apiType)
		require.False(t, called, apiType)
	}
}

// --- redirects -------------------------------------------------------------

func TestVideoAdaptersDoNotFollowCrossHostRedirects(t *testing.T) {
	ctx := context.Background()
	elsewhere := newVideoVendor(t, videoJSONReply(http.StatusOK, `{"id":"stolen","status":"queued"}`))

	for _, code := range []int{http.StatusFound, http.StatusTemporaryRedirect, http.StatusPermanentRedirect} {
		vendor := newVideoVendor(t, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, elsewhere.URL+"/steal", code)
		})

		engines := []VideoJobEngine{
			NewRenderfulEngine("seedance-2.5-t2v", vendor.URL, "rk-secret", &lib.LoggerMock{}, nil),
			NewBytePlusVideoEngine("dreamina-seedance-2-5-260628", vendor.URL, "ark-secret", &lib.LoggerMock{}, nil),
		}
		for _, engine := range engines {
			task, err := engine.CreateVideoJob(ctx, &VideoJobInput{Resolved: videoTestResolved()})
			require.Nil(t, task)
			var vErr *VideoVendorError
			require.ErrorAs(t, err, &vErr)
			require.Equal(t, code, vErr.StatusCode)

			body, err := engine.FetchVideoStatus(ctx, "task-1")
			require.Nil(t, body)
			require.ErrorAs(t, err, &vErr)
			require.Equal(t, code, vErr.StatusCode)
		}
		require.Len(t, vendor.Calls(), 4)
	}
	require.Empty(t, elsewhere.Calls(), "no request, with or without the key, may reach another host")
}

// Control: an unguarded client does carry the key to another port of the same
// IP, which is what the guard prevents.
func TestVideoRedirectGuardIsWhatStopsTheKey(t *testing.T) {
	elsewhere := newVideoVendor(t, videoJSONReply(http.StatusOK, `{"id":"g1","status":"queued"}`))
	vendor := newVideoVendor(t, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, elsewhere.URL+"/steal", http.StatusFound)
	})

	engine := NewRenderfulEngine("seedance-2.5-t2v", vendor.URL, "rk-secret", &lib.LoggerMock{}, &http.Client{})
	_, err := engine.FetchVideoStatus(context.Background(), "g1")
	require.NoError(t, err)
	calls := elsewhere.Calls()
	require.Len(t, calls, 1)
	require.Equal(t, "Bearer rk-secret", calls[0].Auth)
}

func TestVideoAdaptersFollowSameHostRedirects(t *testing.T) {
	const status = `{"id":"g1","status":"processing"}`
	vendor := newVideoVendor(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/generations/g1" {
			http.Redirect(w, r, "/api/v1/generations/g1-moved", http.StatusFound)
			return
		}
		videoJSONReply(http.StatusOK, status)(w, r)
	})

	engine := NewRenderfulEngine("seedance-2.5-t2v", vendor.URL+"/api/v1", "rk-test", &lib.LoggerMock{}, nil)
	body, err := engine.FetchVideoStatus(context.Background(), "g1")
	require.NoError(t, err)
	require.JSONEq(t, status, string(body))

	calls := vendor.Calls()
	require.Len(t, calls, 2)
	require.Equal(t, "/api/v1/generations/g1-moved", calls[1].Path)
	require.Equal(t, "Bearer rk-test", calls[1].Auth)
}

func TestRefuseCrossHostRedirect(t *testing.T) {
	mk := func(raw string) *http.Request {
		u, err := url.Parse(raw)
		require.NoError(t, err)
		return &http.Request{URL: u}
	}
	via := func(raws ...string) []*http.Request {
		out := make([]*http.Request, 0, len(raws))
		for _, r := range raws {
			out = append(out, mk(r))
		}
		return out
	}

	require.NoError(t, refuseCrossHostRedirect(mk("https://api.example.com/b"), via("https://api.example.com/a")))
	require.NoError(t, refuseCrossHostRedirect(mk("https://api.example.com/b"), via("http://api.example.com/a")), "upgrade to https is fine")

	for _, next := range []string{
		"https://evil.example.com/b",
		"https://api.example.com:8443/b",
		"https://sub.api.example.com/b",
		"http://api.example.com/b",
	} {
		require.ErrorIs(t, refuseCrossHostRedirect(mk(next), via("https://api.example.com/a")), http.ErrUseLastResponse, next)
	}

	// Each hop is checked against the previous one.
	require.ErrorIs(t, refuseCrossHostRedirect(mk("https://evil.example.com/c"), via("https://api.example.com/a", "https://api.example.com/b")), http.ErrUseLastResponse)

	ten := make([]string, 10)
	for i := range ten {
		ten[i] = "https://api.example.com/a"
	}
	err := refuseCrossHostRedirect(mk("https://api.example.com/b"), via(ten...))
	require.Error(t, err)
	require.NotErrorIs(t, err, http.ErrUseLastResponse)

	injected := &http.Client{}
	require.Same(t, injected, videoHTTPClient(injected))
	def := videoHTTPClient(nil)
	require.NotNil(t, def)
	require.NotNil(t, def.CheckRedirect)
}
