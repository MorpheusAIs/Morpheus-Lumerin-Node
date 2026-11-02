package proxyapi

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	morrpc "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi/morrpcmessage"
	sessionrepo "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/session"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/videojob"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// fakes
// ---------------------------------------------------------------------------

type videoTestValidator struct{}

func (videoTestValidator) Struct(interface{}) error { return nil }

type videoTestChain struct{}

func (videoTestChain) ModelExists(context.Context, common.Hash) (bool, error) { return true, nil }

type videoTestConn struct{}

func (videoTestConn) TryConnect(context.Context, string) error { return nil }

// fakeVideoVendor is a BytePlus- or Renderful-shaped vendor: POST creates a
// task, GET returns its status body. It records what the provider sent.
type fakeVideoVendor struct {
	srv *httptest.Server

	mu             sync.Mutex
	taskID         string
	createCode     int
	createBody     string
	createDelay    time.Duration
	statusBody     string
	statusDelay    time.Duration
	creates        int
	statusReads    int
	lastCreate     map[string]any
	lastCreatePath string
	lastCreateAuth string
	lastStatusPath string
	lastStatusAuth string
}

func newFakeVideoVendor(t *testing.T, taskID string) *fakeVideoVendor {
	t.Helper()
	v := &fakeVideoVendor{taskID: taskID}
	v.srv = httptest.NewServer(http.HandlerFunc(v.serve))
	t.Cleanup(v.srv.Close)
	return v
}

