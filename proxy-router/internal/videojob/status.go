package videojob

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// BytePlusURLLifetime is how long a BytePlus ModelArk video URL stays valid
// after the task succeeded (the vendor also caps Seedance 2.5 links at 100
// downloads).
const BytePlusURLLifetime = 24 * time.Hour

// Update is one vendor status response mapped onto the job vocabulary. It is
// also what a provider returns for session.video.status.
type Update struct {
	// JobID is set by a provider relaying the status: the consumer checks it
	// is the job it asked about.
	JobID     string     `json:"job_id,omitempty"`
	Status    string     `json:"status"`
	VideoURL  string     `json:"video_url,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Error     *JobError  `json:"error,omitempty"`
	// Receipt is the job's receipt, repeated by a provider relaying the
	// status, for a consumer whose copy of it was lost.
	Receipt *Receipt `json:"receipt,omitempty"`
}

// ParseStatus reads a vendor status body. vendor is the receipt's Vendor.
func ParseStatus(vendor string, body []byte) (Update, error) {
	switch vendor {
	case VendorRenderful:
		return parseRenderfulStatus(body)
	case VendorBytePlus:
		return parseBytePlusStatus(body)
	}
	return Update{}, fmt.Errorf("unknown video vendor %q", vendor)
}

// Renderful GET /generations/{id}:
// {"id","status":"queued|processing|completed|failed","output","outputs":[...],"error",...}
type renderfulStatus struct {
	Status  string   `json:"status"`
	Output  string   `json:"output"`
	Outputs []string `json:"outputs"`
	Error   string   `json:"error"`
}

func parseRenderfulStatus(body []byte) (Update, error) {
	var s renderfulStatus
	if err := json.Unmarshal(body, &s); err != nil {
		return Update{}, fmt.Errorf("renderful status: %w", err)
	}

	switch strings.ToLower(s.Status) {
	case "queued", "pending":
		return Update{Status: StatusQueued}, nil
	case "processing", "running", "in_progress":
		return Update{Status: StatusRunning}, nil
	case "completed", "succeeded", "success":
		url := s.Output
		for _, o := range s.Outputs {
			if o != "" {
				url = o
				break
			}
		}
		// Renderful's own client treats "completed" with no outputs yet as
		// still running and polls again.
		if url == "" {
			return Update{Status: StatusRunning}, nil
		}
		// Renderful does not document how long output links live.
		return Update{Status: StatusSucceeded, VideoURL: url}, nil
	case "failed", "error":
		msg := s.Error
		if msg == "" {
			msg = "the vendor reported the render as failed"
		}
		return Update{Status: StatusFailed, Error: &JobError{Message: msg}}, nil
	case "canceled", "cancelled":
		return Update{Status: StatusCancelled, Error: &JobError{Message: "the vendor cancelled the render"}}, nil
	case "":
		return Update{}, fmt.Errorf("renderful status: empty status")
	}
	// An unknown word is not terminal: keep polling until the watch ends.
	return Update{Status: StatusRunning}, nil
}

// BytePlus ModelArk GET /contents/generations/tasks/{id}:
// {"id","model","status":"queued|running|succeeded|failed|expired|cancelled",
// "content":{"video_url","last_frame_url"},"error":{"code","message"},
// "created_at","updated_at",...} with unix-second timestamps.
type bytePlusStatus struct {
	Status  string `json:"status"`
	Content *struct {
		VideoURL string `json:"video_url"`
	} `json:"content"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	UpdatedAt int64 `json:"updated_at"`
}

func parseBytePlusStatus(body []byte) (Update, error) {
	var s bytePlusStatus
	if err := json.Unmarshal(body, &s); err != nil {
		return Update{}, fmt.Errorf("byteplus status: %w", err)
	}

	jobErr := func(fallback string) *JobError {
		if s.Error != nil && (s.Error.Message != "" || s.Error.Code != "") {
			return &JobError{Code: s.Error.Code, Message: s.Error.Message}
		}
		return &JobError{Message: fallback}
	}

	switch strings.ToLower(s.Status) {
	case "queued":
		return Update{Status: StatusQueued}, nil
	case "running":
		return Update{Status: StatusRunning}, nil
	case "succeeded":
		if s.Content == nil || s.Content.VideoURL == "" {
			return Update{Status: StatusRunning}, nil
		}
		u := Update{Status: StatusSucceeded, VideoURL: s.Content.VideoURL}
		if s.UpdatedAt > 0 {
			exp := time.Unix(s.UpdatedAt, 0).UTC().Add(BytePlusURLLifetime)
			u.ExpiresAt = &exp
		}
		return u, nil
	case "failed":
		return Update{Status: StatusFailed, Error: jobErr("the vendor reported the render as failed")}, nil
	case "expired":
		return Update{Status: StatusExpired, Error: jobErr("the vendor task expired before it finished")}, nil
	case "cancelled", "canceled":
		return Update{Status: StatusCancelled, Error: jobErr("the vendor task was cancelled")}, nil
	case "":
		return Update{}, fmt.Errorf("byteplus status: empty status")
	}
	return Update{Status: StatusRunning}, nil
}
