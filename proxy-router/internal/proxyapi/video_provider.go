package proxyapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	m "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi/morrpcmessage"
	sessionrepo "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/session"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/videojob"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// videoQuoteGraceSeconds covers the time between the consumer opening a
	// session for exactly the quote and the job reaching this provider, when
	// the session's on-chain length cannot be read.
	videoQuoteGraceSeconds = 120
	// videoLengthToleranceSeconds absorbs the rounding between the duration a
	// consumer asks for and the endsAt the contract derives from the stake.
	videoLengthToleranceSeconds = 10
	// videoRequestMaxSkew bounds how old (or how far ahead) a signed video
	// request may be.
	videoRequestMaxSkew = 10 * time.Minute
	// videoStatusMinInterval is how long a relayed vendor status is reused,
	// so a consumer polling hard cannot spend the vendor's read budget.
	videoStatusMinInterval = 5 * time.Second
	// videoTerminalStatusTTL keeps a final status: it will not change.
	videoTerminalStatusTTL = 24 * time.Hour
	// videoEarlyCloseGraceSeconds: a close this close to endsAt is the
	// consumer's expiry close (clock skew), not an early one.
	videoEarlyCloseGraceSeconds = 60
	// videoChainReadAttempts rides out a single failed RPC read before a
	// create is refused for not being able to see the session.
	videoChainReadAttempts = 3
)

var (
	ErrVideoJobStale = errors.New("video request timestamp is outside the accepted window")
	ErrNoVideoJob    = errors.New("no such video job on this session")
)

// providerVideoTask is the provider's record of one vendor task a session
// paid for. It is the idempotency memory for a retried accept, counts its
// quote against the session's length, and is what session.video.status
// reads.
type providerVideoTask struct {
	SessionID    string            `json:"session_id"`
	JobID        string            `json:"job_id"`
	ModelID      string            `json:"model_id"`
	Vendor       string            `json:"vendor"`
	VendorTaskID string            `json:"vendor_task_id"`
	StatusURL    string            `json:"status_url"`
	QuoteSeconds int64             `json:"quote_seconds"`
	Resolved     videojob.Resolved `json:"resolved"`
	CreatedAt    int64             `json:"created_at"`
}

type cachedVideoStatus struct {
	at     time.Time
	update videojob.Update
}

// fresh reports whether a cached status can be served again: a final status
// for a day, a running one for videoStatusMinInterval.
func (c cachedVideoStatus) fresh(now time.Time) bool {
	if videojob.IsTerminal(c.update.Status) {
		return now.Sub(c.at) < videoTerminalStatusTTL
	}
	return now.Sub(c.at) < videoStatusMinInterval
}

type videoProviderState struct {
	mu          sync.Mutex
	statusCache map[string]cachedVideoStatus
	httpClient  *http.Client
}

func videoRequestFresh(timestampMs uint64, now time.Time) bool {
	ts := time.UnixMilli(int64(timestampMs))
	return ts.After(now.Add(-videoRequestMaxSkew)) && ts.Before(now.Add(videoRequestMaxSkew))
}

func (s *MORRPCController) sessionVideoCreate(ctx context.Context, msg m.RPCMessage, sendResponse SendResponse, sourceLog lib.ILogger) error {
	var req m.SessionVideoCreateReq
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return lib.WrapError(ErrUnmarshal, err)
	}
	if err := s.validator.Struct(req); err != nil {
		return lib.WrapError(ErrValidation, err)
	}
	sourceLog = sourceLog.With("session_id", req.SessionID.Hex(), "video_job_id", req.JobID)

	session, err := s.isSessionValid(ctx, req.SessionID, req.Timestamp)
	if err != nil {
		return err
	}

	user, pubKeyHex, err := s.sessionUserPubKey(session.UserAddr())
	if err != nil {
		return err
	}

	sig := req.Signature
	req.Signature = lib.HexString{}
	if !s.morRpc.VerifySignature(req, sig, pubKeyHex, sourceLog) {
		err := fmt.Errorf("invalid signature")
		sourceLog.Error(err)
		return err
	}
	if !videoRequestFresh(req.Timestamp, time.Now()) {
		return ErrVideoJobStale
	}

	// One request per session at a time: a retried accept waits for the
	// first and then reads its record instead of creating a second task, and
	// two jobs cannot both spend the same session time.
	if err := s.sessionSema.Acquire(ctx, req.SessionID); err != nil {
		return fmt.Errorf("request cancelled while waiting in queue: %w", err)
	}
	defer s.sessionSema.Release(req.SessionID)

	res, err := s.service.SessionVideoCreate(ctx, msg.ID, user.PubKey, session, &req, sourceLog)
	if err != nil {
		sourceLog.Warnf("video job refused: %s", err)
		return err
	}
	return sendResponse(res)
}