func (v *fakeVideoVendor) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	w.Header().Set("Content-Type", "application/json")

	v.mu.Lock()
	switch r.Method {
	case http.MethodPost:
		v.creates++
		v.lastCreatePath = r.URL.Path
		v.lastCreateAuth = r.Header.Get("Authorization")
		var parsed map[string]any
		_ = json.Unmarshal(body, &parsed)
		v.lastCreate = parsed
		code, resp, delay := v.createCode, v.createBody, v.createDelay
		if code == 0 {
			code = http.StatusOK
			resp = `{"id":"` + v.taskID + `"}`
		}
		v.mu.Unlock()
		if delay > 0 {
			time.Sleep(delay)
		}
		w.WriteHeader(code)
		_, _ = w.Write([]byte(resp))
	case http.MethodGet:
		v.statusReads++
		v.lastStatusPath = r.URL.Path
		v.lastStatusAuth = r.Header.Get("Authorization")
		resp, delay := v.statusBody, v.statusDelay
		v.mu.Unlock()
		if delay > 0 {
			time.Sleep(delay)
		}
		_, _ = w.Write([]byte(resp))
	default:
		v.mu.Unlock()
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (v *fakeVideoVendor) set(fn func(v *fakeVideoVendor)) {
	v.mu.Lock()
	defer v.mu.Unlock()
	fn(v)
}

func (v *fakeVideoVendor) counts() (creates, statusReads int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.creates, v.statusReads
}

// ---------------------------------------------------------------------------
// harness
// ---------------------------------------------------------------------------

const (
	videoTestModelID    = "0x5eed5eed5eed5eed5eed5eed5eed5eed5eed5eed5eed5eed5eed5eed5eed5eed"
	videoTestSessionID  = "0xabcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	videoTestVendorKey  = "vendor-secret-key"
	videoTestBytePlusID = "cgt-test-1"
	videoTestRenderID   = "gen-test-1"
)

type videoHarnessOpts struct {
	apiType   string // default byteplus-video
	modelName string // default test-seedance
	endsAt    time.Time
	// providerAddr overrides the session's provider (default: this node)
	providerAddr *common.Address
	health       ModelHealthTracker
	// chain, when set, is the receiver's blockchain service: it answers the
	// on-chain session read the length check prefers.
	chain *videoTestChainSessions
}

// videoTestChainSessions is the receiver's BidGetter that can also read a
// session from the chain (videoSessionReader).
type videoTestChainSessions struct {
	mu      sync.Mutex
	session *structs.Session
	err     error
	reads   int
}

func (c *videoTestChainSessions) GetBidByID(context.Context, common.Hash) (*structs.Bid, error) {
	return nil, errors.New("not used by video jobs")
}

func (c *videoTestChainSessions) GetSession(_ context.Context, _ common.Hash) (*structs.Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reads++
	if c.err != nil {
		return nil, c.err
	}
	s := *c.session
	return &s, nil
}

func chainSession(openedAt, endsAt time.Time, closedAt int64) *structs.Session {
	return &structs.Session{
		OpenedAt: big.NewInt(openedAt.Unix()),
		EndsAt:   big.NewInt(endsAt.Unix()),
		ClosedAt: big.NewInt(closedAt),
	}
}

type videoHarness struct {
	c           *MORRPCController
	receiver    *ProxyReceiver
	ss          *storages.SessionStorage
	loader      *config.ModelConfigLoader
	vendor      *fakeVideoVendor
	userKey     *ecdsa.PrivateKey
	userPrKey   lib.HexString
	userAddr    common.Address
	providerPub lib.HexString
	sessionID   common.Hash
	modelID     common.Hash
	rpc         *morrpc.MORRPCMessage
}

func newVideoHarness(t *testing.T, o videoHarnessOpts) *videoHarness {
	t.Helper()
	if o.apiType == "" {
		o.apiType = aiengine.API_TYPE_BYTEPLUS_VIDEO
	}
	if o.modelName == "" {
		o.modelName = "test-seedance"
	}
	if o.endsAt.IsZero() {
		o.endsAt = time.Now().Add(2 * time.Hour)
	}
	taskID := videoTestBytePlusID
	if o.apiType == aiengine.API_TYPE_RENDERFUL {
		taskID = videoTestRenderID
	}
	vendor := newFakeVideoVendor(t, taskID)

	providerKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	providerPrKey := lib.HexString(crypto.FromECDSA(providerKey))
	providerPub := lib.HexString(crypto.FromECDSAPub(&providerKey.PublicKey))
	providerAddr := crypto.PubkeyToAddress(providerKey.PublicKey)
	if o.providerAddr != nil {
		providerAddr = *o.providerAddr
	}

	userKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	userPub := lib.HexString(crypto.FromECDSAPub(&userKey.PublicKey))
	userAddr := crypto.PubkeyToAddress(userKey.PublicKey)

	modelID := common.HexToHash(videoTestModelID)
	sessionID := common.HexToHash(videoTestSessionID)

	ss := storages.NewSessionStorage(storages.NewTestStorage())
	require.NoError(t, ss.AddUserBound(&storages.User{Addr: userAddr.Hex(), PubKey: userPub.Hex()}))
	require.NoError(t, ss.AddSession(&storages.Session{
		Id:           sessionID.Hex(),
		UserAddr:     userAddr.Hex(),
		ProviderAddr: providerAddr.Hex(),
		EndsAt:       big.NewInt(o.endsAt.Unix()),
		ModelID:      modelID.Hex(),
	}))

	entry := `{"modelId":"` + modelID.Hex() + `","modelName":"` + o.modelName + `","apiType":"` + o.apiType +
		`","apiUrl":"` + vendor.srv.URL + `","apiKey":"` + videoTestVendorKey + `"`
	entry += `}`
	content := `{"models":[` + entry + `]}`

	log := &lib.LoggerMock{}
	loader := config.NewModelConfigLoader(filepath.Join(t.TempDir(), "models-config.json"), content,
		videoTestValidator{}, videoTestChain{}, videoTestConn{}, log)
	require.NoError(t, loader.Init())

	repo := sessionrepo.NewSessionRepositoryCached(ss, nil, nil, log)
	// a nil *videoTestChainSessions must stay a nil interface
	var chain BidGetter
	if o.chain != nil {
		chain = o.chain
	}
	receiver := NewProxyReceiver(providerPrKey, providerPub, ss, nil, big.NewInt(1), loader, chain, repo)
	if o.health != nil {
		receiver.SetModelHealthTracker(o.health)
	}
	valid, err := config.NewValidator()
	require.NoError(t, err)
	c := NewMORRPCController(receiver, valid, repo, ss, providerPrKey, nil)

	return &videoHarness{
		c:           c,
		receiver:    receiver,
		ss:          ss,
		loader:      loader,
		vendor:      vendor,
		userKey:     userKey,
		userPrKey:   lib.HexString(crypto.FromECDSA(userKey)),
		userAddr:    userAddr,
		providerPub: providerPub,
		sessionID:   sessionID,
		modelID:     modelID,
		rpc:         morrpc.NewMorRpc(),
	}
}

// handle runs one MORRPC message and returns the single response it sent.
func (h *videoHarness) handle(t *testing.T, m *morrpc.RPCMessage) (*morrpc.RpcResponse, error) {
	t.Helper()
	var got *morrpc.RpcResponse
	sends := 0
	err := h.c.Handle(context.Background(), *m, &lib.LoggerMock{}, func(r *morrpc.RpcResponse) error {
		sends++
		got = r
		return nil
	})
	if err != nil {
		require.Zero(t, sends, "an error must not follow a response")
		return nil, err
	}
	require.Equal(t, 1, sends, "exactly one response")
	return got, nil
}

// open checks the provider signature on a video response and decrypts its
// payload with the user's key.
func (h *videoHarness) open(t *testing.T, resp *morrpc.RpcResponse, reqID string) []byte {
	t.Helper()
	require.NotNil(t, resp)
	require.Equal(t, reqID, resp.ID)
	require.Nil(t, resp.Error)
	require.NotNil(t, resp.Result)

	var res morrpc.SessionVideoRes
	require.NoError(t, json.Unmarshal(*resp.Result, &res))
	sig := res.Signature
	res.Signature = lib.HexString{}
	signed, err := json.Marshal(res)
	require.NoError(t, err)
	require.True(t, lib.VerifySignature(signed, sig, h.providerPub), "response must carry the provider signature")

	plain, err := lib.DecryptString(res.Message, h.userPrKey.Hex())
	require.NoError(t, err, "response must be encrypted to the session user")
	return []byte(plain)
}

func (h *videoHarness) createWith(t *testing.T, key *ecdsa.PrivateKey, jobID string, req videojob.CreateRequest) (*videojob.Receipt, []byte, error) {
	t.Helper()
	m, err := h.rpc.SessionVideoCreateRequest(h.sessionID, jobID, &req, lib.HexString(crypto.FromECDSA(key)), "req-create")
	require.NoError(t, err)
	resp, err := h.handle(t, m)
	if err != nil {
		return nil, nil, err
	}
	plain := h.open(t, resp, "req-create")
	var receipt videojob.Receipt
	require.NoError(t, json.Unmarshal(plain, &receipt))
	return &receipt, plain, nil
}

func (h *videoHarness) create(t *testing.T, jobID string, req videojob.CreateRequest) (*videojob.Receipt, []byte, error) {
	t.Helper()
	return h.createWith(t, h.userKey, jobID, req)
}

func (h *videoHarness) statusWith(t *testing.T, key *ecdsa.PrivateKey, jobID string) (*videojob.Update, error) {
	t.Helper()
	m, err := h.rpc.SessionVideoStatusRequest(h.sessionID, jobID, lib.HexString(crypto.FromECDSA(key)), "req-status")
	require.NoError(t, err)
	resp, err := h.handle(t, m)
	if err != nil {
		return nil, err
	}
	plain := h.open(t, resp, "req-status")
	var update videojob.Update
	require.NoError(t, json.Unmarshal(plain, &update))
	return &update, nil
}

func (h *videoHarness) status(t *testing.T, jobID string) (*videojob.Update, error) {
	t.Helper()
	return h.statusWith(t, h.userKey, jobID)
}

func (h *videoHarness) storedTask(t *testing.T, jobID string) []byte {
	t.Helper()
	raw, err := h.ss.GetVideoTask(h.sessionID.Hex(), jobID)
	require.NoError(t, err)
	return raw
}

// setSessionEnd rewrites the cached session with a new EndsAt.
func (h *videoHarness) setSessionEnd(t *testing.T, endsAt time.Time) {
	t.Helper()
	ses, err := h.ss.GetSession(h.sessionID.Hex())
	require.NoError(t, err)
	require.NotNil(t, ses)
	ses.EndsAt = big.NewInt(endsAt.Unix())
	require.NoError(t, h.ss.AddSession(ses))
}

// signVideoParams signs hand-built params the way the morrpc builders do:
// keccak256 over the JSON without the signature.
func signVideoParams(t *testing.T, key *ecdsa.PrivateKey, params any) lib.HexString {
	t.Helper()
	raw, err := json.Marshal(params)
	require.NoError(t, err)
	sig, err := crypto.Sign(crypto.Keccak256Hash(raw).Bytes(), key)
	require.NoError(t, err)
	return sig
}

func craftedVideoCreate(t *testing.T, key *ecdsa.PrivateKey, sessionID common.Hash, jobID string, job videojob.CreateRequest, ts time.Time) *morrpc.RPCMessage {
	t.Helper()
	message, err := json.Marshal(&job)
	require.NoError(t, err)
	params := morrpc.SessionVideoCreateReq{
		SessionID: sessionID,
		JobID:     jobID,
		Message:   string(message),
		Timestamp: uint64(ts.UnixMilli()),
	}
	params.Signature = signVideoParams(t, key, params)
	raw, err := json.Marshal(params)
	require.NoError(t, err)
	return &morrpc.RPCMessage{ID: "req-create", Method: "session.video.create", Params: raw}
}

func craftedVideoStatus(t *testing.T, key *ecdsa.PrivateKey, sessionID common.Hash, jobID string, ts time.Time) *morrpc.RPCMessage {
	t.Helper()
	params := morrpc.SessionVideoStatusReq{
		SessionID: sessionID,
		JobID:     jobID,
		Timestamp: uint64(ts.UnixMilli()),
	}
	params.Signature = signVideoParams(t, key, params)
	raw, err := json.Marshal(params)
	require.NoError(t, err)
	return &morrpc.RPCMessage{ID: "req-status", Method: "session.video.status", Params: raw}
}

func vpIntPtr(v int) *int { return &v }

func expectedVideoQuote(t *testing.T, h *videoHarness, req videojob.CreateRequest) (videojob.Resolved, int64) {
	t.Helper()
	guide, err := aiengine.VideoGuideFor(h.loader.ModelConfigFromID(h.modelID.Hex()), &lib.LoggerMock{})
	require.NoError(t, err)
	resolved, quote, err := guide.SessionSeconds(&req)
	require.NoError(t, err)
	return resolved, quote
}

const bytePlusSucceededBody = `{"id":"cgt-test-1","model":"test-seedance","status":"succeeded",` +
	`"content":{"video_url":"https://cdn.example.com/clip.mp4"},"created_at":1699999000,"updated_at":1700000000}`

// ---------------------------------------------------------------------------
// session.video.create
// ---------------------------------------------------------------------------

func TestVideoProviderCreateReceiptProviderPoll(t *testing.T) {
	health := &fakeHealthTracker{}
	h := newVideoHarness(t, videoHarnessOpts{health: health})
	req := videojob.CreateRequest{Prompt: "  a cat surfing a wave  "}

	receipt, plain, err := h.create(t, "job-1", req)
	require.NoError(t, err)

	wantResolved, wantQuote := expectedVideoQuote(t, h, req)
	// unknown BytePlus model: 5 s at 720p (weight 2.25) × 75 s per work unit
	require.EqualValues(t, 844, wantQuote)

	require.Equal(t, videojob.ObjectJob, receipt.Object)
	require.Equal(t, videojob.StatusQueued, receipt.Status)
	require.Equal(t, videojob.VendorBytePlus, receipt.Vendor)
	require.Equal(t, videoTestBytePlusID, receipt.VendorTaskID)
	require.Equal(t, h.vendor.srv.URL+"/contents/generations/tasks/"+videoTestBytePlusID, receipt.StatusURL)
	require.Equal(t, wantQuote, receipt.QuoteSeconds)
	require.Equal(t, "job-1", receipt.ClientJobID)
	require.Equal(t, wantResolved, receipt.Resolved)
	require.Equal(t, videojob.Resolved{
		Prompt:          "a cat surfing a wave",
		DurationSeconds: 5,
		AspectRatio:     "16:9",
		Resolution:      "720p",
	}, receipt.Resolved)

	// the provider's own vendor key never reaches the consumer
	require.NotContains(t, string(plain), videoTestVendorKey)
	require.NotContains(t, string(plain), "poll_auth")

	// what the vendor received
	h.vendor.mu.Lock()
	require.Equal(t, 1, h.vendor.creates)
	require.Equal(t, "/contents/generations/tasks", h.vendor.lastCreatePath)
	require.Equal(t, "Bearer "+videoTestVendorKey, h.vendor.lastCreateAuth)
	require.Equal(t, "test-seedance", h.vendor.lastCreate["model"])
	require.Equal(t, "16:9", h.vendor.lastCreate["ratio"])
	require.EqualValues(t, 5, h.vendor.lastCreate["duration"])
	require.Equal(t, "720p", h.vendor.lastCreate["resolution"])
	content, _ := h.vendor.lastCreate["content"].([]any)
	require.Len(t, content, 1)
	require.Equal(t, "a cat surfing a wave", content[0].(map[string]any)["text"])
	h.vendor.mu.Unlock()

	// the record is stored for status and idempotency
	raw := h.storedTask(t, "job-1")
	require.NotNil(t, raw)
	var task providerVideoTask
	require.NoError(t, json.Unmarshal(raw, &task))
	require.Equal(t, "job-1", task.JobID)
	require.Equal(t, videoTestBytePlusID, task.VendorTaskID)
	require.Equal(t, h.modelID.Hex(), task.ModelID)
	require.Equal(t, wantQuote, task.QuoteSeconds)

	// an accepted vendor task counts as a healthy backend call
	require.Equal(t, 1, health.successes)
	require.Equal(t, 0, health.failures)
}

func TestVideoProviderCreateRenderfulReceiptAndStatus(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{apiType: aiengine.API_TYPE_RENDERFUL, modelName: "seedance-2.5-t2v"})
	req := videojob.CreateRequest{Prompt: "a fox in snow", DurationSeconds: vpIntPtr(8), AspectRatio: "9:16", Resolution: "480P"}

	receipt, _, err := h.create(t, "job-r", req)
	require.NoError(t, err)
	wantResolved, wantQuote := expectedVideoQuote(t, h, req)

	require.Equal(t, videojob.VendorRenderful, receipt.Vendor)
	require.Equal(t, videoTestRenderID, receipt.VendorTaskID)
	require.Equal(t, h.vendor.srv.URL+"/generations/"+videoTestRenderID, receipt.StatusURL)
	require.Equal(t, wantQuote, receipt.QuoteSeconds)
	require.EqualValues(t, 600, wantQuote) // 8 s × 480p weight 1 × 75
	require.Equal(t, wantResolved, receipt.Resolved)
	require.Equal(t, "480p", receipt.Resolved.Resolution)

	h.vendor.mu.Lock()
	require.Equal(t, "/generations", h.vendor.lastCreatePath)
	require.Equal(t, "text-to-video", h.vendor.lastCreate["type"])
	require.Equal(t, "seedance-2.5-t2v", h.vendor.lastCreate["model"])
	require.Equal(t, "9:16", h.vendor.lastCreate["aspect_ratio"])
	require.EqualValues(t, 8, h.vendor.lastCreate["duration"])
	require.Equal(t, false, h.vendor.lastCreate["generate_audio"])
	h.vendor.mu.Unlock()

	h.vendor.set(func(v *fakeVideoVendor) {
		v.statusBody = `{"id":"gen-test-1","status":"completed","outputs":["https://cdn.example.com/r.mp4"]}`
	})
	update, err := h.status(t, "job-r")
	require.NoError(t, err)
	require.Equal(t, videojob.StatusSucceeded, update.Status)
	require.Equal(t, "https://cdn.example.com/r.mp4", update.VideoURL)
	h.vendor.mu.Lock()
	require.Equal(t, "/generations/"+videoTestRenderID, h.vendor.lastStatusPath)
	require.Equal(t, "Bearer "+videoTestVendorKey, h.vendor.lastStatusAuth)
	h.vendor.mu.Unlock()
}

