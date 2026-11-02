package proxyapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/videojob"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// Helpers in this file carry a vj prefix so they do not collide with other
// test files of the package.

const (
	vjPollInterval = 20 * time.Millisecond
	vjWaitFor      = 2 * time.Second
	vjTick         = 5 * time.Millisecond
)

var (
	vjModelID       = common.HexToHash("0x000000000000000000000000000000000000000000000000000000000000a11c")
	vjSessionID     = common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000005e55")
	vjUserSessionID = common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000b0b")
)

// --- chain stub -------------------------------------------------------------

type vjSessions struct {
	mu sync.Mutex

	session    *structs.Session
	sessionErr error
	gets       []common.Hash

	boundsMin, boundsMax int64

	guides    []structs.VideoProviderGuide
	guidesErr error
	guideReqs []common.Hash
}

// newVJSessions has the caller's session vjSessionID open on vjModelID.
func newVJSessions() *vjSessions {
	return &vjSessions{session: &structs.Session{Id: vjSessionID.Hex(), ModelAgentId: vjModelID.Hex()}}
}

func (s *vjSessions) VideoGuides(_ context.Context, modelID common.Hash) ([]structs.VideoProviderGuide, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.guideReqs = append(s.guideReqs, modelID)
	if s.guidesErr != nil {
		return nil, s.guidesErr
	}
	return slices.Clone(s.guides), nil
}

// VideoSessionBounds returns boundsMin/boundsMax; zero means no bound.
func (s *vjSessions) VideoSessionBounds(context.Context) (int64, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.boundsMin, s.boundsMax, nil
}

func (s *vjSessions) GetSession(_ context.Context, sessionID common.Hash) (*structs.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gets = append(s.gets, sessionID)
	if s.sessionErr != nil {
		return nil, s.sessionErr
	}
	if s.session == nil {
		return nil, errors.New("session not found")
	}
	cp := *s.session
	return &cp, nil
}

func (s *vjSessions) getCalls() []common.Hash {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.gets)
}

// --- provider stub ----------------------------------------------------------

type vjCreateCall struct {
	sessionID common.Hash
	jobID     string
	req       videojob.CreateRequest
}

type vjStatusCall struct {
	sessionID common.Hash
	jobID     string
	at        time.Time
}

// vjStatusFunc answers the n-th (1-based) status call.
type vjStatusFunc func(ctx context.Context, n int, sessionID common.Hash, jobID string) (*videojob.Update, error)

type vjSender struct {
	mu       sync.Mutex
	createFn func(n int, sessionID common.Hash, jobID string) (*videojob.Receipt, error)
	statusFn vjStatusFunc
	creates  []vjCreateCall
	statuses []vjStatusCall
}

func (s *vjSender) SendVideoJobCreate(_ context.Context, sessionID common.Hash, jobID string, job *videojob.CreateRequest) (*videojob.Receipt, error) {
	s.mu.Lock()
	s.creates = append(s.creates, vjCreateCall{sessionID: sessionID, jobID: jobID, req: *job})
	n := len(s.creates)
	fn := s.createFn
	s.mu.Unlock()
	if fn == nil {
		return vjReceipt(jobID), nil
	}
	return fn(n, sessionID, jobID)
}

func (s *vjSender) SendVideoJobStatus(ctx context.Context, sessionID common.Hash, jobID string) (*videojob.Update, error) {
	s.mu.Lock()
	s.statuses = append(s.statuses, vjStatusCall{sessionID: sessionID, jobID: jobID, at: time.Now()})
	n := len(s.statuses)
	fn := s.statusFn
	s.mu.Unlock()
	if fn == nil {
		return &videojob.Update{Status: videojob.StatusQueued}, nil
	}
	return fn(ctx, n, sessionID, jobID)
}

func (s *vjSender) createCalls() []vjCreateCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.creates)
}

func (s *vjSender) statusCalls() []vjStatusCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.statuses)
}

type vjResult struct {
	update videojob.Update
	err    error
}

func vjOK(status string) vjResult { return vjResult{update: videojob.Update{Status: status}} }

func vjDone(videoURL string) vjResult {
	return vjResult{update: videojob.Update{Status: videojob.StatusSucceeded, VideoURL: videoURL}}
}

func vjFailed(code, msg string) vjResult {
	return vjResult{update: videojob.Update{Status: videojob.StatusFailed, Error: &videojob.JobError{Code: code, Message: msg}}}
}