// sessionVideoStatus has no expiry gate, like session.report: the render
// may finish after the session ended, and the session owner may still ask.
func (s *MORRPCController) sessionVideoStatus(ctx context.Context, msg m.RPCMessage, sendResponse SendResponse, sourceLog lib.ILogger) error {
	var req m.SessionVideoStatusReq
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return lib.WrapError(ErrUnmarshal, err)
	}
	if err := s.validator.Struct(req); err != nil {
		return lib.WrapError(ErrValidation, err)
	}
	sourceLog = sourceLog.With("session_id", req.SessionID.Hex(), "video_job_id", req.JobID)

	session, err := s.sessionRepo.GetSession(ctx, req.SessionID)
	if err != nil {
		return fmt.Errorf("session cannot be loaded %s", err)
	}
	if !s.servesProvider(ctx, session.ProviderAddr()) {
		return fmt.Errorf("session not bound to this provider")
	}

	user, pubKeyHex, err := s.sessionUserPubKey(session.UserAddr())
	if err != nil {
		return err
	}

	sig := req.Signature
	req.Signature = lib.HexString{}
	if !s.morRpc.VerifySignature(req, sig, pubKeyHex, sourceLog) {
		err := fmt.Errorf("invalid signature")
		sourceLog.Error(err)
		return err
	}
	if !videoRequestFresh(req.Timestamp, time.Now()) {
		return ErrVideoJobStale
	}

	// Parallel polls of one session queue here, so only the first reaches
	// the vendor and the rest read the status it cached.
	if err := s.sessionSema.Acquire(ctx, req.SessionID); err != nil {
		return fmt.Errorf("request cancelled while waiting in queue: %w", err)
	}
	defer s.sessionSema.Release(req.SessionID)

	res, err := s.service.SessionVideoStatus(ctx, msg.ID, user.PubKey, req.SessionID, req.JobID, sourceLog)
	if err != nil {
		sourceLog.Warnf("video status failed: %s", err)
		return err
	}
	return sendResponse(res)
}

func (s *ProxyReceiver) loadVideoTask(sessionID common.Hash, jobID string) (*providerVideoTask, error) {
	raw, err := s.sessionStorage.GetVideoTask(sessionID.Hex(), jobID)
	if err != nil || raw == nil {
		return nil, err
	}
	var task providerVideoTask
	if err := json.Unmarshal(raw, &task); err != nil {
		return nil, fmt.Errorf("corrupt video task record: %w", err)
	}
	return &task, nil
}

// sessionVideoSeconds is the session time the jobs already on a session have
// taken: the sum of their quotes.
func (s *ProxyReceiver) sessionVideoSeconds(sessionID common.Hash) (int64, error) {
	records, err := s.sessionStorage.GetVideoTasks(sessionID.Hex())
	if err != nil {
		return 0, err
	}
	var used int64
	for _, raw := range records {
		var task providerVideoTask
		if err := json.Unmarshal(raw, &task); err != nil {
			return 0, fmt.Errorf("corrupt video task record: %w", err)
		}
		used += task.QuoteSeconds
	}
	return used, nil
}

func (s *ProxyReceiver) videoEngine(modelID string, log lib.ILogger) (*config.ModelConfig, aiengine.VideoJobEngine, error) {
	cfg := s.modelConfigLoader.ModelConfigFromID(modelID)
	engine, err := aiengine.NewVideoJobEngine(cfg, log, s.video.httpClient)
	if err != nil {
		return nil, nil, err
	}
	return cfg, engine, nil
}

