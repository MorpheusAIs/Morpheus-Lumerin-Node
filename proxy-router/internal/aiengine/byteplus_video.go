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

// API_TYPE_BYTEPLUS_VIDEO is ByteDance Seedance on BytePlus ModelArk (the
// international Ark API). apiUrl is the API base, modelName the Ark model id
// (e.g. dreamina-seedance-2-5-260628) or a custom inference endpoint id.
const API_TYPE_BYTEPLUS_VIDEO = "byteplus-video"
const BYTEPLUS_VIDEO_DEFAULT_BASE_URL = "https://ark.ap-southeast.bytepluses.com/api/v3"

type BytePlusVideo struct {
	videoOnly
	modelName string
	baseURL   string
	apiKey    string
	client    *http.Client
	log       lib.ILogger
}

func NewBytePlusVideoEngine(modelName, baseURL, apiKey string, log lib.ILogger, httpClient *http.Client) *BytePlusVideo {
	if baseURL == "" {
		baseURL = BYTEPLUS_VIDEO_DEFAULT_BASE_URL
	}
	return &BytePlusVideo{
		modelName: modelName,
		baseURL:   strings.TrimSuffix(baseURL, "/"),
		apiKey:    apiKey,
		client:    videoHTTPClient(httpClient),
		log:       log,
	}
}

func (b *BytePlusVideo) ApiType() string {
	return API_TYPE_BYTEPLUS_VIDEO
}

// BaseVideoSpec follows the Ark create-task reference per model family. A
// custom endpoint id or an unknown model gets the narrow common subset.
func (b *BytePlusVideo) BaseVideoSpec() videojob.Spec {
	allRatios := []string{"16:9", "4:3", "1:1", "3:4", "9:16", "21:9", "adaptive"}
	spec := videojob.Spec{
		MinDuration:  4,
		MaxDuration:  12,
		AspectRatios: []string{"16:9", "4:3", "1:1", "3:4", "9:16", "21:9"},
		Resolutions:  []string{"480p", "720p"},
		Image:        true,
		ImageInline:  true,
		Defaults:     videojob.Defaults{DurationSeconds: 5, AspectRatio: "16:9", Resolution: "720p"},
		Quote:        defaultVideoQuote(),
	}

	name := b.modelName
	switch {
	case strings.HasPrefix(name, "dreamina-seedance-2-5"):
		spec.MaxDuration = 30
		spec.AspectRatios = allRatios
		spec.Resolutions = []string{"480p", "720p", "1080p"}
		spec.Audio = true
	case strings.HasPrefix(name, "dreamina-seedance-2-0-fast"), strings.HasPrefix(name, "dreamina-seedance-2-0-mini"):
		spec.MaxDuration = 15
		spec.AspectRatios = allRatios
		spec.Audio = true
	case strings.HasPrefix(name, "dreamina-seedance-2-0"):
		spec.MaxDuration = 15
		spec.AspectRatios = allRatios
		spec.Resolutions = []string{"480p", "720p", "1080p"}
		spec.Audio = true
	case strings.HasPrefix(name, "seedance-1-5-pro"):
		spec.AspectRatios = allRatios
		spec.Resolutions = []string{"480p", "720p", "1080p"}
		spec.Audio = true
	case strings.HasPrefix(name, "seedance-1-0-pro"):
		spec.MinDuration = 2
		spec.Resolutions = []string{"480p", "720p", "1080p"}
	}
	return spec
}

type bytePlusContent struct {
	Type     string            `json:"type"`
	Text     string            `json:"text,omitempty"`
	ImageURL *bytePlusImageURL `json:"image_url,omitempty"`
	Role     string            `json:"role,omitempty"`
}

type bytePlusImageURL struct {
	URL string `json:"url"`
}

type bytePlusCreateReq struct {
	Model         string            `json:"model"`
	Content       []bytePlusContent `json:"content"`
	Ratio         string            `json:"ratio"`
	Duration      int               `json:"duration"`
	Resolution    string            `json:"resolution"`
	GenerateAudio *bool             `json:"generate_audio,omitempty"`
	Watermark     bool              `json:"watermark"`
}

type bytePlusError struct {
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (b *BytePlusVideo) CreateVideoJob(ctx context.Context, job *VideoJobInput) (*VideoTask, error) {
	ctx, cancel := context.WithTimeout(ctx, videoCreateTimeout)
	defer cancel()

	r := job.Resolved
	body := bytePlusCreateReq{
		Model:      b.modelName,
		Content:    []bytePlusContent{{Type: "text", Text: r.Prompt}},
		Ratio:      r.AspectRatio,
		Duration:   r.DurationSeconds,
		Resolution: r.Resolution,
	}
	if job.Audio {
		audio := r.GenerateAudio
		body.GenerateAudio = &audio
	}

	frame := job.ImageURL
	if frame == "" && job.Image != nil {
		frame = job.Image.DataURI()
	}
	if frame != "" {
		body.Content = append(body.Content, bytePlusContent{
			Type:     "image_url",
			ImageURL: &bytePlusImageURL{URL: frame},
			Role:     "first_frame",
		})
		// The frame's own ratio: Seedance 2.5 accepts nothing else for a
		// first-frame job (any other ratio fails after create), and every
		// other model takes adaptive for image-to-video too.
		body.Ratio = "adaptive"
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.baseURL+"/contents/generations/tasks", bytes.NewReader(payload))
	if err != nil {
		return nil, vendorUnreachable(b.log, "build the vendor request", err)
	}
	setVideoAuth(req, b.apiKey)
	req.Header.Set(c.HEADER_CONTENT_TYPE, c.CONTENT_TYPE_JSON)

	res, err := b.client.Do(req)
	if err != nil {
		return nil, vendorUnreachable(b.log, "byteplus create task", err)
	}
	defer res.Body.Close()

	raw, err := readVideoBody(res)
	if err != nil {
		return nil, vendorUnreachable(b.log, "byteplus create task", err)
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, bytePlusVendorError(res.StatusCode, raw)
	}

	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &created); err != nil || created.ID == "" {
		return nil, fmt.Errorf("byteplus create task: no task id in response")
	}

	task := &VideoTask{
		Vendor: videojob.VendorBytePlus,
		TaskID: created.ID,
	}
	if body.Ratio != r.AspectRatio {
		task.AspectRatio = body.Ratio
	}
	return task, nil
}

func (b *BytePlusVideo) statusURL(taskID string) string {
	return b.baseURL + "/contents/generations/tasks/" + url.PathEscape(taskID)
}

func (b *BytePlusVideo) FetchVideoStatus(ctx context.Context, taskID string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, videoStatusTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.statusURL(taskID), nil)
	if err != nil {
		return nil, vendorUnreachable(b.log, "build the vendor request", err)
	}
	setVideoAuth(req, b.apiKey)

	res, err := b.client.Do(req)
	if err != nil {
		return nil, vendorUnreachable(b.log, "byteplus task status", err)
	}
	defer res.Body.Close()

	raw, err := readVideoBody(res)
	if err != nil {
		return nil, vendorUnreachable(b.log, "byteplus task status", err)
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, bytePlusVendorError(res.StatusCode, raw)
	}
	return raw, nil
}

func bytePlusVendorError(status int, raw []byte) *VideoVendorError {
	e := &VideoVendorError{StatusCode: status}
	var parsed bytePlusError
	if json.Unmarshal(raw, &parsed) == nil && parsed.Error != nil {
		e.Code = parsed.Error.Code
		e.Message = parsed.Error.Message
	}
	return e
}

var _ AIEngineStream = &BytePlusVideo{}
var _ VideoJobEngine = &BytePlusVideo{}