func vjErr(err error) vjResult { return vjResult{err: err} }

// vjSequence answers status calls in order and repeats the last answer.
func vjSequence(results ...vjResult) vjStatusFunc {
	return func(_ context.Context, n int, _ common.Hash, _ string) (*videojob.Update, error) {
		r := results[min(n, len(results))-1]
		if r.err != nil {
			return nil, r.err
		}
		u := r.update
		return &u, nil
	}
}

func vjReceipt(jobID string) *videojob.Receipt {
	return &videojob.Receipt{
		Object:       videojob.ObjectJob,
		Status:       videojob.StatusQueued,
		Vendor:       videojob.VendorRenderful,
		VendorTaskID: "gen-123",
		StatusURL:    "https://api.renderful.example/api/v1/generations/gen-123",
		QuoteSeconds: 120,
		ClientJobID:  jobID,
		Resolved: videojob.Resolved{
			Prompt:          "a fox in the snow",
			DurationSeconds: 5,
			AspectRatio:     "16:9",
			Resolution:      "720p",
		},
	}
}

// --- vendor test server -----------------------------------------------------

type vjVendorHit struct {
	path   string
	auth   string
	accept string
}

type vjVendor struct {
	mu      sync.Mutex
	hits    []vjVendorHit
	respond func(n int, r *http.Request) (int, string, string)
}

func (v *vjVendor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.mu.Lock()
	v.hits = append(v.hits, vjVendorHit{path: r.URL.Path, auth: r.Header.Get("Authorization"), accept: r.Header.Get("Accept")})
	n := len(v.hits)
	v.mu.Unlock()
	code, contentType, body := v.respond(n, r)
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.WriteHeader(code)
	_, _ = io.WriteString(w, body)
}

func (v *vjVendor) hitList() []vjVendorHit {
	v.mu.Lock()
	defer v.mu.Unlock()
	return slices.Clone(v.hits)
}

// newVJVendor starts a TLS server. Call it before newVJService so that the
// service's cleanup (cancel + wait) runs before the server closes.
func newVJVendor(t *testing.T, respond func(n int, r *http.Request) (int, string, string)) (*vjVendor, *httptest.Server) {
	t.Helper()
	v := &vjVendor{respond: respond}
	srv := httptest.NewTLSServer(v)
	t.Cleanup(srv.Close)
	return v, srv
}

// --- service helpers --------------------------------------------------------

// newVJService builds the service on an in-memory store (or store, when
// given). Its context is cancelled and every goroutine awaited on cleanup.
func newVJService(t *testing.T, sender *vjSender, sessions *vjSessions, store *storages.SessionStorage, maxWatch time.Duration) *VideoJobService {
	t.Helper()
	if store == nil {
		store = storages.NewSessionStorage(storages.NewTestStorage())
	}
	ctx, cancel := context.WithCancel(context.Background())
	svc := NewVideoJobService(ctx, sender, sessions, store, vjPollInterval, maxWatch, &lib.LoggerMock{})
	t.Cleanup(func() {
		cancel()
		vjWait(t, svc)
	})
	return svc
}

// vjWait waits for every watcher and async close of svc to return.
func vjWait(t *testing.T, svc *VideoJobService) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		svc.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(vjWaitFor):
		t.Error("video job goroutines did not return")
	}
}

func vjEventuallyStatus(t *testing.T, svc *VideoJobService, id, status string) videojob.Job {
	t.Helper()
	require.Eventually(t, func() bool {
		job, err := svc.Get(id)
		return err == nil && job.Status == status
	}, vjWaitFor, vjTick, "job %s never reached %s", id, status)
	job, err := svc.Get(id)
	require.NoError(t, err)
	return *job
}

func vjStored(t *testing.T, svc *VideoJobService, id string) (videojob.Job, string) {
	t.Helper()
	raw, err := svc.store.GetVideoJob(id)
	require.NoError(t, err)
	require.NotNil(t, raw, "job %s was not stored", id)
	var job videojob.Job
	require.NoError(t, json.Unmarshal(raw, &job))
	return job, string(raw)
}

func vjCreateParams(prompt string) CreateVideoJobParams {
	dur := 5
	return CreateVideoJobParams{
		Request:   videojob.CreateRequest{Prompt: prompt, DurationSeconds: &dur, AspectRatio: "16:9"},
		SessionID: vjSessionID,
	}
}