func TestVideoProviderCreateIdempotentByJobID(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})
	req := videojob.CreateRequest{Prompt: "a red balloon", DurationSeconds: vpIntPtr(6)}

	first, _, err := h.create(t, "job-1", req)
	require.NoError(t, err)
	second, _, err := h.create(t, "job-1", req)
	require.NoError(t, err)

	require.Equal(t, *first, *second)
	creates, _ := h.vendor.counts()
	require.Equal(t, 1, creates, "a retried accept must not start a second vendor task")

	// another job on the same session is a new task
	third, _, err := h.create(t, "job-2", req)
	require.NoError(t, err)
	require.Equal(t, "job-2", third.ClientJobID)
	creates, _ = h.vendor.counts()
	require.Equal(t, 2, creates)
	require.NotNil(t, h.storedTask(t, "job-2"))
}

// A session carries as many jobs as its length pays for: each accepted job
// takes its quote, and a job that no longer fits is refused before the
// vendor. A retried accept is not counted twice.
func TestVideoProviderCreateQuotesAddUpOnSession(t *testing.T) {
	now := time.Now()
	// 1700 s on chain; the default job quotes 844 s.
	chain := &videoTestChainSessions{session: chainSession(now.Add(-10*time.Second), now.Add(1690*time.Second), 0)}
	h := newVideoHarness(t, videoHarnessOpts{chain: chain})
	req := videojob.CreateRequest{Prompt: "a paper boat"}

	first, _, err := h.create(t, "job-1", req)
	require.NoError(t, err)
	require.EqualValues(t, 844, first.QuoteSeconds)
	_, _, err = h.create(t, "job-1", req)
	require.NoError(t, err, "a retry of an accepted job is not a second quote")
	_, _, err = h.create(t, "job-2", req)
	require.NoError(t, err, "844 + 844 fits in 1700")

	_, _, err = h.create(t, "job-3", req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "on top of the 1688 seconds earlier jobs on it took")
	require.Nil(t, h.storedTask(t, "job-3"))
	creates, _ := h.vendor.counts()
	require.Equal(t, 2, creates)

	// a shorter job still fits in what is left
	_, _, err = h.create(t, "job-4", videojob.CreateRequest{Prompt: "a paper boat", DurationSeconds: vpIntPtr(4), Resolution: "480p"})
	require.Error(t, err, "300 s is the minimum session a job quotes, and only 12 s are left")
}

