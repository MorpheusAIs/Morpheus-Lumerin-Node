package videojob

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// MaxImageBytes bounds an inline first frame. The job travels to the provider
// in one RPC message, so the cap keeps that message reasonable.
const MaxImageBytes = 10 << 20

// Image is a decoded inline first frame.
type Image struct {
	MimeType string
	Data     []byte
}

// DataURI renders the image as data:<mime>;base64,<data>.
func (i *Image) DataURI() string {
	return "data:" + i.MimeType + ";base64," + base64.StdEncoding.EncodeToString(i.Data)
}

// DecodeImage reads image_b64: plain base64, or a data URI. The type is taken
// from the bytes, not from the URI, and must be JPEG, PNG or WebP.
func DecodeImage(raw string) (*Image, error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "data:") {
		comma := strings.IndexByte(raw, ',')
		if comma < 0 || !strings.HasSuffix(raw[:comma], ";base64") {
			return nil, fmt.Errorf("%w: image_b64 data URI must be base64", ErrInvalidRequest)
		}
		raw = raw[comma+1:]
	}

	if base64.StdEncoding.DecodedLen(len(raw)) > MaxImageBytes+3 {
		return nil, fmt.Errorf("%w: image_b64 is larger than %d bytes", ErrInvalidRequest, MaxImageBytes)
	}
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: image_b64 is not valid base64", ErrInvalidRequest)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: image_b64 is empty", ErrInvalidRequest)
	}
	if len(data) > MaxImageBytes {
		return nil, fmt.Errorf("%w: image_b64 is larger than %d bytes", ErrInvalidRequest, MaxImageBytes)
	}

	mime := http.DetectContentType(data)
	switch mime {
	case "image/jpeg", "image/png", "image/webp":
	default:
		return nil, fmt.Errorf("%w: image_b64 must be JPEG, PNG or WebP, got %s", ErrInvalidRequest, mime)
	}
	return &Image{MimeType: mime, Data: data}, nil
}