// receiptFor builds the receipt for a stored task. It carries no vendor key:
// the consumer asks this provider for the status.
func receiptFor(task *providerVideoTask) *videojob.Receipt {
	return &videojob.Receipt{
		Object:       videojob.ObjectJob,
		Status:       videojob.StatusQueued,
		Vendor:       task.Vendor,
		VendorTaskID: task.VendorTaskID,
		StatusURL:    task.StatusURL,
		QuoteSeconds: task.QuoteSeconds,
		ClientJobID:  task.JobID,
		Resolved:     task.Resolved,
	}
}

func (s *ProxyReceiver) SessionVideoCreate(ctx context.Context, reqID string, userPubKey string, session *sessionrepo.SessionModel, req *m.SessionVideoCreateReq, sourceLog lib.ILogger) (*m.RpcResponse, error) {
	ctx = lib.ContextWithRequestID(ctx, reqID)
	modelID := session.ModelID().Hex()

	existing, err := s.loadVideoTask(session.ID(), req.JobID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		sourceLog.Infof("video job %s already accepted, returning its receipt", req.JobID)
		return s.videoResponse(receiptFor(existing), userPubKey, reqID)
	}

	var create videojob.CreateRequest
	if err := json.Unmarshal([]byte(req.Message), &create); err != nil {
		return nil, lib.WrapError(ErrUnmarshal, err)
	}

	cfg, engine, err := s.videoEngine(modelID, sourceLog)
	if err != nil {
		return nil, err
	}
	spec, err := aiengine.VideoSpecFor(cfg, engine)
	if err != nil {
		return nil, err
	}
	resolved, quote, err := spec.Guide().SessionSeconds(&create)
	if err != nil {
		return nil, err
	}

	used, err := s.sessionVideoSeconds(session.ID())
	if err != nil {
		return nil, err
	}
	if err := s.checkVideoSessionLength(ctx, session, used, quote, sourceLog); err != nil {
		return nil, err
	}

	// The switch is sent whenever the model has one, false included: both
	// vendors default it on.
	input := &aiengine.VideoJobInput{Resolved: resolved, Audio: spec.Audio}
	if create.ImageURL != "" {
		if err := videojob.ValidateRemoteURL(create.ImageURL); err != nil {
			return nil, fmt.Errorf("%w: image_url must be a public https URL", videojob.ErrInvalidRequest)
		}
		input.ImageURL = create.ImageURL
	} else if create.ImageB64 != "" {
		img, err := videojob.DecodeImage(create.ImageB64)
		if err != nil {
			return nil, err
		}
		input.Image = img
	}

	start := time.Now().Unix()
	task, err := engine.CreateVideoJob(ctx, input)
	s.trackVideoResult(session.ModelID(), err)
	if err != nil {
		return nil, err
	}
	sourceLog.Infof("video job %s accepted by %s as task %s, quote %ds", req.JobID, task.Vendor, task.TaskID, quote)
	if task.AspectRatio != "" {
		resolved.AspectRatio = task.AspectRatio
	}

	record := &providerVideoTask{
		SessionID:    session.ID().Hex(),
		JobID:        req.JobID,
		ModelID:      modelID,
		Vendor:       task.Vendor,
		VendorTaskID: task.TaskID,
		StatusURL:    task.StatusURL,
		QuoteSeconds: quote,
		Resolved:     resolved,
		CreatedAt:    time.Now().Unix(),
	}
	if raw, err := json.Marshal(record); err == nil {
		if err := s.sessionStorage.PutVideoTask(session.ID().Hex(), req.JobID, raw); err != nil {
			// The vendor task exists: still hand out the receipt. Its quote is
			// not counted against the session, and a retry of this accept
			// could start a second task.
			sourceLog.Errorf("failed to store video task record: %s", err)
		}
	}
	s.recordActivity(ctx, session, start, sourceLog)

	return s.videoResponse(receiptFor(record), userPubKey, reqID)
}

// videoSessionReader reads a session from the chain; BlockchainService (the
// receiver's BidGetter) implements it.
type videoSessionReader interface {
	GetSession(ctx context.Context, sessionID common.Hash) (*structs.Session, error)
}

