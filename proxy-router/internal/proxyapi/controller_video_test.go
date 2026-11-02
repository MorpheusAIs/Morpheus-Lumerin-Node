package proxyapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/videojob"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// vjServe calls one handler directly (no routes, so no auth middleware).
func vjServe(t *testing.T, handler gin.HandlerFunc, method, target, body string, header map[string]string, params ...gin.Param) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	ctx.Request = httptest.NewRequest(method, target, reader)
	if body != "" {
		ctx.Request.Header.Set("Content-Type", "application/json")
	}
	for k, v := range header {
		ctx.Request.Header.Set(k, v)
	}
	ctx.Params = params

	handler(ctx)
	return w
}

func vjController(svc *VideoJobService) *ProxyController {
	return &ProxyController{videoJobs: svc, log: &lib.LoggerMock{}}
}

func vjErrorBody(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Error string `json:"error"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body), "body: %s", w.Body.String())
	return body.Error
}

func TestCreateVideoJobHandlerDisabled(t *testing.T) {
	c := vjController(nil)
	for name, h := range map[string]gin.HandlerFunc{
		"create": c.CreateVideoJob,
		"list":   c.ListVideoJobs,
		"get":    c.GetVideoJob,
		"quote":  c.VideoQuote,
	} {
		t.Run(name, func(t *testing.T) {
			w := vjServe(t, h, http.MethodPost, "/v1/video/x", `{"prompt":"x"}`, nil)
			require.Equal(t, http.StatusServiceUnavailable, w.Code)
			require.Equal(t, ErrVideoJobsDisabled.Error(), vjErrorBody(t, w))
		})
	}
}

func TestCreateVideoJobHandler(t *testing.T) {
	sessions := newVJSessions()
	sender := &vjSender{}
	svc := newVJService(t, sender, sessions, nil, 0)
	c := vjController(svc)

	body := `{"prompt":"a lighthouse at night","duration_seconds":5}`
	w := vjServe(t, c.CreateVideoJob, http.MethodPost, "/v1/video/jobs", body, map[string]string{"X-Request-Id": "req-vj-1", "session_id": vjSessionID.Hex()})
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	require.Equal(t, "req-vj-1", w.Header().Get("X-Request-Id"))

	var job videojob.Job
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &job))
	require.True(t, strings.HasPrefix(job.ID, "vjob_"))
	require.Equal(t, videojob.ObjectJob, job.Object)
	require.Equal(t, videojob.StatusQueued, job.Status)
	require.Equal(t, vjModelID.Hex(), job.Model)
	require.Equal(t, vjSessionID.Hex(), job.SessionID)

	creates := sender.createCalls()
	require.Len(t, creates, 1)
	require.Equal(t, vjSessionID, creates[0].sessionID)
	require.Equal(t, "a lighthouse at night", creates[0].req.Prompt)

	// GET returns the same job; the list has it.
	w = vjServe(t, c.GetVideoJob, http.MethodGet, "/v1/video/jobs/"+job.ID, "", nil, gin.Param{Key: "id", Value: job.ID})
	require.Equal(t, http.StatusOK, w.Code)
	var got videojob.Job
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Equal(t, job.ID, got.ID)

	w = vjServe(t, c.ListVideoJobs, http.MethodGet, "/v1/video/jobs", "", nil)
	require.Equal(t, http.StatusOK, w.Code)
	var list VideoJobListRes
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Equal(t, "list", list.Object)
	require.Len(t, list.Data, 1)
	require.Equal(t, job.ID, list.Data[0].ID)
}

func TestCreateVideoJobHandlerSessionHeader(t *testing.T) {
	sessions := newVJSessions()
	sessions.session = &structs.Session{ModelAgentId: vjModelID.Hex()}
	sender := &vjSender{}
	svc := newVJService(t, sender, sessions, nil, 0)
	c := vjController(svc)

	w := vjServe(t, c.CreateVideoJob, http.MethodPost, "/v1/video/jobs", `{"prompt":"x"}`, map[string]string{"session_id": vjUserSessionID.Hex()})
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	var job videojob.Job
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &job))
	require.Equal(t, vjUserSessionID.Hex(), job.SessionID)
	require.Equal(t, vjModelID.Hex(), job.Model)
	require.NotEmpty(t, w.Header().Get("X-Request-Id"), "a request id is generated when none is sent")
	require.Equal(t, vjUserSessionID, sender.createCalls()[0].sessionID)
}

var vjSessionHeader = map[string]string{"session_id": vjSessionID.Hex()}

// A provider that cannot be reached does not fail the submit: the job is
// created unconfirmed and sent again by the watcher.
func TestCreateVideoJobHandlerProviderUnreachableKeepsJob(t *testing.T) {
	sender := &vjSender{createFn: func(int, common.Hash, string) (*videojob.Receipt, error) {
		return nil, lib.WrapError(ErrProvider, errors.New("connection refused"))
	}}
	svc := newVJService(t, sender, newVJSessions(), nil, 0)
	c := vjController(svc)

	w := vjServe(t, c.CreateVideoJob, http.MethodPost, "/v1/video/jobs", `{"prompt":"x"}`, vjSessionHeader)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var job videojob.Job
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &job))
	require.Equal(t, videojob.StatusQueued, job.Status)
	require.Empty(t, job.VendorTaskID)
	require.Len(t, svc.List(), 1)
}

func TestCreateVideoJobHandlerErrors(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		header   map[string]string
		setup    func(*vjSessions, *vjSender)
		wantCode int
		wantErr  string
	}{
		{name: "malformed body", body: `{"prompt":`, wantCode: http.StatusBadRequest},
		{name: "invalid session header", body: `{"prompt":"x"}`, header: map[string]string{"session_id": "0xnothex"}, wantCode: http.StatusBadRequest},
		{name: "missing prompt", body: `{}`, header: vjSessionHeader, wantCode: http.StatusBadRequest, wantErr: "prompt is required"},
		{name: "no session", body: fmt.Sprintf(`{"model":%q,"prompt":"x"}`, vjModelID.Hex()), wantCode: http.StatusBadRequest, wantErr: "session_id is required"},
		{
			name: "provider refused", body: `{"prompt":"x"}`, header: vjSessionHeader,
			setup: func(_ *vjSessions, s *vjSender) {
				s.createFn = func(int, common.Hash, string) (*videojob.Receipt, error) {
					return nil, fmt.Errorf("%w: content policy", ErrProviderRefused)
				}
			},
			wantCode: http.StatusUnprocessableEntity, wantErr: "content policy",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sessions := newVJSessions()
			sender := &vjSender{}
			if tc.setup != nil {
				tc.setup(sessions, sender)
			}
			svc := newVJService(t, sender, sessions, nil, 0)
			c := vjController(svc)

			w := vjServe(t, c.CreateVideoJob, http.MethodPost, "/v1/video/jobs", tc.body, tc.header)
			require.Equal(t, tc.wantCode, w.Code, w.Body.String())
			msg := vjErrorBody(t, w)
			require.NotEmpty(t, msg)
			if tc.wantErr != "" {
				require.Contains(t, msg, tc.wantErr)
			}
			require.Empty(t, svc.List(), "a failed submit leaves no job")
		})
	}
}

func TestVideoJobHandlersNotFound(t *testing.T) {
	svc := newVJService(t, &vjSender{}, newVJSessions(), nil, 0)
	c := vjController(svc)
	param := gin.Param{Key: "id", Value: "vjob_missing"}

	w := vjServe(t, c.GetVideoJob, http.MethodGet, "/v1/video/jobs/vjob_missing", "", nil, param)
	require.Equal(t, http.StatusNotFound, w.Code)
	require.Equal(t, ErrVideoJobNotFound.Error(), vjErrorBody(t, w))
}

func TestVideoQuoteHandler(t *testing.T) {
	spec := vjSpec(10, 3600)
	sessions := newVJSessions()
	sessions.guides = []structs.VideoProviderGuide{
		{Provider: common.HexToAddress("0xaa"), BidID: common.HexToHash("0xa1"), PricePerSecondWei: "1000", Guide: spec.Guide()},
	}
	svc := newVJService(t, &vjSender{}, sessions, nil, 0)
	c := vjController(svc)

	w := vjServe(t, c.VideoQuote, http.MethodPost, "/v1/video/quote", `{"prompt":"x"}`, nil)
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Equal(t, "model is required", vjErrorBody(t, w))

	w = vjServe(t, c.VideoQuote, http.MethodPost, "/v1/video/quote", `{"prompt":"x","model":"bad"}`, nil)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = vjServe(t, c.VideoQuote, http.MethodPost, "/v1/video/quote", `{"prompt":`, nil)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = vjServe(t, c.VideoQuote, http.MethodPost, "/v1/video/quote", `{"prompt":"x","duration_seconds":6,"resolution":"1080p"}`, map[string]string{"model_id": vjModelID.Hex()})
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var res VideoQuoteRes
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	require.Equal(t, "list", res.Object)
	require.Len(t, res.Data, 1)
	require.Equal(t, int64(120), res.Data[0].SessionSeconds)
	require.Equal(t, "120000", res.Data[0].SessionCostWei)

	sessions.mu.Lock()
	sessions.guidesErr = errors.New("chain unavailable")
	sessions.mu.Unlock()
	w = vjServe(t, c.VideoQuote, http.MethodPost, "/v1/video/quote", fmt.Sprintf(`{"prompt":"x","model":%q}`, vjModelID.Hex()), nil)
	require.Equal(t, http.StatusBadGateway, w.Code)
}
