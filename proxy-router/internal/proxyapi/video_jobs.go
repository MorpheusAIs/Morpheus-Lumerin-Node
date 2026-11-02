package proxyapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/videojob"
	"github.com/ethereum/go-ethereum/common"
)

const (
	DefaultVideoJobPollInterval = 10 * time.Second
	DefaultVideoJobMaxWatch     = 72 * time.Hour

	videoPollMaxBackoff = 2 * time.Minute
	videoCreateAttempts = 2
	// videoMaxRefusals ends a job whose provider keeps refusing to report
	// it for a reason a retry cannot fix (record gone, key no longer bound).
	videoMaxRefusals = 3
)

var (
	ErrVideoJobNotFound  = errors.New("video job not found")
	ErrVideoJobsDisabled = errors.New("video jobs are not enabled on this node")
)

// VideoSessionService is the chain side a consumer video job needs.
// Implemented by blockchainapi.BlockchainService.
type VideoSessionService interface {
	VideoGuides(ctx context.Context, modelID common.Hash) ([]structs.VideoProviderGuide, error)
	VideoSessionBounds(ctx context.Context) (minSeconds, maxSeconds int64, err error)
	GetSession(ctx context.Context, sessionID common.Hash) (*structs.Session, error)
}

// videoJobSender is the provider side: ProxyServiceSender.
type videoJobSender interface {
	SendVideoJobCreate(ctx context.Context, sessionID common.Hash, jobID string, job *videojob.CreateRequest) (*videojob.Receipt, error)
	SendVideoJobStatus(ctx context.Context, sessionID common.Hash, jobID string) (*videojob.Update, error)
}

// CreateVideoJobParams is one submit. SessionID is a session the caller
// opened for its jobs: the quotes of all jobs on it must fit in its length.
type CreateVideoJobParams struct {
	Request   videojob.CreateRequest
	SessionID common.Hash
}

// videoJobState is a job as this node tracks it.
type videoJobState struct {
	job videojob.Job
}

// VideoJobService runs the consumer half of video jobs: it sends each job on
// the session the caller opened for it, then asks the provider for the job's
// status until the job ends. It never opens or closes a session: that is the
// caller's.
type VideoJobService struct {
	ctx      context.Context
	sender   videoJobSender
	sessions VideoSessionService
	store    *storages.SessionStorage
	log      lib.ILogger

	// checkURL vets a relayed video URL before it is recorded: callers open
	// it directly, and it came from the vendor through the provider.
	checkURL func(string) error

	pollInterval time.Duration
	maxWatch     time.Duration

	mu   sync.Mutex
	jobs map[string]*videoJobState
	wg   sync.WaitGroup
}

// NewVideoJobService builds the service. ctx bounds every watcher it starts:
// pass the application context.
func NewVideoJobService(ctx context.Context, sender videoJobSender, sessions VideoSessionService, store *storages.SessionStorage, pollInterval, maxWatch time.Duration, log lib.ILogger) *VideoJobService {
	if pollInterval <= 0 {
		pollInterval = DefaultVideoJobPollInterval
	}
	if maxWatch <= 0 {
		maxWatch = DefaultVideoJobMaxWatch
	}
	return &VideoJobService{
		ctx:          ctx,
		sender:       sender,
		sessions:     sessions,
		store:        store,
		log:          log.Named("VIDEO_JOBS"),
		checkURL:     videojob.ValidateRemoteURL,
		pollInterval: pollInterval,
		maxWatch:     maxWatch,
		jobs:         make(map[string]*videoJobState),
	}
}

// Resume reloads stored jobs and restarts a watcher for each one that has
// not ended. A restarted node has no vendor key, so it asks the provider.
func (s *VideoJobService) Resume() {
	records, err := s.store.GetVideoJobs()
	if err != nil {
		s.log.Errorf("failed to load video jobs: %s", err)
		return
	}
	for _, raw := range records {
		var job videojob.Job
		if err := json.Unmarshal(raw, &job); err != nil {
			s.log.Warnf("skipping corrupt video job record: %s", err)
			continue
		}
		state := &videoJobState{job: job}
		s.mu.Lock()
		if _, ok := s.jobs[job.ID]; ok {
			s.mu.Unlock()
			continue
		}
		s.jobs[job.ID] = state
		s.mu.Unlock()
		if !videojob.IsTerminal(job.Status) {
			s.startWatcher(state)
		}
	}
}

// Wait blocks until every watcher has returned (after ctx is cancelled).
func (s *VideoJobService) Wait() {
	s.wg.Wait()
}