// Two copies of the same accept arriving together (a consumer retry racing
// the original) must still start one vendor task.
func TestVideoProviderCreateConcurrentRetryStartsOneTask(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})
	h.vendor.set(func(v *fakeVideoVendor) { v.createDelay = 150 * time.Millisecond })

	req := videojob.CreateRequest{Prompt: "rain on a window"}
	msgs := make([]*morrpc.RPCMessage, 2)
	for i := range msgs {
		m, err := h.rpc.SessionVideoCreateRequest(h.sessionID, "job-1", &req, h.userPrKey, "req-create")
		require.NoError(t, err)
		msgs[i] = m
	}

	var wg sync.WaitGroup
	resps := make([]*morrpc.RpcResponse, 2)
	errs := make([]error, 2)
	for i := range msgs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = h.c.Handle(context.Background(), *msgs[i], &lib.LoggerMock{}, func(r *morrpc.RpcResponse) error {
				resps[i] = r
				return nil
			})
		}(i)
	}
	wg.Wait()

	receipts := make([]videojob.Receipt, 2)
	for i := range msgs {
		require.NoError(t, errs[i])
		require.NoError(t, json.Unmarshal(h.open(t, resps[i], "req-create"), &receipts[i]))
	}
	require.Equal(t, receipts[0], receipts[1])
	creates, _ := h.vendor.counts()
	require.Equal(t, 1, creates)
}