// --- Create -----------------------------------------------------------------

func TestVideoJobCreateSendsOnSessionAndPersists(t *testing.T) {
	sessions := newVJSessions()
	sender := &vjSender{
		createFn: func(_ int, _ common.Hash, jobID string) (*videojob.Receipt, error) {
			return vjReceipt(jobID), nil
		},
	}
	svc := newVJService(t, sender, sessions, nil, 0)

	job, err := svc.Create(context.Background(), vjCreateParams("a fox in the snow"))
	require.NoError(t, err)

	creates := sender.createCalls()
	require.Len(t, creates, 1)
	require.Equal(t, vjSessionID, creates[0].sessionID)
	require.True(t, strings.HasPrefix(creates[0].jobID, "vjob_"), "job id %q", creates[0].jobID)
	require.Len(t, creates[0].jobID, len("vjob_")+32)
	require.Equal(t, "a fox in the snow", creates[0].req.Prompt)
	require.NotNil(t, creates[0].req.DurationSeconds)
	require.Equal(t, 5, *creates[0].req.DurationSeconds)

	require.Equal(t, creates[0].jobID, job.ID)
	require.Equal(t, videojob.ObjectJob, job.Object)
	require.Equal(t, videojob.StatusQueued, job.Status)
	require.Equal(t, vjModelID.Hex(), job.Model)
	require.Equal(t, vjSessionID.Hex(), job.SessionID)
	require.Equal(t, int64(120), job.QuoteSeconds)
	require.Equal(t, videojob.VendorRenderful, job.Vendor)
	require.Equal(t, "gen-123", job.VendorTaskID)
	require.Equal(t, 5, job.Resolved.DurationSeconds)
	require.False(t, job.CreatedAt.IsZero())

	stored, raw := vjStored(t, svc, job.ID)
	require.Equal(t, job.ID, stored.ID)
	require.Equal(t, job.SessionID, stored.SessionID)
	require.NotContains(t, raw, "poll_auth")

	got, err := svc.Get(job.ID)
	require.NoError(t, err)
	require.Equal(t, job.ID, got.ID)
	require.Len(t, svc.List(), 1)

	_, err = svc.Get("vjob_unknown")
	require.ErrorIs(t, err, ErrVideoJobNotFound)
}

func TestVideoJobCreateWithSessionIDUsesSessionModel(t *testing.T) {
	sessions := newVJSessions()
	sessions.session = &structs.Session{Id: vjUserSessionID.Hex(), ModelAgentId: vjModelID.Hex()}
	sender := &vjSender{}
	svc := newVJService(t, sender, sessions, nil, 0)

	job, err := svc.Create(context.Background(), CreateVideoJobParams{
		Request:   videojob.CreateRequest{Prompt: "waves"},
		SessionID: vjUserSessionID,
	})
	require.NoError(t, err)

	require.Equal(t, vjModelID.Hex(), job.Model)
	require.Equal(t, vjUserSessionID.Hex(), job.SessionID)
	require.Contains(t, sessions.getCalls(), vjUserSessionID)

	creates := sender.createCalls()
	require.Len(t, creates, 1)
	require.Equal(t, vjUserSessionID, creates[0].sessionID)
}

func TestVideoJobCreateRejectsInvalidParams(t *testing.T) {
	sessions := newVJSessions()
	sender := &vjSender{}
	svc := newVJService(t, sender, sessions, nil, 0)

	_, err := svc.Create(context.Background(), CreateVideoJobParams{Request: videojob.CreateRequest{Prompt: "   "}, SessionID: vjSessionID})
	require.ErrorIs(t, err, videojob.ErrInvalidRequest)

	// No session: the caller opens one for the quote first.
	_, err = svc.Create(context.Background(), CreateVideoJobParams{Request: videojob.CreateRequest{Prompt: "x"}})
	require.ErrorIs(t, err, videojob.ErrInvalidRequest)
	require.Contains(t, err.Error(), "session_id is required")

	require.Empty(t, sender.createCalls())
	require.Empty(t, svc.List())
}