func newVideoJobID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "vjob_" + hex.EncodeToString(b), nil
}

// Create sends the job on the caller's session and starts watching it. A
// failed submit leaves the session open: the caller can retry on it or close
// it.
func (s *VideoJobService) Create(ctx context.Context, p CreateVideoJobParams) (*videojob.Job, error) {
	if strings.TrimSpace(p.Request.Prompt) == "" {
		return nil, fmt.Errorf("%w: prompt is required", videojob.ErrInvalidRequest)
	}
	jobID, err := newVideoJobID()
	if err != nil {
		return nil, err
	}
	log := s.log.With("video_job_id", jobID)

	sessionID := p.SessionID
	if sessionID == (common.Hash{}) {
		return nil, fmt.Errorf("%w: session_id is required: open a session for the job's quote first", videojob.ErrInvalidRequest)
	}
	var modelID common.Hash
	if session, err := s.sessions.GetSession(ctx, sessionID); err == nil {
		modelID = common.HexToHash(session.ModelAgentId)
	}

	var receipt *videojob.Receipt
	for attempt := 1; ; attempt++ {
		receipt, err = s.sender.SendVideoJobCreate(ctx, sessionID, jobID, &p.Request)
		if err == nil || errors.Is(err, ErrProviderRefused) || attempt >= videoCreateAttempts {
			break
		}
		// The same job id: a provider that already accepted it answers with
		// the same receipt instead of starting a second render.
		log.Warnf("video job create attempt %d failed, retrying: %s", attempt, err)
	}
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	state := &videoJobState{
		job: videojob.Job{
			Object:       videojob.ObjectJob,
			ID:           jobID,
			Model:        modelID.Hex(),
			Status:       videojob.StatusQueued,
			SessionID:    sessionID.Hex(),
			QuoteSeconds: receipt.QuoteSeconds,
			Vendor:       receipt.Vendor,
			VendorTaskID: receipt.VendorTaskID,
			StatusURL:    receipt.StatusURL,
			Resolved:     receipt.Resolved,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	s.mu.Lock()
	s.jobs[jobID] = state
	s.persistLocked(state)
	job := state.job
	s.mu.Unlock()

	log.Infof("video job accepted: vendor %s task %s", receipt.Vendor, receipt.VendorTaskID)
	s.startWatcher(state)
	return &job, nil
}

// Get returns a job by id.
func (s *VideoJobService) Get(id string) (*videojob.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.jobs[id]
	if !ok {
		return nil, ErrVideoJobNotFound
	}
	job := state.job
	return &job, nil
}

// List returns every job this node knows, newest first.
func (s *VideoJobService) List() []videojob.Job {
	s.mu.Lock()
	jobs := make([]videojob.Job, 0, len(s.jobs))
	for _, state := range s.jobs {
		jobs = append(jobs, state.job)
	}
	s.mu.Unlock()
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].CreatedAt.After(jobs[j].CreatedAt) })
	return jobs
}

// VideoQuote is the session one provider would sell for a job.
type VideoQuote struct {
	Provider          common.Address     `json:"provider"`
	BidID             common.Hash        `json:"bid_id"`
	PricePerSecondWei string             `json:"price_per_second_wei"`
	SessionSeconds    int64              `json:"session_seconds,omitempty"`
	SessionCostWei    string             `json:"session_cost_wei,omitempty"`
	Resolved          *videojob.Resolved `json:"resolved,omitempty"`
	Error             string             `json:"error,omitempty"`
}

// Quote prices a job at every provider of the model, before anything opens,
// with the session raised to the contract minimum as the open would.
func (s *VideoJobService) Quote(ctx context.Context, modelID common.Hash, req *videojob.CreateRequest) ([]VideoQuote, error) {
	guides, err := s.sessions.VideoGuides(ctx, modelID)
	if err != nil {
		return nil, err
	}
	minSeconds, maxSeconds, err := s.sessions.VideoSessionBounds(ctx)
	if err != nil {
		return nil, err
	}
	quotes := make([]VideoQuote, 0, len(guides))
	for _, g := range guides {
		q := VideoQuote{Provider: g.Provider, BidID: g.BidID, PricePerSecondWei: g.PricePerSecondWei, Error: g.Error}
		if g.Guide != nil {
			resolved, secs, err := g.Guide.SessionSeconds(req)
			secs = max(secs, minSeconds)
			if err == nil && maxSeconds > 0 && secs > maxSeconds {
				err = fmt.Errorf("%w: quote %ds, contract max %ds", videojob.ErrQuoteTooLong, secs, maxSeconds)
			}
			if err != nil {
				q.Error = err.Error()
			} else {
				q.SessionSeconds = secs
				q.Resolved = &resolved
				if price, ok := new(big.Int).SetString(g.PricePerSecondWei, 10); ok {
					q.SessionCostWei = price.Mul(price, new(big.Int).SetInt64(secs)).String()
				}
			}
		}
		quotes = append(quotes, q)
	}
	return quotes, nil
}

