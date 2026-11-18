package proxyapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	sessionrepo "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/session"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"io"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	morrpc "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi/morrpcmessage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/videojob"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

// The pong carries the guide of every configured video model, even with the
// health checker off: without it no consumer can size a session here.
func TestVideoGuideInPongWithoutHealthChecker(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})
	nonce := lib.HexString("0x01020304")
	req, err := h.rpc.PingRequest("ping-1", h.userPrKey, nonce)
	require.NoError(t, err)

	resp, err := h.handle(t, req)
	require.NoError(t, err)
	require.NotNil(t, resp.Result)

	var pong morrpc.PongRes
	require.NoError(t, json.Unmarshal(*resp.Result, &pong))
	var found *system.ModelHealthReport
	for i := range pong.Models {
		if strings.EqualFold(pong.Models[i].ModelID, h.modelID.Hex()) {
			found = &pong.Models[i]
		}
	}
	require.NotNil(t, found, "the video model is in the pong")
	require.Equal(t, system.ModelHealthStatusSkipped, found.Status)
	require.NotNil(t, found.Video)
	require.Equal(t, videojob.ObjectGuide, found.Video.Object)
}

// The pong guide is built from the current config, replacing whatever the
// checker cached at its last sweep (a reload may have changed the config).
func TestWithVideoGuidesKeepsAndFillsReports(t *testing.T) {
	h := newVideoHarness(t, videoHarnessOpts{})
	stale := &videojob.Guide{Object: videojob.ObjectGuide, Summary: "from an old sweep"}

	out := h.receiver.withVideoGuides([]system.ModelHealthReport{{ModelID: h.modelID.Hex(), Status: system.ModelHealthStatusSkipped, Video: stale}}, &lib.LoggerMock{})
	require.Len(t, out, 1)
	require.NotNil(t, out[0].Video)
	require.NotEqual(t, stale.Summary, out[0].Video.Summary)

	in := []system.ModelHealthReport{{ModelID: strings.ToUpper(h.modelID.Hex()[2:]), Status: system.ModelHealthStatusHealthy}}
	in[0].ModelID = "0x" + in[0].ModelID
	out = h.receiver.withVideoGuides(in, &lib.LoggerMock{})
	require.Len(t, out, 1)
	require.NotNil(t, out[0].Video)
	require.Nil(t, in[0].Video, "the checker's slice is not modified")
}

// A session closed before its end (refunded) does not get the clip through
// the relay; the status is bound to the job asked about.
func TestVideoProviderStatusWithholdsClipAfterEarlyClose(t *testing.T) {
	now := time.Now()
	chain := &videoTestChainSessions{session: chainSession(now.Add(-time.Minute), now.Add(time.Hour), 0)}
	h := newVideoHarness(t, videoHarnessOpts{chain: chain})
	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)

	// closed early: closedAt < endsAt
	chain.mu.Lock()
	chain.session = chainSession(now.Add(-time.Minute), now.Add(time.Hour), now.Unix())
	chain.mu.Unlock()
	h.vendor.set(func(v *fakeVideoVendor) { v.statusBody = bytePlusSucceededBody })

	update, err := h.status(t, "job-1")
	require.NoError(t, err)
	require.Equal(t, videojob.StatusCancelled, update.Status)
	require.Empty(t, update.VideoURL)
	require.Equal(t, "session_closed_early", update.Error.Code)
	require.Equal(t, "job-1", update.JobID)
}

// A session that ran to its end (the expiry close of a delivered clip) still
// gets its link.
func TestVideoProviderStatusServesClipAfterNormalEnd(t *testing.T) {
	now := time.Now()
	chain := &videoTestChainSessions{session: chainSession(now.Add(-time.Hour), now.Add(time.Hour), 0)}
	h := newVideoHarness(t, videoHarnessOpts{chain: chain})
	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)

	chain.mu.Lock()
	chain.session = chainSession(now.Add(-time.Hour), now.Add(-time.Minute), now.Add(-30*time.Second).Unix())
	chain.mu.Unlock()
	h.vendor.set(func(v *fakeVideoVendor) { v.statusBody = bytePlusSucceededBody })

	update, err := h.status(t, "job-1")
	require.NoError(t, err)
	require.Equal(t, videojob.StatusSucceeded, update.Status)
	require.Equal(t, "https://cdn.example.com/clip.mp4", update.VideoURL)
	require.Equal(t, "job-1", update.JobID)

	// A final status is cached: no second vendor read.
	_, err = h.status(t, "job-1")
	require.NoError(t, err)
	_, reads := h.vendor.counts()
	require.Equal(t, 1, reads)
}

