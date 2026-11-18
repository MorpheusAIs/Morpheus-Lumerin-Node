// Package videojob holds the documents a hosted text-to-video job is made of:
// what a caller submits (CreateRequest), what the provider returns once the
// vendor accepted the task (Receipt), what the consumer tracks while it asks
// the provider for the status (Job) and what a provider advertises before
// anyone opens a session (Guide). Vendors differ in field names and status words; these
// documents do not.
package videojob

import (
	"time"
)

const (
	ObjectJob   = "video.job"
	ObjectGuide = "video.guide"
)

// Vendor names a status body format. It travels in the receipt so the
// consumer knows how to read the vendor's status response.
const (
	VendorRenderful = "renderful"
	VendorBytePlus  = "byteplus"
)

// Job status values, whatever the vendor calls them.
const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
	StatusExpired   = "expired"
	StatusCancelled = "cancelled"
)

// IsTerminal reports whether a job in this status will not change again.
func IsTerminal(status string) bool {
	switch status {
	case StatusSucceeded, StatusFailed, StatusExpired, StatusCancelled:
		return true
	}
	return false
}

// CreateRequest is the VideoJobCreate document. Only Prompt is required;
// omitted fields take the guide's defaults.
type CreateRequest struct {
	// Model is the on-chain model id POST /v1/video/quote prices the job for.
	// A submitted job takes its model from its session, and the provider
	// serves the session's model.
	Model           string `json:"model,omitempty"`
	Prompt          string `json:"prompt"`
	DurationSeconds *int   `json:"duration_seconds,omitempty"`
	AspectRatio     string `json:"aspect_ratio,omitempty"`
	Resolution      string `json:"resolution,omitempty"`
	GenerateAudio   *bool  `json:"generate_audio,omitempty"`
	// ImageURL is a first frame the vendor can fetch itself.
	ImageURL string `json:"image_url,omitempty"`
	// ImageB64 is a first frame sent inline, as plain base64 or a data URI.
	// Ignored when ImageURL is set.
	ImageB64 string `json:"image_b64,omitempty"`
}

// HasImage reports whether the request carries a first frame.
func (r *CreateRequest) HasImage() bool {
	return r.ImageURL != "" || r.ImageB64 != ""
}

// Resolved is the request after defaults were applied and limits checked.
// The receipt echoes it so the caller can see what was assumed.
type Resolved struct {
	Prompt          string `json:"prompt"`
	DurationSeconds int    `json:"duration_seconds"`
	AspectRatio     string `json:"aspect_ratio"`
	Resolution      string `json:"resolution"`
	GenerateAudio   bool   `json:"generate_audio"`
	HasImage        bool   `json:"has_image"`
}

// Receipt is what the provider returns once the vendor accepted the task.
// The provider keeps its vendor key: the consumer reads the job's status
// through the provider (session.video.status), never from the vendor, since
// neither vendor offers a key that can only read.
type Receipt struct {
	Object string `json:"object"`
	Status string `json:"status"`
	// Vendor names the status body format (VendorRenderful, VendorBytePlus).
	Vendor       string   `json:"vendor"`
	VendorTaskID string   `json:"vendor_task_id"`
	QuoteSeconds int64    `json:"quote_seconds"`
	ClientJobID  string   `json:"client_job_id"`
	Resolved     Resolved `json:"resolved"`
}

// JobError is the vendor's reason for a failed job.
type JobError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// Job is the consumer's local record of a video job, as the local API
// returns it.
type Job struct {
	Object       string `json:"object"`
	ID           string `json:"id"`
	Model        string `json:"model"`
	Status       string `json:"status"`
	SessionID    string `json:"session_id"`
	QuoteSeconds int64  `json:"quote_seconds"`
	Vendor       string `json:"vendor"`
	// VendorTaskID is empty until the provider confirmed the job.
	VendorTaskID string   `json:"vendor_task_id"`
	Resolved     Resolved `json:"resolved"`
	// VideoURL is the vendor's MP4 link, present once the job succeeded. It
	// is a bearer link: anyone holding it can download until it expires.
	VideoURL string `json:"video_url,omitempty"`
	// ExpiresAt is when VideoURL stops working, when the vendor documents it.
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Error     *JobError  `json:"error"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
