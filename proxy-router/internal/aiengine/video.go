package aiengine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	c "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/videojob"
)

// VideoJobEngine is implemented by adapters for hosted text-to-video vendors.
// CreateVideoJob starts a render and returns as soon as the vendor accepted
// it; nobody waits for the clip on this side.
type VideoJobEngine interface {
	// BaseVideoSpec is what the vendor offers for this model, priced with the
	// default quote.
	BaseVideoSpec() videojob.Spec
	CreateVideoJob(ctx context.Context, job *VideoJobInput) (*VideoTask, error)
	// FetchVideoStatus returns the vendor's status body for a task, for
	// videojob.ParseStatus.
	FetchVideoStatus(ctx context.Context, taskID string) ([]byte, error)
	// ProbeVideoVendor asks the vendor something free that only a working
	// account can answer: it proves the vendor answers and accepts the key,
	// without starting a render. It cannot see an inactive model or an empty
	// balance; only a create shows those.
	ProbeVideoVendor(ctx context.Context) error
}

type VideoJobInput struct {
	Resolved videojob.Resolved
	// Audio is true when the served spec offers generate_audio: the adapter
	// then sends the switch either way, since both vendors default it on.
	Audio bool
	// ImageURL is a first frame the vendor fetches itself.
	ImageURL string
	// Image is an inline first frame; set only when ImageURL is empty.
	Image *videojob.Image
}

// VideoTask is the vendor's handle on a render.
type VideoTask struct {
	Vendor string
	TaskID string
	// AspectRatio is the ratio the vendor was asked for when the adapter
	// changed the job's (a first frame keeps its own ratio).
	AspectRatio string
}

// VideoVendorError is a non-2xx answer from the vendor, with what its body
// said.
type VideoVendorError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *VideoVendorError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.StatusCode)
	}
	if e.Code != "" {
		return fmt.Sprintf("video vendor refused the request (%d %s): %s", e.StatusCode, e.Code, msg)
	}
	return fmt.Sprintf("video vendor refused the request (%d): %s", e.StatusCode, msg)
}

var (
	ErrVideoNotSupported = errors.New("this is a video model: submit a video job (POST /v1/video/jobs) instead")
	// ErrVideoVendorUnreachable is what a consumer sees of a transport
	// failure: the wrapped error names the provider's apiUrl, so it stays in
	// the provider's log.
	ErrVideoVendorUnreachable = errors.New("video vendor did not answer")
)

const (
	// videoCreateTimeout stays under the consumer's 60 s wait for the
	// receipt. Neither vendor takes an idempotency key, so a create that
	// times out after the vendor accepted it leaves an orphan task: give the
	// vendor all the time the consumer allows.
	videoCreateTimeout = 50 * time.Second
	videoStatusTimeout = 15 * time.Second
	// vendor bodies are small JSON documents
	maxVideoVendorBody = 1 << 20
)

// IsVideoApiType reports whether apiType is a hosted video vendor adapter.
func IsVideoApiType(apiType string) bool {
	return apiType == API_TYPE_RENDERFUL || apiType == API_TYPE_BYTEPLUS_VIDEO
}

// NewVideoJobEngine builds the video adapter for a models-config entry.
func NewVideoJobEngine(cfg *config.ModelConfig, log lib.ILogger, httpClient *http.Client) (VideoJobEngine, error) {
	if cfg == nil || cfg.ApiType == "" {
		return nil, fmt.Errorf("model is not configured on this provider")
	}
	switch cfg.ApiType {
	case API_TYPE_RENDERFUL:
		return NewRenderfulEngine(cfg.ModelName, cfg.ApiURL, cfg.ApiKey, log, httpClient), nil
	case API_TYPE_BYTEPLUS_VIDEO:
		return NewBytePlusVideoEngine(cfg.ModelName, cfg.ApiURL, cfg.ApiKey, log, httpClient), nil
	}
	return nil, fmt.Errorf("this model does not serve video jobs")
}

// VideoSpecFor is the spec a video model is served and advertised with: the
// vendor's limits for the model id and the default quote.
func VideoSpecFor(cfg *config.ModelConfig, engine VideoJobEngine) (videojob.Spec, error) {
	spec := engine.BaseVideoSpec()
	spec.ModelName = cfg.ModelName
	if err := spec.Validate(); err != nil {
		return videojob.Spec{}, fmt.Errorf("invalid video spec for %s: %w", cfg.ModelName, err)
	}
	return spec, nil
}