func TestVideoProviderCreateQuoteLongerThanSessionRefused(t *testing.T) {
	// default job quotes 844 s; 700 s left + 120 s grace is not enough
	h := newVideoHarness(t, videoHarnessOpts{endsAt: time.Now().Add(700 * time.Second)})

	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a slow sunrise"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "needs a 844-second session")
	creates, _ := h.vendor.counts()
	require.Zero(t, creates, "a job the session cannot pay for must not reach the vendor")
	require.Nil(t, h.storedTask(t, "job-1"))
}

func TestVideoProviderCreateQuoteWithinGraceAccepted(t *testing.T) {
	// 730 s left is less than the 844 s quote, but within the 120 s grace
	h := newVideoHarness(t, videoHarnessOpts{endsAt: time.Now().Add(730 * time.Second)})

	receipt, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a slow sunrise"})
	require.NoError(t, err)
	require.EqualValues(t, 844, receipt.QuoteSeconds)
}

func TestVideoProviderCreateInvalidRequestRefused(t *testing.T) {
	cases := []struct {
		name string
		req  videojob.CreateRequest
	}{
		{"empty prompt", videojob.CreateRequest{Prompt: ""}},
		{"blank prompt", videojob.CreateRequest{Prompt: "   "}},
		{"prompt too long", videojob.CreateRequest{Prompt: strings.Repeat("a", videojob.MaxPromptLength+1)}},
		{"duration above range", videojob.CreateRequest{Prompt: "x", DurationSeconds: vpIntPtr(99)}},
		{"duration below range", videojob.CreateRequest{Prompt: "x", DurationSeconds: vpIntPtr(1)}},
		{"unknown aspect ratio", videojob.CreateRequest{Prompt: "x", AspectRatio: "5:4"}},
		{"unknown resolution", videojob.CreateRequest{Prompt: "x", Resolution: "4k"}},
		{"audio not offered", videojob.CreateRequest{Prompt: "x", GenerateAudio: func() *bool { b := true; return &b }()}},
		{"image_url over http", videojob.CreateRequest{Prompt: "x", ImageURL: "http://example.com/frame.png"}},
		{"image_url loopback", videojob.CreateRequest{Prompt: "x", ImageURL: "https://127.0.0.1/frame.png"}},
		{"image_url private", videojob.CreateRequest{Prompt: "x", ImageURL: "https://10.1.2.3/frame.png"}},
		{"image_url link-local", videojob.CreateRequest{Prompt: "x", ImageURL: "https://169.254.169.254/latest/meta-data"}},
		{"image_url with credentials", videojob.CreateRequest{Prompt: "x", ImageURL: "https://user:pw@example.com/frame.png"}},
		{"image_url not a url", videojob.CreateRequest{Prompt: "x", ImageURL: "file:///etc/passwd"}},
		{"image_b64 not base64", videojob.CreateRequest{Prompt: "x", ImageB64: "not base64 !!"}},
		{"image_b64 not an image", videojob.CreateRequest{Prompt: "x", ImageB64: "aGVsbG8gd29ybGQ="}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newVideoHarness(t, videoHarnessOpts{})
			_, _, err := h.create(t, "job-1", tc.req)
			require.Error(t, err)
			require.ErrorIs(t, err, videojob.ErrInvalidRequest)
			creates, _ := h.vendor.counts()
			require.Zero(t, creates)
			require.Nil(t, h.storedTask(t, "job-1"))

			// a refused request leaves the session free for a valid job
			_, _, err = h.create(t, "job-2", videojob.CreateRequest{Prompt: "a valid prompt"})
			require.NoError(t, err)
		})
	}
}

func TestVideoProviderCreateForwardsFirstFrame(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})

	receipt, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "animate this", ImageURL: "https://images.example.com/frame.png"})
	require.NoError(t, err)
	require.True(t, receipt.Resolved.HasImage)

	h.vendor.mu.Lock()
	defer h.vendor.mu.Unlock()
	content, _ := h.vendor.lastCreate["content"].([]any)
	require.Len(t, content, 2)
	frame := content[1].(map[string]any)
	require.Equal(t, "image_url", frame["type"])
	require.Equal(t, "first_frame", frame["role"])
	require.Equal(t, "https://images.example.com/frame.png", frame["image_url"].(map[string]any)["url"])
}

func TestVideoProviderCreateStaleTimestampRefused(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})
	job := videojob.CreateRequest{Prompt: "a paper boat"}

	for _, ts := range []time.Time{time.Now().Add(-11 * time.Minute), time.Now().Add(11 * time.Minute)} {
		_, err := h.handle(t, craftedVideoCreate(t, h.userKey, h.sessionID, "job-1", job, ts))
		require.ErrorIs(t, err, ErrVideoJobStale)
	}
	creates, _ := h.vendor.counts()
	require.Zero(t, creates)

	// the same hand-built request with a fresh timestamp is accepted, so the
	// refusal above is the timestamp and not the crafting
	resp, err := h.handle(t, craftedVideoCreate(t, h.userKey, h.sessionID, "job-1", job, time.Now()))
	require.NoError(t, err)
	var receipt videojob.Receipt
	require.NoError(t, json.Unmarshal(h.open(t, resp, "req-create"), &receipt))
	require.Equal(t, "job-1", receipt.ClientJobID)
}