func readChainSession(ctx context.Context, reader videoSessionReader, sessionID common.Hash) (*structs.Session, error) {
	var err error
	for attempt := 1; attempt <= videoChainReadAttempts; attempt++ {
		var onchain *structs.Session
		if onchain, err = reader.GetSession(ctx, sessionID); err == nil {
			return onchain, nil
		}
		if attempt < videoChainReadAttempts {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
			}
		}
	}
	return nil, err
}

// checkVideoSessionLength refuses a job its session does not pay for. A
// session's length is the price of the clips made on it: the on-chain length
// (endsAt - openedAt) must cover the quotes of the jobs already on it (used)
// plus this one, and a session closed on chain (refunded) pays for nothing.
// The cached session cannot show a close, so a failed chain read refuses the
// job. Without a chain reader at all the time left is used, with grace for
// the time it took the consumer to open the session and send the job.
func (s *ProxyReceiver) checkVideoSessionLength(ctx context.Context, session *sessionrepo.SessionModel, used, quote int64, log lib.ILogger) error {
	need := used + quote
	if reader, ok := s.service.(videoSessionReader); ok {
		onchain, err := readChainSession(ctx, reader, session.ID())
		if err != nil {
			log.Warnf("cannot read session from chain: %s", err)
			return fmt.Errorf("cannot read the session from chain, try again")
		}
		if onchain.ClosedAt != nil && onchain.ClosedAt.Sign() > 0 {
			return fmt.Errorf("the session is closed")
		}
		if onchain.OpenedAt != nil && onchain.EndsAt != nil {
			length := onchain.EndsAt.Int64() - onchain.OpenedAt.Int64()
			if length+videoLengthToleranceSeconds < need {
				return videoSessionTooShort(quote, used, fmt.Sprintf("the session is %d seconds long", length))
			}
			return nil
		}
		log.Warnf("chain session %s has no timestamps, checking the time left instead", session.ID().Hex())
	}

	remaining := session.EndsAt().Int64() - time.Now().Unix()
	if remaining+videoQuoteGraceSeconds < need {
		return videoSessionTooShort(quote, used, fmt.Sprintf("the session has %d seconds left", remaining))
	}
	return nil
}

func videoSessionTooShort(quote, used int64, session string) error {
	if used == 0 {
		return fmt.Errorf("this job needs a %d-second session, %s", quote, session)
	}
	return fmt.Errorf("this job needs %d seconds of session on top of the %d seconds earlier jobs on it took, %s", quote, used, session)
}

// trackVideoResult feeds a vendor create into the model health tracker. Only
// a refusal of the request itself (400, 413, 422: bad parameters, content
// policy) is the consumer's fault; 401/402/403 are the provider's vendor
// account (bad key, no balance, model not activated), which a consumer cannot
// cause and which make every job on the model fail.
func (s *ProxyReceiver) trackVideoResult(modelID common.Hash, err error) {
	if s.modelHealth == nil {
		return
	}
	var vendorErr *aiengine.VideoVendorError
	if errors.As(err, &vendorErr) {
		switch vendorErr.StatusCode {
		case 400, 413, 422:
		default:
			s.modelHealth.ReportPromptFailure(modelID, vendorErr.StatusCode)
		}
		return
	}
	s.trackPromptResult(modelID, err, 0, false)
}