// A chat prompt sent to a video model is the caller's mistake and does not
// count against the model's health.
func TestChatOnVideoModelDoesNotMarkItUnhealthy(t *testing.T) {
	tracker := &countingHealth{}
	r := &ProxyReceiver{modelHealth: tracker}
	model := common.HexToHash("0x01")

	r.trackPromptResult(model, aiengine.ErrVideoNotSupported, 0, false)
	r.trackPromptResult(model, lib.WrapError(errors.New("failed to prompt"), aiengine.ErrVideoNotSupported), 0, false)
	require.Zero(t, tracker.failures)

	r.trackPromptResult(model, errors.New("dial tcp: refused"), 0, false)
	require.Equal(t, 1, tracker.failures)
}

// Account-level vendor refusals (bad key, no balance) count against the
// model; a refusal of the request itself does not.
func TestVideoVendorAccountErrorsCountAsFailures(t *testing.T) {
	tracker := &countingHealth{}
	r := &ProxyReceiver{modelHealth: tracker}
	model := common.HexToHash("0x01")

	for _, status := range []int{400, 413, 422} {
		r.trackVideoResult(model, &aiengine.VideoVendorError{StatusCode: status})
	}
	require.Zero(t, tracker.failures)
	for _, status := range []int{401, 402, 403, 429, 500} {
		r.trackVideoResult(model, &aiengine.VideoVendorError{StatusCode: status})
	}
	require.Equal(t, 5, tracker.failures)
}

type countingHealth struct {
	failures  int
	successes int
}

func (c *countingHealth) ReportPromptSuccess(common.Hash)      { c.successes++ }
func (c *countingHealth) ReportPromptFailure(common.Hash, int) { c.failures++ }
func (c *countingHealth) ReportTeeFailure(common.Hash)         {}

func TestPublicAddrRefusesTranslatedIPv4(t *testing.T) {
	for _, bad := range []string{"https://[64:ff9b::a00:5]/x", "https://[2002:a00:5::1]/x"} {
		require.ErrorIs(t, videojob.ValidateRemoteURL(bad), videojob.ErrRemoteURL, bad)
	}
}

// Refusals a retry can cure (a wallet switched away and back, a contract
// provider whose chain reads fail) do not end the job.
func TestVideoJobSignatureRefusalsAreNotPermanent(t *testing.T) {
	require.False(t, permanentStatusRefusal(fmt.Errorf("%w: invalid signature", ErrProviderRefused)))
	require.False(t, permanentStatusRefusal(fmt.Errorf("%w: session not bound to this provider", ErrProviderRefused)))
	require.True(t, permanentStatusRefusal(fmt.Errorf("%w: %s", ErrProviderRefused, ErrNoVideoJob)))
	require.False(t, permanentStatusRefusal(ErrNoVideoJob), "only a provider's answer counts")
}

// An unsigned or tampered models list carries no guide the consumer prices
// a session with.
func TestPongModelsSignatureGatesGuides(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	rpc := morrpc.NewMorRpc()

	models := []system.ModelHealthReport{{ModelID: "0x01", Status: system.ModelHealthStatusSkipped, Video: &videojob.Guide{Object: videojob.ObjectGuide}}}
	res, err := rpc.PongResponce("1", crypto.FromECDSA(key), lib.HexString{0x01}, "v", models)
	require.NoError(t, err)

	var pong morrpc.PongRes
	require.NoError(t, json.Unmarshal(*res.Result, &pong))
	require.True(t, pongModelsSigned(*res.Result, pong.ModelsSignature, addr))
	require.False(t, pongModelsSigned(*res.Result, nil, addr), "unsigned")
	other, err := crypto.GenerateKey()
	require.NoError(t, err)
	require.False(t, pongModelsSigned(*res.Result, pong.ModelsSignature, crypto.PubkeyToAddress(other.PublicKey)), "another signer")
	tampered := json.RawMessage(strings.Replace(string(*res.Result), `"skipped"`, `"healthy"`, 1))
	require.False(t, pongModelsSigned(tampered, pong.ModelsSignature, addr), "tampered")
}