func TestVideoProviderCreateVendorRefusalNotStored(t *testing.T) {
	health := &fakeHealthTracker{}
	h := newVideoHarness(t, videoHarnessOpts{health: health})
	h.vendor.set(func(v *fakeVideoVendor) {
		v.createCode = http.StatusBadRequest
		v.createBody = `{"error":{"code":"InputTextSensitiveContentDetected","message":"the prompt was rejected"}}`
	})

	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "something the vendor refuses"})
	require.Error(t, err)
	var vendorErr *aiengine.VideoVendorError
	require.ErrorAs(t, err, &vendorErr)
	require.Equal(t, http.StatusBadRequest, vendorErr.StatusCode)
	// the plaintext error the TCP handler relays names the vendor's reason
	require.Contains(t, err.Error(), "the prompt was rejected")
	require.Contains(t, err.Error(), "InputTextSensitiveContentDetected")
	require.NotContains(t, err.Error(), videoTestVendorKey)
	require.Nil(t, h.storedTask(t, "job-1"), "a refused job must not occupy the session")
	// a vendor's refusal of the request is the client's fault, not the backend's
	require.Equal(t, 0, health.failures)
	require.Equal(t, 0, health.successes)

	h.vendor.set(func(v *fakeVideoVendor) { v.createCode = 0 })
	receipt, _, err := h.create(t, "job-2", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)
	require.Equal(t, "job-2", receipt.ClientJobID)
	creates, _ := h.vendor.counts()
	require.Equal(t, 2, creates)
}

func TestVideoProviderCreateVendorServerErrorCountsAsFailure(t *testing.T) {
	health := &fakeHealthTracker{}
	h := newVideoHarness(t, videoHarnessOpts{health: health})
	h.vendor.set(func(v *fakeVideoVendor) {
		v.createCode = http.StatusInternalServerError
		v.createBody = `{"error":{"code":"InternalServiceError","message":"try later"}}`
	})

	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.Error(t, err)
	require.Nil(t, h.storedTask(t, "job-1"))
	require.Equal(t, 1, health.failures)
	require.Equal(t, http.StatusInternalServerError, health.lastFailStatus)
}

func TestVideoProviderCreateVendorWithoutTaskIDRefused(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})
	h.vendor.set(func(v *fakeVideoVendor) {
		v.createCode = http.StatusOK
		v.createBody = `{"status":"ok"}`
	})

	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "no task id")
	require.Nil(t, h.storedTask(t, "job-1"))
}

func TestVideoProviderCreateNonVideoModelRefused(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{apiType: "openai", modelName: "llama"})

	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not serve video jobs")
	creates, _ := h.vendor.counts()
	require.Zero(t, creates)
}

func TestVideoProviderCreateExpiredSessionRefused(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{endsAt: time.Now().Add(-time.Minute)})

	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "session expired")
	creates, _ := h.vendor.counts()
	require.Zero(t, creates)
}

func TestVideoProviderCreateSessionOfAnotherProviderRefused(t *testing.T) {
	other := common.HexToAddress("0x2222222222222222222222222222222222222222")
	h := newVideoHarness(t, videoHarnessOpts{providerAddr: &other})

	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not bound to this provider")
	creates, _ := h.vendor.counts()
	require.Zero(t, creates)
}

func TestVideoProviderCreateValidationRefused(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})

	_, _, err := h.create(t, strings.Repeat("j", 65), videojob.CreateRequest{Prompt: "a valid prompt"})
	require.ErrorIs(t, err, ErrValidation)
	creates, _ := h.vendor.counts()
	require.Zero(t, creates)
}

func TestVideoProviderCreateWrongSignerRefused(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})
	other, err := crypto.GenerateKey()
	require.NoError(t, err)

	_, _, err = h.createWith(t, other, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid signature")
	creates, _ := h.vendor.counts()
	require.Zero(t, creates)
	require.Nil(t, h.storedTask(t, "job-1"))
}

// ---------------------------------------------------------------------------
// session.video.status
// ---------------------------------------------------------------------------

func TestVideoProviderStatusRelaysVendorStatus(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})
	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)

	h.vendor.set(func(v *fakeVideoVendor) { v.statusBody = bytePlusSucceededBody })
	update, err := h.status(t, "job-1")
	require.NoError(t, err)
	require.Equal(t, videojob.StatusSucceeded, update.Status)
	require.Equal(t, "https://cdn.example.com/clip.mp4", update.VideoURL)
	require.NotNil(t, update.ExpiresAt)
	require.True(t, update.ExpiresAt.Equal(time.Unix(1700000000, 0).Add(videojob.BytePlusURLLifetime)))
	require.Nil(t, update.Error)

	h.vendor.mu.Lock()
	require.Equal(t, 1, h.vendor.statusReads)
	require.Equal(t, "/contents/generations/tasks/"+videoTestBytePlusID, h.vendor.lastStatusPath)
	require.Equal(t, "Bearer "+videoTestVendorKey, h.vendor.lastStatusAuth)
	h.vendor.mu.Unlock()
}

func TestVideoProviderStatusRelaysVendorFailure(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})
	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)

	h.vendor.set(func(v *fakeVideoVendor) {
		v.statusBody = `{"id":"cgt-test-1","status":"failed","error":{"code":"OutputVideoSensitiveContentDetected","message":"blocked"}}`
	})
	update, err := h.status(t, "job-1")
	require.NoError(t, err)
	require.Equal(t, videojob.StatusFailed, update.Status)
	require.NotNil(t, update.Error)
	require.Equal(t, "OutputVideoSensitiveContentDetected", update.Error.Code)
	require.Equal(t, "blocked", update.Error.Message)
}

