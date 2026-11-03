package videojob

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
)

// Field names in the guide. Resolve reads the limits back from these entries,
// so the consumer and the provider validate a request against the same
// document.
const (
	FieldPrompt          = "prompt"
	FieldDurationSeconds = "duration_seconds"
	FieldAspectRatio     = "aspect_ratio"
	FieldResolution      = "resolution"
	FieldGenerateAudio   = "generate_audio"
	FieldImageURL        = "image_url"
	FieldImageB64        = "image_b64"
)

// MaxPromptLength bounds the prompt the guide accepts.
const MaxPromptLength = 4000

var (
	ErrInvalidRequest = errors.New("invalid video job request")
	ErrQuoteTooLong   = errors.New("video job quote exceeds the longest session this model sells")
)

// Guide is what a provider advertises for a video model before anyone opens
// a session: the fields a job may set, their limits and defaults, plain
// instructions for a form or an agent, and the weights that turn a job into a
// session length.
type Guide struct {
	Object       string   `json:"object"`
	ModelName    string   `json:"model_name,omitempty"`
	Summary      string   `json:"summary,omitempty"`
	Defaults     Defaults `json:"defaults"`
	Fields       []Field  `json:"fields"`
	Instructions string   `json:"instructions,omitempty"`
	Quote        Quote    `json:"quote"`
}

type Defaults struct {
	DurationSeconds int    `json:"duration_seconds"`
	AspectRatio     string `json:"aspect_ratio"`
	Resolution      string `json:"resolution"`
	GenerateAudio   bool   `json:"generate_audio"`
}