func TestVideoJobCreateProviderRefusedKeepsCallerSession(t *testing.T) {
	sessions := newVJSessions()
	sessions.session = &structs.Session{ModelAgentId: vjModelID.Hex()}
	sender := &vjSender{
		createFn: func(int, common.Hash, string) (*videojob.Receipt, error) {
			return nil, fmt.Errorf("%w: quote longer than the session", ErrProviderRefused)
		},
	}
	svc := newVJService(t, sender, sessions, nil, 0)

	_, err := svc.Create(context.Background(), CreateVideoJobParams{
		Request:   videojob.CreateRequest{Prompt: "x"},
		SessionID: vjUserSessionID,
	})
	require.ErrorIs(t, err, ErrProviderRefused)
	require.Len(t, sender.createCalls(), 1)

	vjWait(t, svc)
	require.Empty(t, svc.List(), "a refused submit leaves no job")
}

func TestVideoJobCreateRetriesTransportErrorWithSameID(t *testing.T) {
	sessions := newVJSessions()
	sender := &vjSender{
		createFn: func(n int, _ common.Hash, jobID string) (*videojob.Receipt, error) {
			if n == 1 {
				return nil, lib.WrapError(ErrProvider, errors.New("connection reset"))
			}
			return vjReceipt(jobID), nil
		},
	}
	svc := newVJService(t, sender, sessions, nil, 0)

	job, err := svc.Create(context.Background(), vjCreateParams("x"))
	require.NoError(t, err)

	creates := sender.createCalls()
	require.Len(t, creates, 2)
	require.Equal(t, creates[0].jobID, creates[1].jobID, "a resend must reuse the job id")
	require.Equal(t, creates[0].jobID, job.ID)
	require.Equal(t, creates[0].sessionID, creates[1].sessionID)
}

func TestVideoJobCreateGivesUpAfterSecondTransportError(t *testing.T) {
	sessions := newVJSessions()
	sender := &vjSender{
		createFn: func(int, common.Hash, string) (*videojob.Receipt, error) {
			return nil, lib.WrapError(ErrProvider, errors.New("connection refused"))
		},
	}
	svc := newVJService(t, sender, sessions, nil, 0)

	_, err := svc.Create(context.Background(), vjCreateParams("x"))
	require.ErrorIs(t, err, ErrProvider)
	require.Len(t, sender.createCalls(), videoCreateAttempts)

	vjWait(t, svc)
	require.Empty(t, svc.List(), "the caller's session stays open for a retry")
}

// --- watcher through the provider -------------------------------------------

func TestVideoJobWatchRelaySucceeded(t *testing.T) {
	sessions := newVJSessions()
	sender := &vjSender{statusFn: vjSequence(
		vjOK(videojob.StatusQueued),
		vjOK(videojob.StatusRunning),
		vjDone("https://cdn.renderful.example/out/gen-123.mp4"),
	)}
	svc := newVJService(t, sender, sessions, nil, 0)

	job, err := svc.Create(context.Background(), vjCreateParams("x"))
	require.NoError(t, err)

	done := vjEventuallyStatus(t, svc, job.ID, videojob.StatusSucceeded)
	require.Equal(t, "https://cdn.renderful.example/out/gen-123.mp4", done.VideoURL)
	require.Nil(t, done.Error)

	// The watcher returns on its own once the job ended.
	vjWait(t, svc)
	calls := sender.statusCalls()
	require.Len(t, calls, 3)
	for _, c := range calls {
		require.Equal(t, vjSessionID, c.sessionID)
		require.Equal(t, job.ID, c.jobID)
	}

	stored, _ := vjStored(t, svc, job.ID)
	require.Equal(t, videojob.StatusSucceeded, stored.Status)
	require.Equal(t, done.VideoURL, stored.VideoURL)
}

// A failed render ends the job and leaves the session alone: it is the
// caller's, and other jobs may run on it.
func TestVideoJobWatchRelayFailureEndsJob(t *testing.T) {
	for _, final := range []vjResult{
		vjFailed("content_policy", "prompt rejected"),
		{update: videojob.Update{Status: videojob.StatusExpired, Error: &videojob.JobError{Message: "expired"}}},
		{update: videojob.Update{Status: videojob.StatusCancelled, Error: &videojob.JobError{Message: "cancelled"}}},
	} {
		t.Run(final.update.Status, func(t *testing.T) {
			sessions := newVJSessions()
			sender := &vjSender{statusFn: vjSequence(vjOK(videojob.StatusRunning), final)}
			svc := newVJService(t, sender, sessions, nil, 0)

			job, err := svc.Create(context.Background(), vjCreateParams("x"))
			require.NoError(t, err)

			ended := vjEventuallyStatus(t, svc, job.ID, final.update.Status)
			vjWait(t, svc)

			require.NotNil(t, ended.Error)
			require.Equal(t, final.update.Error.Message, ended.Error.Message)
			require.Len(t, sender.statusCalls(), 2)

			stored, _ := vjStored(t, svc, job.ID)
			require.Equal(t, final.update.Status, stored.Status)
		})
	}
}