// VideoGuideFor renders the advertised guide of a video model, or an error
// when the model is not a video model.
func VideoGuideFor(cfg *config.ModelConfig, log lib.ILogger) (*videojob.Guide, error) {
	engine, err := NewVideoJobEngine(cfg, log, nil)
	if err != nil {
		return nil, err
	}
	spec, err := VideoSpecFor(cfg, engine)
	if err != nil {
		return nil, err
	}
	return spec.Guide(), nil
}

// defaultVideoQuote prices the cheapest clip at the 300-second contract
// minimum (4 s × 1 × 75) and weighs 720p by the vendors' published 480p/720p
// price ratio (~2.25). A provider sets the price itself with the bid's
// pricePerSecond.
func defaultVideoQuote() videojob.Quote {
	return videojob.Quote{
		SecondsPerWorkUnit: 75,
		ResolutionWeights:  map[string]float64{"480p": 1, "720p": 2.25, "1080p": 5},
		AudioWeight:        1,
		ImageWeight:        1,
		MinSessionSeconds:  300,
		MaxSessionSeconds:  86400,
	}
}

// videoHTTPClient keeps the vendor key on the vendor's host: a redirect to
// another host is not followed. An injected client (the factory's, or a
// test's) keeps its transport but gets the same guard unless it set its own.
func videoHTTPClient(httpClient *http.Client) *http.Client {
	if httpClient == nil {
		return &http.Client{CheckRedirect: refuseCrossHostRedirect}
	}
	if httpClient.CheckRedirect != nil {
		return httpClient
	}
	guarded := *httpClient
	guarded.CheckRedirect = refuseCrossHostRedirect
	return &guarded
}

func refuseCrossHostRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	prev := via[len(via)-1].URL
	if req.URL.Host != prev.Host || (prev.Scheme == "https" && req.URL.Scheme != "https") {
		return http.ErrUseLastResponse
	}
	return nil
}

func setVideoAuth(req *http.Request, apiKey string) {
	req.Header.Set(c.HEADER_ACCEPT, c.CONTENT_TYPE_JSON)
	if apiKey != "" {
		req.Header.Set(c.HEADER_AUTHORIZATION, c.BEARER+" "+apiKey)
	}
}

func readVideoBody(res *http.Response) ([]byte, error) {
	return io.ReadAll(io.LimitReader(res.Body, maxVideoVendorBody))
}

// vendorUnreachable logs a transport failure with its detail and returns the
// error the consumer may see. The cause is not wrapped, since its text names
// the vendor's host; a timeout is kept as a fact, since the health report
// tells it apart.
func vendorUnreachable(log lib.ILogger, what string, err error) error {
	log.Warnf("%s: %s", what, err)
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return fmt.Errorf("%s: %w: %w", what, ErrVideoVendorUnreachable, context.DeadlineExceeded)
	}
	return fmt.Errorf("%s: %w", what, ErrVideoVendorUnreachable)
}

// errUnexpectedProbeAnswer marks a 2xx probe answer in a shape the vendor
// does not use: the apiUrl reaches something else.
const errUnexpectedProbeAnswer = "unexpected answer to the health probe"

// videoOnly answers the streaming methods of a video adapter: a chat request
// to a video model gets a clear error instead of a vendor call.
type videoOnly struct{}

func (videoOnly) Prompt(ctx context.Context, prompt *gcs.OpenAICompletionRequestExtra, cb gcs.CompletionCallback) error {
	return ErrVideoNotSupported
}

func (videoOnly) AudioTranscription(ctx context.Context, prompt *gcs.AudioTranscriptionRequest, cb gcs.CompletionCallback) error {
	return ErrVideoNotSupported
}

func (videoOnly) AudioSpeech(ctx context.Context, prompt *gcs.AudioSpeechRequest, cb gcs.CompletionCallback) error {
	return ErrVideoNotSupported
}

func (videoOnly) Embeddings(ctx context.Context, prompt *gcs.EmbeddingsRequest, cb gcs.CompletionCallback) error {
	return ErrVideoNotSupported
}

func (videoOnly) Decisions(ctx context.Context, prompt *gcs.DecisionsRequest, cb gcs.CompletionCallback) error {
	return ErrVideoNotSupported
}
