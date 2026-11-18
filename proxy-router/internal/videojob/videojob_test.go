package videojob

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func testSpec() Spec {
	return Spec{
		ModelName:    "Seedance 2.5",
		MinDuration:  4,
		MaxDuration:  30,
		AspectRatios: []string{"16:9", "9:16", "1:1"},
		Resolutions:  []string{"480p", "720p"},
		Audio:        true,
		Image:        true,
		ImageInline:  true,
		Defaults:     Defaults{DurationSeconds: 5, AspectRatio: "16:9", Resolution: "720p"},
		Quote: Quote{
			SecondsPerWorkUnit: 75,
			ResolutionWeights:  map[string]float64{"480p": 1, "720p": 2.25},
			AudioWeight:        1.2,
			ImageWeight:        1.1,
			MinSessionSeconds:  300,
			MaxSessionSeconds:  7200,
		},
	}
}

func intPtr(v int) *int    { return &v }
func boolPtr(v bool) *bool { return &v }

func TestSpecValidate(t *testing.T) {
	require.NoError(t, testSpec().Validate())

	bad := testSpec()
	bad.Defaults.Resolution = "1080p"
	require.Error(t, bad.Validate())

	bad = testSpec()
	bad.Audio = false
	bad.Defaults.GenerateAudio = true
	require.Error(t, bad.Validate())

	bad = testSpec()
	bad.Quote.SecondsPerWorkUnit = 0
	require.Error(t, bad.Validate())
}

func TestResolveAppliesDefaults(t *testing.T) {
	g := testSpec().Guide()
	r, err := g.Resolve(&CreateRequest{Prompt: "  a harbour at dusk  "})
	require.NoError(t, err)
	require.Equal(t, Resolved{
		Prompt:          "a harbour at dusk",
		DurationSeconds: 5,
		AspectRatio:     "16:9",
		Resolution:      "720p",
		GenerateAudio:   false,
	}, r)
}

func TestResolveRejectsOutOfRange(t *testing.T) {
	g := testSpec().Guide()
	cases := []CreateRequest{
		{Prompt: ""},
		{Prompt: "x", DurationSeconds: intPtr(3)},
		{Prompt: "x", DurationSeconds: intPtr(31)},
		{Prompt: "x", AspectRatio: "2:1"},
		{Prompt: "x", Resolution: "1080p"},
	}
	for _, c := range cases {
		_, err := g.Resolve(&c)
		require.ErrorIs(t, err, ErrInvalidRequest, "%+v", c)
	}
}

func TestResolveRespectsOfferedFields(t *testing.T) {
	spec := testSpec()
	spec.Audio = false
	spec.Image = false
	spec.ImageInline = false
	g := spec.Guide()

	_, err := g.Resolve(&CreateRequest{Prompt: "x", GenerateAudio: boolPtr(true)})
	require.ErrorIs(t, err, ErrInvalidRequest)
	_, err = g.Resolve(&CreateRequest{Prompt: "x", GenerateAudio: boolPtr(false)})
	require.NoError(t, err)
	_, err = g.Resolve(&CreateRequest{Prompt: "x", ImageURL: "https://example.com/a.png"})
	require.ErrorIs(t, err, ErrInvalidRequest)
}