// --- backoff and give-up ----------------------------------------------------

func TestVideoJobPollErrorsBackOffAndRecover(t *testing.T) {
	sessions := newVJSessions()
	sender := &vjSender{statusFn: vjSequence(
		vjErr(errors.New("provider unreachable")),
		vjErr(errors.New("provider unreachable")),
		vjOK(videojob.StatusRunning),
		vjDone("https://cdn.example/v.mp4"),
	)}
	svc := newVJService(t, sender, sessions, nil, 0)

	job, err := svc.Create(context.Background(), vjCreateParams("x"))
	require.NoError(t, err)

	vjEventuallyStatus(t, svc, job.ID, videojob.StatusSucceeded)
	vjWait(t, svc)

	calls := sender.statusCalls()
	require.Len(t, calls, 4)
	// Timers never fire early, so these lower bounds are exact: the wait
	// doubles after each failed poll.
	require.GreaterOrEqual(t, calls[1].at.Sub(calls[0].at), 2*vjPollInterval)
	require.GreaterOrEqual(t, calls[2].at.Sub(calls[1].at), 4*vjPollInterval)
	require.GreaterOrEqual(t, calls[3].at.Sub(calls[2].at), vjPollInterval)
}

func TestVideoJobWatchTimeout(t *testing.T) {
	sessions := newVJSessions()
	sender := &vjSender{statusFn: vjSequence(vjOK(videojob.StatusRunning))}
	svc := newVJService(t, sender, sessions, nil, time.Millisecond)

	job, err := svc.Create(context.Background(), vjCreateParams("x"))
	require.NoError(t, err)

	ended := vjEventuallyStatus(t, svc, job.ID, videojob.StatusFailed)
	require.NotNil(t, ended.Error)
	require.Equal(t, "watch_timeout", ended.Error.Code)
	vjWait(t, svc)

	require.Empty(t, sender.statusCalls(), "the watch is over before the first poll")
}

// --- Resume -----------------------------------------------------------------

func vjPutJob(t *testing.T, store *storages.SessionStorage, job videojob.Job) {
	t.Helper()
	raw, err := json.Marshal(job)
	require.NoError(t, err)
	require.NoError(t, store.PutVideoJob(job.ID, raw))
}

func TestVideoJobResumeWatchesUnfinishedJobs(t *testing.T) {
	store := storages.NewSessionStorage(storages.NewTestStorage())
	now := time.Now().UTC()
	pending := videojob.Job{
		Object: videojob.ObjectJob, ID: "vjob_pending", Model: vjModelID.Hex(), Status: videojob.StatusRunning,
		SessionID: vjSessionID.Hex(), Vendor: videojob.VendorBytePlus, VendorTaskID: "cgt-1",
		StatusURL: "https://ark.example.com/tasks/cgt-1", CreatedAt: now, UpdatedAt: now,
	}
	finished := videojob.Job{
		Object: videojob.ObjectJob, ID: "vjob_finished", Model: vjModelID.Hex(), Status: videojob.StatusSucceeded,
		SessionID: vjUserSessionID.Hex(), Vendor: videojob.VendorBytePlus, VendorTaskID: "cgt-0",
		VideoURL: "https://ark.example.com/old.mp4", CreatedAt: now.Add(-time.Hour), UpdatedAt: now,
	}
	vjPutJob(t, store, pending)
	vjPutJob(t, store, finished)
	require.NoError(t, store.PutVideoJob("vjob_corrupt", []byte("{not json")))

	sessions := newVJSessions()
	sender := &vjSender{statusFn: vjSequence(vjOK(videojob.StatusRunning), vjDone("https://ark.example.com/new.mp4"))}
	svc := newVJService(t, sender, sessions, store, 0)

	svc.Resume()
	// A second Resume does not add a second watcher for the same job.
	svc.Resume()

	list := svc.List()
	require.Len(t, list, 2, "the corrupt record is skipped")
	require.Equal(t, pending.ID, list[0].ID, "newest first")
	require.Equal(t, finished.ID, list[1].ID)

	done := vjEventuallyStatus(t, svc, pending.ID, videojob.StatusSucceeded)
	require.Equal(t, "https://ark.example.com/new.mp4", done.VideoURL)
	vjWait(t, svc)

	calls := sender.statusCalls()
	require.Len(t, calls, 2)
	for _, c := range calls {
		require.Equal(t, pending.ID, c.jobID)
		require.Equal(t, vjSessionID, c.sessionID)
	}
	old, err := svc.Get(finished.ID)
	require.NoError(t, err)
	require.Equal(t, videojob.StatusSucceeded, old.Status)
	require.Equal(t, "https://ark.example.com/old.mp4", old.VideoURL)
}