func TestVideoProviderStatusUnknownJob(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})

	// no job on the session yet
	_, err := h.status(t, "job-1")
	require.ErrorIs(t, err, ErrNoVideoJob)

	_, _, err = h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)

	// another job id on the same session
	_, err = h.status(t, "job-2")
	require.ErrorIs(t, err, ErrNoVideoJob)
	_, reads := h.vendor.counts()
	require.Zero(t, reads)
}

func TestVideoProviderStatusAfterSessionEnded(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})
	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)

	// the render outlives the session: status has no expiry gate
	h.setSessionEnd(t, time.Now().Add(-time.Hour))
	h.vendor.set(func(v *fakeVideoVendor) { v.statusBody = bytePlusSucceededBody })

	update, err := h.status(t, "job-1")
	require.NoError(t, err)
	require.Equal(t, videojob.StatusSucceeded, update.Status)
	require.Equal(t, "https://cdn.example.com/clip.mp4", update.VideoURL)

	// create stays gated by the session's end
	_, _, err = h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "session expired")
}

func TestVideoProviderStatusCached(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})
	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)

	h.vendor.set(func(v *fakeVideoVendor) { v.statusBody = `{"id":"cgt-test-1","status":"running"}` })
	update, err := h.status(t, "job-1")
	require.NoError(t, err)
	require.Equal(t, videojob.StatusRunning, update.Status)

	// a second poll inside the interval reuses the relayed status
	h.vendor.set(func(v *fakeVideoVendor) { v.statusBody = bytePlusSucceededBody })
	update, err = h.status(t, "job-1")
	require.NoError(t, err)
	require.Equal(t, videojob.StatusRunning, update.Status)
	_, reads := h.vendor.counts()
	require.Equal(t, 1, reads, "a poll within the interval must not reach the vendor")

	// once the cached entry is older than the interval the vendor is read again
	h.receiver.video.mu.Lock()
	for id, c := range h.receiver.video.statusCache {
		c.at = c.at.Add(-videoStatusMinInterval - time.Second)
		h.receiver.video.statusCache[id] = c
	}
	h.receiver.video.mu.Unlock()

	update, err = h.status(t, "job-1")
	require.NoError(t, err)
	require.Equal(t, videojob.StatusSucceeded, update.Status)
	_, reads = h.vendor.counts()
	require.Equal(t, 2, reads)
}

func TestVideoProviderStatusVendorErrorSurfaced(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})
	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)

	h.vendor.set(func(v *fakeVideoVendor) { v.statusBody = `not json` })
	_, err = h.status(t, "job-1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "byteplus status")

	// a failed read is not cached: the next poll reaches the vendor again
	h.vendor.set(func(v *fakeVideoVendor) { v.statusBody = bytePlusSucceededBody })
	update, err := h.status(t, "job-1")
	require.NoError(t, err)
	require.Equal(t, videojob.StatusSucceeded, update.Status)
	_, reads := h.vendor.counts()
	require.Equal(t, 2, reads)
}

func TestVideoProviderStatusStaleTimestampRefused(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})
	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)
	h.vendor.set(func(v *fakeVideoVendor) { v.statusBody = bytePlusSucceededBody })

	_, err = h.handle(t, craftedVideoStatus(t, h.userKey, h.sessionID, "job-1", time.Now().Add(-11*time.Minute)))
	require.ErrorIs(t, err, ErrVideoJobStale)
	_, reads := h.vendor.counts()
	require.Zero(t, reads)

	resp, err := h.handle(t, craftedVideoStatus(t, h.userKey, h.sessionID, "job-1", time.Now()))
	require.NoError(t, err)
	var update videojob.Update
	require.NoError(t, json.Unmarshal(h.open(t, resp, "req-status"), &update))
	require.Equal(t, videojob.StatusSucceeded, update.Status)
}

func TestVideoProviderStatusWrongSignerRefused(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})
	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)
	h.vendor.set(func(v *fakeVideoVendor) { v.statusBody = bytePlusSucceededBody })

	other, err := crypto.GenerateKey()
	require.NoError(t, err)
	_, err = h.statusWith(t, other, "job-1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid signature")
	_, reads := h.vendor.counts()
	require.Zero(t, reads)
}

func TestVideoProviderStatusSessionOfAnotherProviderRefused(t *testing.T) {
	other := common.HexToAddress("0x2222222222222222222222222222222222222222")
	h := newVideoHarness(t, videoHarnessOpts{providerAddr: &other})

	// a record left by some other path must not be readable through a node
	// that does not serve the session
	raw, err := json.Marshal(&providerVideoTask{
		SessionID: h.sessionID.Hex(), JobID: "job-1", ModelID: h.modelID.Hex(),
		Vendor: videojob.VendorBytePlus, VendorTaskID: videoTestBytePlusID,
	})
	require.NoError(t, err)
	require.NoError(t, h.ss.PutVideoTask(h.sessionID.Hex(), "job-1", raw))

	_, err = h.status(t, "job-1")
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "not bound to this provider"), err.Error())
	_, reads := h.vendor.counts()
	require.Zero(t, reads)
}

func TestVideoProviderUnknownMethod(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})
	_, err := h.handle(t, &morrpc.RPCMessage{ID: "x", Method: "session.video.cancel", Params: json.RawMessage(`{}`)})
	require.True(t, errors.Is(err, ErrUnknownMethod))
}

// ---------------------------------------------------------------------------
// session length: on-chain length first, time left as the fallback
// ---------------------------------------------------------------------------

// A session whose on-chain length covers the quote is accepted even when
// little of it is left: the length is what the user paid.
func TestVideoProviderCreateChainLengthCoversQuote(t *testing.T) {
	now := time.Now()
	chain := &videoTestChainSessions{session: chainSession(now.Add(-5000*time.Second), now.Add(60*time.Second), 0)}
	h := newVideoHarness(t, videoHarnessOpts{endsAt: now.Add(60 * time.Second), chain: chain})

	receipt, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)
	require.EqualValues(t, 844, receipt.QuoteSeconds)
	chain.mu.Lock()
	require.Equal(t, 1, chain.reads)
	chain.mu.Unlock()
}