// The consumer's expiry close lands about a minute after endsAt; one a few
// seconds before it (clock skew) is not an early close either.
func TestVideoProviderStatusEarlyCloseGrace(t *testing.T) {
	now := time.Now()
	chain := &videoTestChainSessions{session: chainSession(now.Add(-time.Hour), now.Add(time.Hour), 0)}
	h := newVideoHarness(t, videoHarnessOpts{chain: chain})
	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)

	ends := now.Add(-time.Minute)
	chain.mu.Lock()
	chain.session = chainSession(now.Add(-time.Hour), ends, ends.Add(-20*time.Second).Unix())
	chain.mu.Unlock()
	h.vendor.set(func(v *fakeVideoVendor) { v.statusBody = bytePlusSucceededBody })

	update, err := h.status(t, "job-1")
	require.NoError(t, err)
	require.Equal(t, videojob.StatusSucceeded, update.Status)
	require.NotEmpty(t, update.VideoURL)
}

// One failed chain read is retried before the create is refused.
func TestVideoProviderCreateRetriesChainRead(t *testing.T) {
	now := time.Now()
	chain := &flakyChain{videoTestChainSessions: videoTestChainSessions{session: chainSession(now, now.Add(time.Hour), 0)}, failures: 1}
	h := newVideoHarness(t, videoHarnessOpts{})
	h.receiver.service = chain

	_, _, err := h.create(t, "job-1", videojob.CreateRequest{Prompt: "a valid prompt"})
	require.NoError(t, err)
	require.Equal(t, 2, chain.calls)
}

type flakyChain struct {
	videoTestChainSessions
	failures int
	calls    int
}

func (c *flakyChain) GetSession(ctx context.Context, id common.Hash) (*structs.Session, error) {
	c.calls++
	if c.calls <= c.failures {
		return nil, errors.New("429 too many requests")
	}
	return c.videoTestChainSessions.GetSession(ctx, id)
}

// --- provider answers nobody can use --------------------------------------

// rawResponder answers one TCP request with body, whatever the request was.
func rawResponder(t *testing.T, body string) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
				_, _ = conn.Read(make([]byte, 64<<10))
				_, _ = io.WriteString(conn, body)
			}(conn)
		}
	}()
	return ln.Addr().String()
}

func newRawSender(t *testing.T) *ProxyServiceSender {
	t.Helper()
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	ss := storages.NewSessionStorage(storages.NewTestStorage())
	repo := sessionrepo.NewSessionRepositoryCached(ss, nil, nil, lib.NewTestLogger())
	return NewProxySender(big.NewInt(1), e2eKeyProvider{key: lib.HexString(crypto.FromECDSA(key))}, nil, ss, repo, 90*time.Second, 3, 20, lib.NewTestLogger())
}

// The JSON literal null decodes into a nil response without an error; it
// must come back as an error, not as a nil pointer the caller dereferences.
func TestRPCRequestNullResponseIsAnError(t *testing.T) {
	p := newRawSender(t)
	for _, body := range []string{"null", "null\n", "  null  \n"} {
		addr := rawResponder(t, body)
		req := &morrpc.RPCMessage{ID: "1", Method: "network.ping", Params: json.RawMessage(`{}`)}
		var msg *morrpc.RpcResponse
		require.NotPanics(t, func() {
			var err error
			msg, _, err = p.rpcRequestWithReadTimeout(addr, req, 2*time.Second, 1<<20)
			require.ErrorIs(t, err, ErrDecode)
			require.ErrorIs(t, err, ErrEmpty)
		})
		require.Nil(t, msg)
	}
}

// An error-only answer to network.ping (what a provider sends when the ping
// fails validation, or what any endpoint can send) is an error, not a
// dereference of a missing result. Ping runs in bare goroutines on the
// consumer, so a panic here would take the node down.
func TestPingErrorOnlyResponseIsAnError(t *testing.T) {
	p := newRawSender(t)
	for _, tc := range []struct {
		body string
		want error
	}{
		{`{"id":"1","error":{"code":1,"message":"boom"}}` + "\n", ErrResponseErr},
		{`{"id":"1"}` + "\n", ErrInvalidResponse},
	} {
		addr := rawResponder(t, tc.body)
		require.NotPanics(t, func() {
			_, _, _, err := p.Ping(context.Background(), addr, common.HexToAddress("0x01"))
			require.ErrorIs(t, err, tc.want, tc.body)
		})
	}
}