func TestGuideRoundTripsAsJSON(t *testing.T) {
	g := testSpec().Guide()
	raw, err := json.Marshal(g)
	require.NoError(t, err)

	var back Guide
	require.NoError(t, json.Unmarshal(raw, &back))
	require.Equal(t, ObjectGuide, back.Object)

	// The consumer resolves against the guide it received, with the same result.
	want, err := g.Resolve(&CreateRequest{Prompt: "x", DurationSeconds: intPtr(10), Resolution: "480p"})
	require.NoError(t, err)
	got, err := back.Resolve(&CreateRequest{Prompt: "x", DurationSeconds: intPtr(10), Resolution: "480p"})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestQuoteSessionSeconds(t *testing.T) {
	q := testSpec().Quote

	// 4s 480p: 4 × 1 × 75 = 300, the floor.
	secs, err := q.SessionSeconds(Resolved{DurationSeconds: 4, Resolution: "480p"})
	require.NoError(t, err)
	require.EqualValues(t, 300, secs)

	// Below the floor is clamped up.
	secs, err = q.SessionSeconds(Resolved{DurationSeconds: 2, Resolution: "480p"})
	require.NoError(t, err)
	require.EqualValues(t, 300, secs)

	// 10s 720p with audio and image: 10 × 2.25 × 1.2 × 1.1 × 75 = 2227.5 → 2228.
	secs, err = q.SessionSeconds(Resolved{DurationSeconds: 10, Resolution: "720p", GenerateAudio: true, HasImage: true})
	require.NoError(t, err)
	require.EqualValues(t, 2228, secs)

	// An unknown resolution weighs 1.
	secs, err = q.SessionSeconds(Resolved{DurationSeconds: 8, Resolution: "1080p"})
	require.NoError(t, err)
	require.EqualValues(t, 600, secs)

	// Above the ceiling is refused.
	_, err = q.SessionSeconds(Resolved{DurationSeconds: 60, Resolution: "720p"})
	require.ErrorIs(t, err, ErrQuoteTooLong)
}

func TestParseRenderfulStatus(t *testing.T) {
	u, err := ParseStatus(VendorRenderful, []byte(`{"id":"gen_1","status":"queued"}`))
	require.NoError(t, err)
	require.Equal(t, StatusQueued, u.Status)

	u, err = ParseStatus(VendorRenderful, []byte(`{"id":"gen_1","status":"processing"}`))
	require.NoError(t, err)
	require.Equal(t, StatusRunning, u.Status)

	// Completed with no outputs yet is not done.
	u, err = ParseStatus(VendorRenderful, []byte(`{"id":"gen_1","status":"completed","outputs":[]}`))
	require.NoError(t, err)
	require.Equal(t, StatusRunning, u.Status)

	u, err = ParseStatus(VendorRenderful, []byte(`{"id":"gen_1","status":"completed","outputs":["https://cdn.example/out.mp4"]}`))
	require.NoError(t, err)
	require.Equal(t, StatusSucceeded, u.Status)
	require.Equal(t, "https://cdn.example/out.mp4", u.VideoURL)
	require.Nil(t, u.ExpiresAt)

	u, err = ParseStatus(VendorRenderful, []byte(`{"id":"gen_1","status":"completed","output":"https://cdn.example/one.mp4"}`))
	require.NoError(t, err)
	require.Equal(t, "https://cdn.example/one.mp4", u.VideoURL)

	u, err = ParseStatus(VendorRenderful, []byte(`{"id":"gen_1","status":"failed","error":"content policy"}`))
	require.NoError(t, err)
	require.Equal(t, StatusFailed, u.Status)
	require.Equal(t, "content policy", u.Error.Message)

	_, err = ParseStatus(VendorRenderful, []byte(`not json`))
	require.Error(t, err)
}

func TestParseBytePlusStatus(t *testing.T) {
	u, err := ParseStatus(VendorBytePlus, []byte(`{"id":"cgt-1","status":"running","created_at":1761000000,"updated_at":1761000010}`))
	require.NoError(t, err)
	require.Equal(t, StatusRunning, u.Status)

	u, err = ParseStatus(VendorBytePlus, []byte(`{
		"id":"cgt-1","model":"dreamina-seedance-2-5-260628","status":"succeeded",
		"content":{"video_url":"https://ark-content-generation-ap-southeast-1.tos-ap-southeast-1.volces.com/x.mp4"},
		"error":null,"created_at":1761000000,"updated_at":1761000600}`))
	require.NoError(t, err)
	require.Equal(t, StatusSucceeded, u.Status)
	require.Equal(t, "https://ark-content-generation-ap-southeast-1.tos-ap-southeast-1.volces.com/x.mp4", u.VideoURL)
	require.Equal(t, time.Unix(1761000600, 0).UTC().Add(24*time.Hour), *u.ExpiresAt)

	u, err = ParseStatus(VendorBytePlus, []byte(`{"id":"cgt-1","status":"failed","error":{"code":"OutputVideoSensitiveContentDetected","message":"sensitive"}}`))
	require.NoError(t, err)
	require.Equal(t, StatusFailed, u.Status)
	require.Equal(t, &JobError{Code: "OutputVideoSensitiveContentDetected", Message: "sensitive"}, u.Error)

	u, err = ParseStatus(VendorBytePlus, []byte(`{"id":"cgt-1","status":"expired"}`))
	require.NoError(t, err)
	require.Equal(t, StatusExpired, u.Status)
	require.NotEmpty(t, u.Error.Message)

	_, err = ParseStatus("acme", []byte(`{}`))
	require.Error(t, err)
}

func TestValidateRemoteURL(t *testing.T) {
	for _, good := range []string{
		"https://api.renderful.ai/api/v1/generations/gen_1",
		"https://cdn.example.com./clip.mp4",
		"https://1.2.3.example.com/clip.mp4",
		"https://ark-acg-ap-southeast-1.tos-ap-southeast-1.volces.com/x.mp4?X-Tos-Expires=86400",
	} {
		require.NoError(t, ValidateRemoteURL(good), good)
	}
	for _, bad := range []string{
		"http://api.renderful.ai/x",
		"https://user:pw@api.renderful.ai/x",
		"https://127.0.0.1/x",
		"https://10.0.0.5/x",
		"https://169.254.169.254/latest/meta-data",
		"https://[::1]/x",
		"file:///etc/passwd",
		"/relative",
		// names that are local by convention
		"https://localhost/x",
		"https://LOCALHOST:8082/x",
		"https://node.localhost/x",
		"https://cdn.internal/x",
		"https://printer.local/x",
		"https://intranet/x",
		// IPv4 spellings a browser dials that net/netip does not parse
		"https://127.1/x",
		"https://2130706433/x",
		"https://0x7f000001/x",
		"https://0177.0.0.1/x",
		"https://10.1/x",
		"https://127.0.0.1./x",
		"https://127.0x.0.1/x",
		"https://0x.0x.0x.0x/x",
		"https://0x7f.1/x",
		"https://127.0.0.1../x",
		"https://cdn..example.com/x",
		"https://localhost.localdomain/x",
		"https://\uff11\uff12\uff17.0.0.1/x",
		"https://127.0.0.1\u3002/x",
		// IPv6 forms that reach a private host
		"https://[::7f00:1]/x",
		"https://[::ffff:10.0.0.5]/x",
		"https://[fec0::1]/x",
		"https://[fec0::1%25eth0]/x",
		"https://[2001:0:4136:e378:8000:63bf:3fff:fdd2]/x",
	} {
		require.ErrorIs(t, ValidateRemoteURL(bad), ErrRemoteURL, bad)
	}
}

func TestIsPublicAddr(t *testing.T) {
	require.True(t, isPublicAddr(netip.MustParseAddr("8.8.8.8")))
	for _, addr := range []string{"127.0.0.1", "::ffff:192.168.1.1", "100.64.1.1", "fd00::1", "64:ff9b::a00:1", "2001::1", "::7f00:1", "fec0::1", "2606:4700::1%en0"} {
		require.False(t, isPublicAddr(netip.MustParseAddr(addr)), addr)
	}
}

var pngHeader = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 13, 'I', 'H', 'D', 'R'}

func TestDecodeImage(t *testing.T) {
	b64 := base64.StdEncoding.EncodeToString(pngHeader)

	img, err := DecodeImage(b64)
	require.NoError(t, err)
	require.Equal(t, "image/png", img.MimeType)
	require.Equal(t, "data:image/png;base64,"+b64, img.DataURI())

	img, err = DecodeImage("data:image/jpeg;base64," + b64)
	require.NoError(t, err)
	require.Equal(t, "image/png", img.MimeType, "type comes from the bytes")

	_, err = DecodeImage(base64.StdEncoding.EncodeToString([]byte("hello, not an image")))
	require.True(t, errors.Is(err, ErrInvalidRequest))

	_, err = DecodeImage("!!!")
	require.ErrorIs(t, err, ErrInvalidRequest)

	_, err = DecodeImage("data:image/png," + b64)
	require.ErrorIs(t, err, ErrInvalidRequest)
}