// The on-chain length is checked with a 10 s tolerance and no grace: a
// session opened for less than the quote is refused even with time left.
func TestVideoProviderCreateChainLengthTooShortRefused(t *testing.T) {
	now := time.Now()
	for _, tc := range []struct {
		length int64
		ok     bool
	}{
		{length: 833, ok: false}, // 833 + 10 < 844
		{length: 834, ok: true},  // 834 + 10 = 844
		{length: 300, ok: false},
	} {
		chain := &videoTestChainSessions{session: chainSession(now, now.Add(time.Duration(tc.length)*time.Second), 0)}
		h := newVideoHarness(t, videoHarnessOpts{chain: chain}) // cached session ends in 2 h

		_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
		creates, _ := h.vendor.counts()
		if tc.ok {
			require.NoError(t, err, "length %d", tc.length)
			require.Equal(t, 1, creates)
			continue
		}
		require.Error(t, err, "length %d", tc.length)
		require.Contains(t, err.Error(), "needs a 844-second session")
		require.Contains(t, err.Error(), "seconds long")
		require.Zero(t, creates)
		require.Nil(t, h.storedTask(t, "job-1"))
	}
}

// When the chain cannot be read the time left (+120 s grace) decides.
// A failed chain read refuses the job: the cached session cannot show a
// close, so a refunded session would otherwise pass. A chain answer without
// timestamps falls back to the time left.
func TestVideoProviderCreateChainReadErrorRefuses(t *testing.T) {
	chain := &videoTestChainSessions{err: errors.New("rpc down")}
	h := newVideoHarness(t, videoHarnessOpts{endsAt: time.Now().Add(3600 * time.Second), chain: chain})
	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot read the session from chain")
	creates, _ := h.vendor.counts()
	require.Zero(t, creates, "no vendor call")

	chain = &videoTestChainSessions{session: &structs.Session{}}
	h = newVideoHarness(t, videoHarnessOpts{endsAt: time.Now().Add(700 * time.Second), chain: chain})
	_, _, err = h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "seconds left")
}

// A session the user already closed on chain has paid only for the seconds
// it ran, yet its endsAt - openedAt still covers the quote, so the provider
// starts (and pays for) a render nobody pays it for. The provider removes a
// closed session from its cache on the SessionClosed event, but
// isSessionValid then re-reads it from the chain without looking at
// closedAt, and checkVideoSessionLength reads the on-chain session (with
// ClosedAt) and ignores it too.
func TestVideoProviderCreateRefusesSessionClosedOnChain(t *testing.T) {

	now := time.Now()
	chain := &videoTestChainSessions{session: chainSession(now.Add(-30*time.Second), now.Add(3600*time.Second), now.Add(-20*time.Second).Unix())}
	h := newVideoHarness(t, videoHarnessOpts{chain: chain})

	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.Error(t, err, "a session closed on chain must not start a vendor task")
	creates, _ := h.vendor.counts()
	require.Zero(t, creates)
}

// The status cache is meant to stop a consumer polling hard from spending
// the vendor's read budget, but concurrent polls all miss it: the cache is
// filled only after the vendor answers and nothing collapses in-flight
// reads (no singleflight / per-session lock in SessionVideoStatus).
func TestVideoProviderStatusConcurrentPollsReadVendorOnce(t *testing.T) {

	h := newVideoHarness(t, videoHarnessOpts{})
	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)
	h.vendor.set(func(v *fakeVideoVendor) {
		v.statusBody = `{"id":"cgt-test-1","status":"running"}`
		v.statusDelay = 300 * time.Millisecond
	})

	const polls = 4
	msgs := make([]*morrpc.RPCMessage, polls)
	for i := range msgs {
		m, err := h.rpc.SessionVideoStatusRequest(h.sessionID, "job-1", h.userPrKey, "req-status")
		require.NoError(t, err)
		msgs[i] = m
	}
	var wg sync.WaitGroup
	errs := make([]error, polls)
	for i := range msgs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = h.c.Handle(context.Background(), *msgs[i], &lib.LoggerMock{}, func(*morrpc.RpcResponse) error { return nil })
		}(i)
	}
	wg.Wait()
	for _, err := range errs {
		require.NoError(t, err)
	}
	_, reads := h.vendor.counts()
	require.Equal(t, 1, reads, "parallel polls within the interval must share one vendor read")
}

// ---------------------------------------------------------------------------
// served spec reaches the adapter
// ---------------------------------------------------------------------------

// The generate_audio switch follows the served spec: an audio model always
// sends it (vendors default it on), and a model without audio refuses
// generate_audio and never sends the switch.
func TestVideoProviderCreateAudioSwitchFollowsServedSpec(t *testing.T) {
	const model = "dreamina-seedance-2-5-260628"

	h := newVideoHarness(t, videoHarnessOpts{modelName: model})
	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)
	h.vendor.mu.Lock()
	require.Equal(t, false, h.vendor.lastCreate["generate_audio"], "an audio model must send the switch off explicitly")
	h.vendor.mu.Unlock()

	on := true
	h = newVideoHarness(t, videoHarnessOpts{modelName: model})
	receipt, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt", GenerateAudio: &on})
	require.NoError(t, err)
	require.True(t, receipt.Resolved.GenerateAudio)
	h.vendor.mu.Lock()
	require.Equal(t, true, h.vendor.lastCreate["generate_audio"])
	h.vendor.mu.Unlock()

	h = newVideoHarness(t, videoHarnessOpts{modelName: "seedance-1-0-pro-250528"})
	_, _, err = h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt", GenerateAudio: &on})
	require.ErrorIs(t, err, videojob.ErrInvalidRequest)
	_, _, err = h.create(t, "job-2", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)
	h.vendor.mu.Lock()
	require.NotContains(t, h.vendor.lastCreate, "generate_audio", "a model without audio has no switch to send")
	h.vendor.mu.Unlock()
}