// --- Quote ------------------------------------------------------------------

func vjSpec(secondsPerUnit float64, maxSession int64) videojob.Spec {
	return videojob.Spec{
		ModelName:    "Seedance",
		MinDuration:  2,
		MaxDuration:  10,
		AspectRatios: []string{"16:9", "9:16"},
		Resolutions:  []string{"720p", "1080p"},
		Defaults:     videojob.Defaults{DurationSeconds: 5, AspectRatio: "16:9", Resolution: "720p"},
		Quote: videojob.Quote{
			SecondsPerWorkUnit: secondsPerUnit,
			ResolutionWeights:  map[string]float64{"720p": 1, "1080p": 2},
			MinSessionSeconds:  60,
			MaxSessionSeconds:  maxSession,
		},
	}
}

func TestVideoQuotePerProvider(t *testing.T) {
	specA, specB := vjSpec(10, 3600), vjSpec(30, 500)
	require.NoError(t, specA.Validate())
	require.NoError(t, specB.Validate())

	provA := common.HexToAddress("0x00000000000000000000000000000000000000aa")
	provB := common.HexToAddress("0x00000000000000000000000000000000000000bb")
	provC := common.HexToAddress("0x00000000000000000000000000000000000000cc")
	bidA := common.HexToHash("0xa1")
	sessions := newVJSessions()
	sessions.guides = []structs.VideoProviderGuide{
		{Provider: provA, BidID: bidA, PricePerSecondWei: "1000", Guide: specA.Guide()},
		{Provider: provB, BidID: common.HexToHash("0xb1"), PricePerSecondWei: "7", Guide: specB.Guide()},
		{Provider: provC, BidID: common.HexToHash("0xc1"), PricePerSecondWei: "5", Error: "provider did not answer"},
	}
	svc := newVJService(t, &vjSender{}, sessions, nil, 0)

	dur := 6
	quotes, err := svc.Quote(context.Background(), vjModelID, &videojob.CreateRequest{Prompt: "x", DurationSeconds: &dur, Resolution: "1080p"})
	require.NoError(t, err)
	require.Len(t, quotes, 3)

	// A: 6s × 2 (1080p) × 10 = 120s at 1000 wei/s.
	require.Equal(t, provA, quotes[0].Provider)
	require.Equal(t, bidA, quotes[0].BidID)
	require.Empty(t, quotes[0].Error)
	require.Equal(t, int64(120), quotes[0].SessionSeconds)
	require.Equal(t, "120000", quotes[0].SessionCostWei)
	require.NotNil(t, quotes[0].Resolved)
	require.Equal(t, videojob.Resolved{Prompt: "x", DurationSeconds: 6, AspectRatio: "16:9", Resolution: "1080p"}, *quotes[0].Resolved)

	// B: 6 × 2 × 30 = 360s at 7 wei/s.
	require.Empty(t, quotes[1].Error)
	require.Equal(t, int64(360), quotes[1].SessionSeconds)
	require.Equal(t, "2520", quotes[1].SessionCostWei)

	// C: no guide, its error passes through.
	require.Equal(t, provC, quotes[2].Provider)
	require.Equal(t, "provider did not answer", quotes[2].Error)
	require.Zero(t, quotes[2].SessionSeconds)
	require.Empty(t, quotes[2].SessionCostWei)
	require.Nil(t, quotes[2].Resolved)

	// The contract minimum raises a short quote and its maximum refuses a long one.
	sessions.mu.Lock()
	sessions.boundsMin, sessions.boundsMax = 300, 200
	sessions.mu.Unlock()
	quotes, err = svc.Quote(context.Background(), vjModelID, &videojob.CreateRequest{Prompt: "x", DurationSeconds: &dur, Resolution: "1080p"})
	require.NoError(t, err)
	require.Contains(t, quotes[0].Error, "contract max", "120 s is under the 200 s max but the 300 s min is not")
	sessions.mu.Lock()
	sessions.boundsMin, sessions.boundsMax = 300, 86400
	sessions.mu.Unlock()
	quotes, err = svc.Quote(context.Background(), vjModelID, &videojob.CreateRequest{Prompt: "x", DurationSeconds: &dur, Resolution: "1080p"})
	require.NoError(t, err)
	require.Equal(t, int64(300), quotes[0].SessionSeconds, "raised to the contract minimum")
	require.Equal(t, "300000", quotes[0].SessionCostWei)
	sessions.mu.Lock()
	sessions.boundsMin, sessions.boundsMax = 0, 0
	sessions.mu.Unlock()

	// Prompt only: defaults (5s 720p) → 50s, raised to the 60s minimum.
	quotes, err = svc.Quote(context.Background(), vjModelID, &videojob.CreateRequest{Prompt: "x"})
	require.NoError(t, err)
	require.Equal(t, int64(60), quotes[0].SessionSeconds)
	require.Equal(t, "60000", quotes[0].SessionCostWei)
	require.Equal(t, int64(150), quotes[1].SessionSeconds)

	// 10s 1080p: A 200s; B 600s is over its 500s maximum.
	dur = 10
	quotes, err = svc.Quote(context.Background(), vjModelID, &videojob.CreateRequest{Prompt: "x", DurationSeconds: &dur, Resolution: "1080p"})
	require.NoError(t, err)
	require.Equal(t, int64(200), quotes[0].SessionSeconds)
	require.Contains(t, quotes[1].Error, "exceeds")
	require.Empty(t, quotes[1].SessionCostWei)
	require.Nil(t, quotes[1].Resolved)

	// An invalid request is an error per provider, not for the whole call.
	dur = 20
	quotes, err = svc.Quote(context.Background(), vjModelID, &videojob.CreateRequest{Prompt: "x", DurationSeconds: &dur})
	require.NoError(t, err)
	require.Len(t, quotes, 3)
	for _, q := range quotes[:2] {
		require.Contains(t, q.Error, "duration_seconds 20")
		require.Zero(t, q.SessionSeconds)
		require.Empty(t, q.SessionCostWei)
		require.Nil(t, q.Resolved)
	}
	require.Equal(t, "provider did not answer", quotes[2].Error)

	quotes, err = svc.Quote(context.Background(), vjModelID, &videojob.CreateRequest{Prompt: ""})
	require.NoError(t, err)
	require.Contains(t, quotes[0].Error, "prompt is required")

	sessions.mu.Lock()
	require.NotEmpty(t, sessions.guideReqs)
	for _, m := range sessions.guideReqs {
		require.Equal(t, vjModelID, m)
	}
	sessions.mu.Unlock()
}

