package aiengine

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/stretchr/testify/require"
)

// Renderful's Seedance rows have different limits: the base spec follows the
// catalog so a quote is never sold for a clip the vendor refuses.
func TestRenderfulBaseVideoSpecPerSeedanceModel(t *testing.T) {
	cases := []struct {
		model       string
		min, max    int
		resolutions []string
		audio       bool
	}{
		{"seedance-2.5-t2v", 4, 30, []string{"480p", "720p"}, true},
		{"seedance-2.0-t2v", 4, 15, []string{"480p", "720p"}, true},
		{"seedance-2.0-fast-t2v", 4, 15, []string{"480p", "720p"}, true},
		{"seedance-1.5-pro", 4, 12, []string{"480p", "720p"}, true},
		{"seedance-1.0-pro", 2, 12, []string{"720p", "1080p"}, false},
		{"kling-v2-master", 4, 30, []string{"480p", "720p"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.model, func(t *testing.T) {
			spec := NewRenderfulEngine(tc.model, "", "", &lib.LoggerMock{}, nil).BaseVideoSpec()
			require.Equal(t, tc.min, spec.MinDuration)
			require.Equal(t, tc.max, spec.MaxDuration)
			require.Equal(t, tc.resolutions, spec.Resolutions)
			require.Equal(t, tc.audio, spec.Audio)
			require.NoError(t, spec.Validate())
		})
	}
}

// A Seedance first-frame job on Renderful keeps the frame's ratio, and the
// task says so for the receipt.
func TestRenderfulSeedanceFirstFrameIsAdaptive(t *testing.T) {
	vendor := newVideoVendor(t, videoJSONReply(http.StatusAccepted, `{"id":"gen_1","status":"queued"}`))
	engine := NewRenderfulEngine("seedance-2.5-t2v", vendor.URL, "k", &lib.LoggerMock{}, nil)

	res := videoTestResolved()
	res.AspectRatio = "16:9"
	res.HasImage = true
	task, err := engine.CreateVideoJob(context.Background(), &VideoJobInput{Resolved: res, ImageURL: "https://img.example.com/a.png", Audio: true})
	require.NoError(t, err)
	require.Equal(t, "adaptive", task.AspectRatio)

	body := decodeVideoBody(t, vendor.Calls()[0].Body)
	require.Equal(t, "image-to-video", body["type"])
	require.Equal(t, "seedance-2.5-i2v", body["model"])
	require.Equal(t, "adaptive", body["aspect_ratio"])

	// Text to video keeps the asked ratio and reports no change.
	task, err = engine.CreateVideoJob(context.Background(), &VideoJobInput{Resolved: videoTestResolved(), Audio: true})
	require.NoError(t, err)
	require.Empty(t, task.AspectRatio)
}

func TestVideoTransportErrorHidesTheURL(t *testing.T) {
	for _, engine := range []VideoJobEngine{
		NewRenderfulEngine("seedance-2.5-t2v", "http://10.255.255.1:1", "k", &lib.LoggerMock{}, nil),
		NewBytePlusVideoEngine("dreamina-seedance-2-5-260628", "http://10.255.255.1:1", "k", &lib.LoggerMock{}, nil),
	} {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := engine.CreateVideoJob(ctx, &VideoJobInput{Resolved: videoTestResolved()})
		require.True(t, errors.Is(err, ErrVideoVendorUnreachable), "%v", err)
		require.NotContains(t, err.Error(), "10.255.255.1")

		_, err = engine.FetchVideoStatus(ctx, "task-1")
		require.True(t, errors.Is(err, ErrVideoVendorUnreachable), "%v", err)
		require.NotContains(t, err.Error(), "10.255.255.1")
	}
}