func (s *ProxyReceiver) SessionVideoStatus(ctx context.Context, reqID string, userPubKey string, sessionID common.Hash, jobID string, sourceLog lib.ILogger) (*m.RpcResponse, error) {
	task, err := s.loadVideoTask(sessionID, jobID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrNoVideoJob
	}
	// Keyed by session and job: a job id is the consumer's choice, so it
	// alone could reach another session's cached status.
	cacheKey := task.SessionID + ":" + task.JobID

	s.video.mu.Lock()
	cached, ok := s.video.statusCache[cacheKey]
	s.video.mu.Unlock()
	if ok && cached.fresh(time.Now()) {
		return s.videoResponse(cached.update, userPubKey, reqID)
	}

	_, engine, err := s.videoEngine(task.ModelID, sourceLog)
	if err != nil {
		return nil, err
	}
	body, err := engine.FetchVideoStatus(ctx, task.VendorTaskID)
	if err != nil {
		return nil, err
	}
	update, err := videojob.ParseStatus(task.Vendor, body)
	if err != nil {
		return nil, err
	}
	if update.VideoURL != "" {
		if err := s.withheldIfClosedEarly(ctx, sessionID, &update); err != nil {
			return nil, err
		}
	}
	// Bound to the job asked about, so an old signed answer cannot stand in
	// for another job's status.
	update.JobID = task.JobID

	now := time.Now()
	s.video.mu.Lock()
	if s.video.statusCache == nil {
		s.video.statusCache = make(map[string]cachedVideoStatus)
	}
	for id, c := range s.video.statusCache {
		if !c.fresh(now) {
			delete(s.video.statusCache, id)
		}
	}
	s.video.statusCache[cacheKey] = cachedVideoStatus{at: now, update: update}
	s.video.mu.Unlock()

	return s.videoResponse(update, userPubKey, reqID)
}

// withheldIfClosedEarly drops the video link when the session was closed
// before its end: the consumer took the refund for the unused time, so a
// clip not delivered by then is not handed over. Clips delivered before the
// close stay delivered, and a close within videoEarlyCloseGraceSeconds of
// endsAt is the session's normal end.
func (s *ProxyReceiver) withheldIfClosedEarly(ctx context.Context, sessionID common.Hash, update *videojob.Update) error {
	reader, ok := s.service.(videoSessionReader)
	if !ok {
		return nil
	}
	onchain, err := readChainSession(ctx, reader, sessionID)
	if err != nil {
		return fmt.Errorf("cannot read the session from chain, try again")
	}
	if onchain.ClosedAt == nil || onchain.ClosedAt.Sign() == 0 || onchain.EndsAt == nil ||
		onchain.ClosedAt.Int64()+videoEarlyCloseGraceSeconds >= onchain.EndsAt.Int64() {
		return nil
	}
	*update = videojob.Update{
		Status: videojob.StatusCancelled,
		Error:  &videojob.JobError{Code: "session_closed_early", Message: "the session was closed before the clip was delivered"},
	}
	return nil
}

func (s *ProxyReceiver) videoResponse(payload interface{}, userPubKey string, reqID string) (*m.RpcResponse, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	encrypted, err := lib.EncryptString(string(raw), lib.RemoveHexPrefix(userPubKey))
	if err != nil {
		return nil, err
	}
	return s.morRpc.SessionVideoResponse(encrypted, s.privateKeyHex, reqID)
}

// withVideoGuides puts the current guide of every video model in models-config
// into the pong. The health checker attaches guides on its sweep, but it may
// be disabled (MODEL_HEALTH_CHECK_DISABLED), not have reached the model yet,
// or hold a guide from before a config reload; without the right guide a
// consumer sizes the session wrong or not at all.
func (s *ProxyReceiver) withVideoGuides(reports []system.ModelHealthReport, log lib.ILogger) []system.ModelHealthReport {
	if s.modelConfigLoader == nil {
		return reports
	}
	ids, cfgs := s.modelConfigLoader.GetAll()

	out := make([]system.ModelHealthReport, len(reports))
	copy(out, reports)
	index := make(map[string]int, len(out))
	for i, r := range out {
		index[strings.ToLower(r.ModelID)] = i
	}

	for i := range cfgs {
		cfg := cfgs[i]
		if !aiengine.IsVideoApiType(cfg.ApiType) {
			continue
		}
		id := ids[i].Hex()
		j, reported := index[strings.ToLower(id)]
		// Built from the current config, not taken from the last sweep: after
		// a models-config reload the provider prices jobs with the new block.
		guide, err := aiengine.VideoGuideFor(&cfg, log)
		if err != nil {
			if reported {
				out[j].Video = nil
			}
			continue
		}
		if reported {
			out[j].Video = guide
			continue
		}
		out = append(out, system.ModelHealthReport{
			ModelID:     id,
			Status:      system.ModelHealthStatusSkipped,
			LastChecked: time.Now().Unix(),
			Video:       guide,
		})
	}
	return out
}