func TestVideoQuotePassesGuidesError(t *testing.T) {
	sessions := newVJSessions()
	sessions.guidesErr = errors.New("chain unavailable")
	svc := newVJService(t, &vjSender{}, sessions, nil, 0)

	quotes, err := svc.Quote(context.Background(), vjModelID, &videojob.CreateRequest{Prompt: "x"})
	require.ErrorIs(t, err, sessions.guidesErr)
	require.Nil(t, quotes)
}

// A relayed link that is not a public https URL never reaches the caller:
// the job fails.
func TestVideoJobNonPublicVideoURLFails(t *testing.T) {
	for _, bad := range []string{"http://cdn.example.com/v.mp4", "https://127.0.0.1/v.mp4", "javascript:alert(1)"} {
		t.Run(bad, func(t *testing.T) {
			sessions := newVJSessions()
			sender := &vjSender{statusFn: vjSequence(vjDone(bad))}
			svc := newVJService(t, sender, sessions, nil, 0)

			job, err := svc.Create(context.Background(), vjCreateParams("x"))
			require.NoError(t, err)
			ended := vjEventuallyStatus(t, svc, job.ID, videojob.StatusFailed)
			vjWait(t, svc)

			require.Empty(t, ended.VideoURL)
			require.NotNil(t, ended.Error)
			require.Equal(t, "invalid_video_url", ended.Error.Code)

			stored, _ := vjStored(t, svc, job.ID)
			require.Empty(t, stored.VideoURL)
		})
	}
}
