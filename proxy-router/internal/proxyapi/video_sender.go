package proxyapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	msgs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi/morrpcmessage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/videojob"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// The provider answers a create once the vendor accepted the task; its
	// own vendor call is bounded well inside this.
	videoCreateReadTimeout = 60 * time.Second
	videoStatusReadTimeout = 30 * time.Second
	// A receipt or a status is a small JSON document.
	videoMaxResponseBytes = 4 << 20
	// videoResponseMaxSkew bounds the age of a signed response, so an old
	// answer cannot be replayed for a new request.
	videoResponseMaxSkew = 10 * time.Minute
)

// ErrProviderRefused is a plaintext refusal from the provider (bad request,
// quote longer than the session, a second job on the session, a vendor
// refusal). Retrying the same request does not help.
var ErrProviderRefused = errors.New("provider refused the video request")

// videoProvider resolves the provider of a session without an expiry check:
// a status read is valid after the session ended.
func (p *ProxyServiceSender) videoProvider(ctx context.Context, sessionID common.Hash) (bool, *storages.User, error) {
	session, err := p.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		return false, nil, ErrSessionNotFound
	}
	provider, err := p.sessionStorage.GetUser(session.ProviderAddr().Hex())
	if err != nil {
		return false, nil, fmt.Errorf("error reading provider: %w", err)
	}
	if provider == nil {
		return false, nil, ErrProviderNotFound
	}
	return session.IsTee(), provider, nil
}

// SendVideoJobCreate sends a video job on the session to its provider and
// returns the receipt. jobID makes a resend of the same job safe: the
// provider answers it with the receipt it already issued.
func (p *ProxyServiceSender) SendVideoJobCreate(ctx context.Context, sessionID common.Hash, jobID string, job *videojob.CreateRequest) (*videojob.Receipt, error) {
	requestID := lib.RequestIDFromContext(ctx)
	if requestID == "" {
		requestID = lib.GenerateRequestID()
	}

	isTee, provider, err := p.videoProvider(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if err := p.verifyTEEAttestation(ctx, provider.Url, provider.Addr, isTee); err != nil {
		return nil, lib.WrapError(ErrTeeAttestation, err)
	}

	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return nil, ErrMissingPrKey
	}
	req, err := p.morRPC.SessionVideoCreateRequest(sessionID, jobID, job, prKey, requestID)
	if err != nil {
		return nil, lib.WrapError(ErrCreateReq, err)
	}

	plain, err := p.videoRPC(provider, req, prKey, videoCreateReadTimeout)
	if err != nil {
		return nil, err
	}

	var receipt videojob.Receipt
	if err := json.Unmarshal(plain, &receipt); err != nil {
		return nil, lib.WrapError(ErrInvalidResponse, err)
	}
	if receipt.Object != videojob.ObjectJob || receipt.VendorTaskID == "" || receipt.ClientJobID != jobID {
		return nil, lib.WrapError(ErrInvalidResponse, fmt.Errorf("not a receipt for job %s", jobID))
	}
	return &receipt, nil
}

// SendVideoJobStatus asks the session's provider for the vendor status of
// one of the session's jobs.
func (p *ProxyServiceSender) SendVideoJobStatus(ctx context.Context, sessionID common.Hash, jobID string) (*videojob.Update, error) {
	requestID := lib.RequestIDFromContext(ctx)
	if requestID == "" {
		requestID = lib.GenerateRequestID()
	}

	_, provider, err := p.videoProvider(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return nil, ErrMissingPrKey
	}
	req, err := p.morRPC.SessionVideoStatusRequest(sessionID, jobID, prKey, requestID)
	if err != nil {
		return nil, lib.WrapError(ErrCreateReq, err)
	}

	plain, err := p.videoRPC(provider, req, prKey, videoStatusReadTimeout)
	if err != nil {
		return nil, err
	}

	var update videojob.Update
	if err := json.Unmarshal(plain, &update); err != nil {
		return nil, lib.WrapError(ErrInvalidResponse, err)
	}
	if update.Status == "" {
		return nil, lib.WrapError(ErrInvalidResponse, ErrEmpty)
	}
	if update.JobID != jobID {
		return nil, lib.WrapError(ErrInvalidResponse, fmt.Errorf("status is for job %q, not %s", update.JobID, jobID))
	}
	return &update, nil
}

// videoRPC sends one request and returns the decrypted, signature-checked
// payload of the single response.
func (p *ProxyServiceSender) videoRPC(provider *storages.User, req *msgs.RPCMessage, prKey lib.HexString, readTimeout time.Duration) ([]byte, error) {
	msg, code, err := p.rpcRequestWithReadTimeout(provider.Url, req, readTimeout, videoMaxResponseBytes)
	if err != nil {
		return nil, lib.WrapError(ErrProvider, fmt.Errorf("code: %d, error: %s", code, err))
	}
	if msg.Error != nil {
		return nil, fmt.Errorf("%w: %s", ErrProviderRefused, msg.Error.Message)
	}
	if msg.Result == nil {
		return nil, lib.WrapError(ErrInvalidResponse, ErrEmpty)
	}

	var res msgs.SessionVideoRes
	if err := json.Unmarshal(*msg.Result, &res); err != nil {
		return nil, lib.WrapError(ErrInvalidResponse, err)
	}
	signature := res.Signature
	res.Signature = lib.HexString{}

	providerPubKey, err := lib.StringToHexString(provider.PubKey)
	if err != nil {
		return nil, lib.WrapError(ErrInvalidResponse, err)
	}
	if !p.validateMsgSignature(res, signature, providerPubKey) {
		return nil, ErrInvalidSig
	}
	if age := time.Since(time.UnixMilli(int64(res.Timestamp))); age > videoResponseMaxSkew || age < -videoResponseMaxSkew {
		return nil, lib.WrapError(ErrInvalidResponse, fmt.Errorf("response timestamp is %s off", age.Round(time.Second)))
	}

	plain, err := lib.DecryptString(res.Message, prKey.Hex())
	if err != nil {
		return nil, lib.WrapError(ErrDecrFailed, err)
	}
	return []byte(plain), nil
}
