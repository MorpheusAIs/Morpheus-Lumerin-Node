package aiengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	c "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/videojob"
)

// API_TYPE_RENDERFUL is the Renderful generations API, an aggregator that
// fronts Seedance, Kling, Wan, Veo and others. apiUrl is the API base,
// modelName the text-to-video model id (e.g. seedance-2.5-t2v).
const API_TYPE_RENDERFUL = "renderful"
const RENDERFUL_DEFAULT_BASE_URL = "https://api.renderful.ai/api/v1"

type Renderful struct {
	videoOnly
	modelName  string
	imageModel string
	baseURL    string
	apiKey     string
	client     *http.Client
	log        lib.ILogger
}

// NewRenderfulEngine builds the adapter. Renderful lists image-to-video as a
// separate model: a "-t2v" modelName maps to its "-i2v" sibling for jobs with
// a first frame, and any other modelName takes no first frame.
func NewRenderfulEngine(modelName, baseURL, apiKey string, log lib.ILogger, httpClient *http.Client) *Renderful {
	if baseURL == "" {
		baseURL = RENDERFUL_DEFAULT_BASE_URL
	}
	imageModel := ""
	if strings.HasSuffix(modelName, "-t2v") {
		imageModel = strings.TrimSuffix(modelName, "-t2v") + "-i2v"
	}
	return &Renderful{
		modelName:  modelName,
		imageModel: imageModel,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		apiKey:     apiKey,
		client:     videoHTTPClient(httpClient),
		log:        log,
	}
}

func (r *Renderful) ApiType() string {
	return API_TYPE_RENDERFUL
}

func (r *Renderful) isSeedance() bool {
	return strings.HasPrefix(r.modelName, "seedance-")
}

// hasAudioSwitch: Seedance 1.5 and later take generate_audio on Renderful;
// Seedance 1.0 makes no sound, and other models name the switch differently
// (enable_audio, sound, ...).
func (r *Renderful) hasAudioSwitch() bool {
	return r.isSeedance() && !strings.HasPrefix(r.modelName, "seedance-1.0")
}

// BaseVideoSpec follows Renderful's catalog for the Seedance family; any
// other catalog model gets the Seedance 2.5 limits. Inline first frames are
// off: image_url must be a URL the vendor can fetch.
func (r *Renderful) BaseVideoSpec() videojob.Spec {
	spec := videojob.Spec{
		MinDuration:  4,
		MaxDuration:  30,
		AspectRatios: []string{"16:9", "9:16", "1:1", "4:3", "3:4", "21:9", "adaptive"},
		Resolutions:  []string{"480p", "720p"},
		Audio:        r.hasAudioSwitch(),
		Image:        r.imageModel != "",
		Defaults:     videojob.Defaults{DurationSeconds: 5, AspectRatio: "16:9", Resolution: "720p"},
		Quote:        defaultVideoQuote(),
	}
	switch {
	case strings.HasPrefix(r.modelName, "seedance-2.0"):
		spec.MaxDuration = 15
	case strings.HasPrefix(r.modelName, "seedance-1.5"):
		spec.MaxDuration = 12
	case strings.HasPrefix(r.modelName, "seedance-1.0"):
		spec.MinDuration = 2
		spec.MaxDuration = 12
		spec.Resolutions = []string{"720p", "1080p"}
	}
	return spec
}

type renderfulCreateReq struct {
	Type          string `json:"type"`
	Model         string `json:"model"`
	Prompt        string `json:"prompt"`
	ImageURL      string `json:"image_url,omitempty"`
	AspectRatio   string `json:"aspect_ratio"`
	Resolution    string `json:"resolution"`
	Duration      int    `json:"duration"`
	GenerateAudio *bool  `json:"generate_audio,omitempty"`
}