type Field struct {
	Name        string   `json:"name"`
	Required    bool     `json:"required"`
	Type        string   `json:"type,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Minimum     *int     `json:"minimum,omitempty"`
	Maximum     *int     `json:"maximum,omitempty"`
	Description string   `json:"description,omitempty"`
}

// Quote turns a resolved job into a session length:
//
//	work           = duration × resolutionWeight × audioWeight (if audio) × imageWeight (if image)
//	sessionSeconds = clamp(ceil(work × secondsPerWorkUnit), minSession, maxSession)
//
// pricePerSecond × sessionSeconds is the price of the clip.
type Quote struct {
	SecondsPerWorkUnit float64            `json:"seconds_per_work_unit"`
	ResolutionWeights  map[string]float64 `json:"resolution_weights,omitempty"`
	AudioWeight        float64            `json:"audio_weight,omitempty"`
	ImageWeight        float64            `json:"image_weight,omitempty"`
	MinSessionSeconds  int64              `json:"min_session_seconds"`
	MaxSessionSeconds  int64              `json:"max_session_seconds"`
}

// SessionSeconds returns the session length that pays for this job. A job
// whose quote is longer than MaxSessionSeconds is refused with
// ErrQuoteTooLong; a missing weight counts as 1.
func (q Quote) SessionSeconds(r Resolved) (int64, error) {
	work := float64(r.DurationSeconds) * weightOr1(q.ResolutionWeights[r.Resolution])
	if r.GenerateAudio {
		work *= weightOr1(q.AudioWeight)
	}
	if r.HasImage {
		work *= weightOr1(q.ImageWeight)
	}

	secs := int64(math.Ceil(work * q.SecondsPerWorkUnit))
	if secs < q.MinSessionSeconds {
		secs = q.MinSessionSeconds
	}
	if q.MaxSessionSeconds > 0 && secs > q.MaxSessionSeconds {
		return secs, fmt.Errorf("%w: quote %ds, max %ds", ErrQuoteTooLong, secs, q.MaxSessionSeconds)
	}
	return secs, nil
}

func weightOr1(w float64) float64 {
	if w <= 0 {
		return 1
	}
	return w
}

// Validate checks that the quote can price a job.
func (q Quote) Validate() error {
	if q.SecondsPerWorkUnit <= 0 {
		return fmt.Errorf("quote.secondsPerWorkUnit must be positive")
	}
	if q.MinSessionSeconds < 0 || q.MaxSessionSeconds < 0 {
		return fmt.Errorf("quote session bounds must not be negative")
	}
	if q.MaxSessionSeconds > 0 && q.MaxSessionSeconds < q.MinSessionSeconds {
		return fmt.Errorf("quote.maxSessionSeconds is below quote.minSessionSeconds")
	}
	for res, w := range q.ResolutionWeights {
		if w <= 0 {
			return fmt.Errorf("quote.resolutionWeights[%s] must be positive", res)
		}
	}
	if q.AudioWeight < 0 || q.ImageWeight < 0 {
		return fmt.Errorf("quote weights must not be negative")
	}
	return nil
}

// Spec is the provider's description of one video model: what the vendor
// accepts and how the provider prices it. Guide renders it for the wire.
type Spec struct {
	ModelName    string
	MinDuration  int
	MaxDuration  int
	AspectRatios []string
	Resolutions  []string
	// Audio is true when generate_audio is a real switch at the vendor.
	Audio bool
	// Image is true when a first frame is accepted (image_url).
	Image bool
	// ImageInline is true when the first frame may also be sent as bytes
	// (image_b64).
	ImageInline bool
	Defaults    Defaults
	Quote       Quote
}

// Validate checks that the spec describes a model a job can be resolved
// against.
func (s Spec) Validate() error {
	if s.MinDuration <= 0 || s.MaxDuration < s.MinDuration {
		return fmt.Errorf("invalid duration range %d-%d", s.MinDuration, s.MaxDuration)
	}
	if len(s.AspectRatios) == 0 || len(s.Resolutions) == 0 {
		return fmt.Errorf("aspect ratios and resolutions must not be empty")
	}
	if s.Defaults.DurationSeconds < s.MinDuration || s.Defaults.DurationSeconds > s.MaxDuration {
		return fmt.Errorf("default duration %d is outside %d-%d", s.Defaults.DurationSeconds, s.MinDuration, s.MaxDuration)
	}
	if !slices.Contains(s.AspectRatios, s.Defaults.AspectRatio) {
		return fmt.Errorf("default aspect ratio %q is not offered", s.Defaults.AspectRatio)
	}
	if !slices.Contains(s.Resolutions, s.Defaults.Resolution) {
		return fmt.Errorf("default resolution %q is not offered", s.Defaults.Resolution)
	}
	if s.Defaults.GenerateAudio && !s.Audio {
		return fmt.Errorf("default generate_audio is true but audio is not offered")
	}
	if s.ImageInline && !s.Image {
		return fmt.Errorf("inline images need image input")
	}
	return s.Quote.Validate()
}

// Guide renders the spec as the advertised guide.
func (s Spec) Guide() *Guide {
	minDur, maxDur := s.MinDuration, s.MaxDuration
	fields := []Field{
		{Name: FieldPrompt, Required: true, Type: "string", Description: "What the clip shows."},
		{Name: FieldDurationSeconds, Type: "integer", Minimum: &minDur, Maximum: &maxDur, Description: "Clip length in seconds. Changes the quote."},
		{Name: FieldAspectRatio, Type: "string", Enum: slices.Clone(s.AspectRatios)},
		{Name: FieldResolution, Type: "string", Enum: slices.Clone(s.Resolutions), Description: "Changes the quote."},
	}
	if s.Audio {
		fields = append(fields, Field{Name: FieldGenerateAudio, Type: "boolean", Description: "Generate a soundtrack. Changes the quote."})
	}
	if s.Image {
		fields = append(fields, Field{Name: FieldImageURL, Type: "string", Description: "Optional first frame as a public https URL the vendor can fetch."})
	}
	if s.ImageInline {
		fields = append(fields, Field{Name: FieldImageB64, Type: "string", Description: "Optional first frame as base64 or a data URI. JPEG, PNG or WebP."})
	}

	summary := fmt.Sprintf("Text to video. Clip length %d-%d seconds.", s.MinDuration, s.MaxDuration)
	if s.Image {
		summary = fmt.Sprintf("Text to video, or text plus one image as the first frame. Clip length %d-%d seconds.", s.MinDuration, s.MaxDuration)
	}
	instructions := "Ask for the scene in a sentence. Offer length, widescreen or vertical, and sound only if the user cares. Otherwise use defaults. Show the session quote before opening."

	weights := make(map[string]float64, len(s.Quote.ResolutionWeights))
	for k, v := range s.Quote.ResolutionWeights {
		weights[k] = v
	}
	quote := s.Quote
	quote.ResolutionWeights = weights

	return &Guide{
		Object:       ObjectGuide,
		ModelName:    s.ModelName,
		Summary:      summary,
		Defaults:     s.Defaults,
		Fields:       fields,
		Instructions: instructions,
		Quote:        quote,
	}
}

func (g *Guide) field(name string) *Field {
	for i := range g.Fields {
		if g.Fields[i].Name == name {
			return &g.Fields[i]
		}
	}
	return nil
}

// Resolve applies the guide's defaults to req and checks it against the
// guide's limits.
func (g *Guide) Resolve(req *CreateRequest) (Resolved, error) {
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return Resolved{}, fmt.Errorf("%w: prompt is required", ErrInvalidRequest)
	}
	if len(prompt) > MaxPromptLength {
		return Resolved{}, fmt.Errorf("%w: prompt is longer than %d characters", ErrInvalidRequest, MaxPromptLength)
	}

	r := Resolved{
		Prompt:          prompt,
		DurationSeconds: g.Defaults.DurationSeconds,
		AspectRatio:     g.Defaults.AspectRatio,
		Resolution:      g.Defaults.Resolution,
		GenerateAudio:   g.Defaults.GenerateAudio,
		HasImage:        req.HasImage(),
	}

	if req.DurationSeconds != nil {
		r.DurationSeconds = *req.DurationSeconds
	}
	if f := g.field(FieldDurationSeconds); f != nil {
		if (f.Minimum != nil && r.DurationSeconds < *f.Minimum) || (f.Maximum != nil && r.DurationSeconds > *f.Maximum) {
			return Resolved{}, fmt.Errorf("%w: duration_seconds %d is outside %s", ErrInvalidRequest, r.DurationSeconds, rangeString(f))
		}
	}

	if req.AspectRatio != "" {
		r.AspectRatio = req.AspectRatio
	}
	if err := checkEnum(g.field(FieldAspectRatio), r.AspectRatio); err != nil {
		return Resolved{}, err
	}

	if req.Resolution != "" {
		r.Resolution = strings.ToLower(req.Resolution)
	}
	if err := checkEnum(g.field(FieldResolution), r.Resolution); err != nil {
		return Resolved{}, err
	}

	if req.GenerateAudio != nil {
		if *req.GenerateAudio && g.field(FieldGenerateAudio) == nil {
			return Resolved{}, fmt.Errorf("%w: this model does not offer generate_audio", ErrInvalidRequest)
		}
		r.GenerateAudio = *req.GenerateAudio
	}

	if req.ImageURL != "" && g.field(FieldImageURL) == nil {
		return Resolved{}, fmt.Errorf("%w: this model does not accept a first frame", ErrInvalidRequest)
	}
	if req.ImageURL == "" && req.ImageB64 != "" && g.field(FieldImageB64) == nil {
		return Resolved{}, fmt.Errorf("%w: this model does not accept an inline first frame; send image_url", ErrInvalidRequest)
	}

	return r, nil
}

func checkEnum(f *Field, value string) error {
	if f == nil || len(f.Enum) == 0 {
		return nil
	}
	if !slices.Contains(f.Enum, value) {
		return fmt.Errorf("%w: %s %q is not one of %s", ErrInvalidRequest, f.Name, value, strings.Join(f.Enum, ", "))
	}
	return nil
}

func rangeString(f *Field) string {
	lo, hi := "-", "-"
	if f.Minimum != nil {
		lo = fmt.Sprint(*f.Minimum)
	}
	if f.Maximum != nil {
		hi = fmt.Sprint(*f.Maximum)
	}
	return lo + "-" + hi
}

// SessionSeconds resolves req against the guide and returns its quote.
func (g *Guide) SessionSeconds(req *CreateRequest) (Resolved, int64, error) {
	r, err := g.Resolve(req)
	if err != nil {
		return Resolved{}, 0, err
	}
	secs, err := g.Quote.SessionSeconds(r)
	if err != nil {
		return r, secs, err
	}
	return r, secs, nil
}