func (s *VideoJobService) startWatcher(state *videoJobState) {
	s.mu.Lock()
	jobID := state.job.ID
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.watch(s.ctx, jobID)
	}()
}

func (s *VideoJobService) watch(ctx context.Context, jobID string) {
	log := s.log.With("video_job_id", jobID)
	interval := s.pollInterval
	timer := time.NewTimer(interval)
	defer timer.Stop()
	refusals := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		s.mu.Lock()
		state, ok := s.jobs[jobID]
		if !ok || videojob.IsTerminal(state.job.Status) {
			s.mu.Unlock()
			return
		}
		job := state.job
		s.mu.Unlock()

		if time.Since(job.CreatedAt) > s.maxWatch {
			log.Warnf("no final status after %s, giving up", s.maxWatch)
			s.apply(jobID, videojob.Update{
				Status: videojob.StatusFailed,
				Error:  &videojob.JobError{Code: "watch_timeout", Message: fmt.Sprintf("no final status from the vendor after %s", s.maxWatch)},
			}, log)
			return
		}

		update, err := s.sender.SendVideoJobStatus(ctx, common.HexToHash(job.SessionID), job.ID)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if permanentStatusRefusal(err) {
				refusals++
				if refusals >= videoMaxRefusals {
					log.Warnf("the provider keeps refusing the status: %s", err)
					s.apply(jobID, videojob.Update{
						Status: videojob.StatusFailed,
						Error:  &videojob.JobError{Code: "provider_refused", Message: err.Error()},
					}, log)
					return
				}
			} else {
				refusals = 0
			}
			interval = min(interval*2, videoPollMaxBackoff)
			log.Warnf("video status poll failed, next try in %s: %s", interval, err)
			timer.Reset(interval)
			continue
		}
		interval = s.pollInterval
		refusals = 0

		if s.apply(jobID, *update, log) {
			return
		}
		timer.Reset(interval)
	}
}

// apply records an update and reports whether the job has ended.
func (s *VideoJobService) apply(jobID string, update videojob.Update, log lib.ILogger) bool {
	// A link that is not a public https URL is never handed to a caller: the
	// job fails.
	if update.VideoURL != "" {
		if err := s.checkURL(update.VideoURL); err != nil {
			update = videojob.Update{
				Status: videojob.StatusFailed,
				Error:  &videojob.JobError{Code: "invalid_video_url", Message: err.Error()},
			}
		}
	}
	s.mu.Lock()
	state, ok := s.jobs[jobID]
	if !ok {
		s.mu.Unlock()
		return true
	}
	changed := state.job.Status != update.Status || state.job.VideoURL != update.VideoURL
	state.job.Status = update.Status
	if update.VideoURL != "" {
		state.job.VideoURL = update.VideoURL
		state.job.ExpiresAt = update.ExpiresAt
	}
	if update.Error != nil {
		state.job.Error = update.Error
	}
	terminal := videojob.IsTerminal(update.Status)
	if changed || terminal {
		state.job.UpdatedAt = time.Now().UTC()
		s.persistLocked(state)
	}
	s.mu.Unlock()

	if changed {
		log.Infof("video job is %s", update.Status)
	}
	return terminal
}

// permanentStatusRefusal is a provider answer to a status request that no
// retry changes: the provider has no record of the task. A signature or
// provider-binding refusal is not: the wallet may be switched back, and a
// contract provider refuses while its chain reads fail.
func permanentStatusRefusal(err error) bool {
	return errors.Is(err, ErrProviderRefused) && strings.Contains(err.Error(), ErrNoVideoJob.Error())
}

// persistLocked writes the job record. Callers hold s.mu.
func (s *VideoJobService) persistLocked(state *videoJobState) {
	raw, err := json.Marshal(state.job)
	if err != nil {
		s.log.Errorf("failed to encode video job %s: %s", state.job.ID, err)
		return
	}
	if err := s.store.PutVideoJob(state.job.ID, raw); err != nil {
		s.log.Errorf("failed to store video job %s: %s", state.job.ID, err)
	}
}