type renderfulError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (r *Renderful) CreateVideoJob(ctx context.Context, job *VideoJobInput) (*VideoTask, error) {
	if job.ImageURL == "" && job.Image != nil {
		return nil, &VideoVendorError{StatusCode: http.StatusBadRequest, Message: "renderful takes the first frame as image_url only"}
	}

	ctx, cancel := context.WithTimeout(ctx, videoCreateTimeout)
	defer cancel()

	res := job.Resolved
	body := renderfulCreateReq{
		Type:        "text-to-video",
		Model:       r.modelName,
		Prompt:      res.Prompt,
		AspectRatio: res.AspectRatio,
		Resolution:  res.Resolution,
		Duration:    res.DurationSeconds,
	}
	if job.ImageURL != "" {
		if r.imageModel == "" {
			return nil, &VideoVendorError{StatusCode: http.StatusBadRequest, Message: "this model takes no first frame"}
		}
		body.Type = "image-to-video"
		body.Model = r.imageModel
		body.ImageURL = job.ImageURL
		// Seedance keeps the frame's own ratio for a first-frame job; 2.5
		// fails such a task after accepting it if any other ratio is sent.
		if r.isSeedance() {
			body.AspectRatio = "adaptive"
		}
	}
	if job.Audio {
		audio := res.GenerateAudio
		body.GenerateAudio = &audio
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.baseURL+"/generations", bytes.NewReader(payload))
	if err != nil {
		return nil, vendorUnreachable(r.log, "build the vendor request", err)
	}
	setVideoAuth(req, r.apiKey)
	req.Header.Set(c.HEADER_CONTENT_TYPE, c.CONTENT_TYPE_JSON)

	httpRes, err := r.client.Do(req)
	if err != nil {
		return nil, vendorUnreachable(r.log, "renderful create generation", err)
	}
	defer httpRes.Body.Close()

	raw, err := readVideoBody(httpRes)
	if err != nil {
		return nil, vendorUnreachable(r.log, "renderful create generation", err)
	}
	if httpRes.StatusCode < 200 || httpRes.StatusCode > 299 {
		return nil, renderfulVendorError(httpRes.StatusCode, raw)
	}

	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &created); err != nil || created.ID == "" {
		return nil, fmt.Errorf("renderful create generation: no generation id in response")
	}

	task := &VideoTask{
		Vendor: videojob.VendorRenderful,
		TaskID: created.ID,
	}
	if body.AspectRatio != res.AspectRatio {
		task.AspectRatio = body.AspectRatio
	}
	return task, nil
}

// statusURL is built from the id rather than read from poll_url, which is a
// host-relative path that already carries the /api/v1 prefix.
func (r *Renderful) statusURL(id string) string {
	return r.baseURL + "/generations/" + url.PathEscape(id)
}

func (r *Renderful) FetchVideoStatus(ctx context.Context, taskID string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, videoStatusTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.statusURL(taskID), nil)
	if err != nil {
		return nil, vendorUnreachable(r.log, "build the vendor request", err)
	}
	setVideoAuth(req, r.apiKey)

	res, err := r.client.Do(req)
	if err != nil {
		return nil, vendorUnreachable(r.log, "renderful generation status", err)
	}
	defer res.Body.Close()

	raw, err := readVideoBody(res)
	if err != nil {
		return nil, vendorUnreachable(r.log, "renderful generation status", err)
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, renderfulVendorError(res.StatusCode, raw)
	}
	return raw, nil
}

func renderfulVendorError(status int, raw []byte) *VideoVendorError {
	e := &VideoVendorError{StatusCode: status}
	var parsed renderfulError
	if json.Unmarshal(raw, &parsed) == nil {
		e.Message = parsed.Message
		if e.Message == "" {
			e.Message = parsed.Error
		} else if parsed.Error != "" && parsed.Error != parsed.Message {
			e.Code = parsed.Error
		}
	}
	return e
}

var _ AIEngineStream = &Renderful{}
var _ VideoJobEngine = &Renderful{}
