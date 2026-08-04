package proxyapi

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

type fakeHealthTracker struct {
	successes int
	failures  int
	tee       int
}

func (f *fakeHealthTracker) ReportPromptSuccess(modelID common.Hash) { f.successes++ }
func (f *fakeHealthTracker) ReportPromptFailure(modelID common.Hash) { f.failures++ }
func (f *fakeHealthTracker) ReportTeeFailure(modelID common.Hash)    { f.tee++ }

func TestTrackPromptResult(t *testing.T) {
	modelID := common.HexToHash("0x01")

	tests := []struct {
		name          string
		transportErr  error
		engineStatus  int
		gotEngineErr  bool
		wantSuccesses int
		wantFailures  int
	}{
		{"clean completion counts as success", nil, 0, false, 1, 0},
		{"transport error counts as failure", errors.New("connection refused"), 0, false, 0, 1},
		{"backend 500 counts as failure", nil, 500, true, 0, 1},
		{"backend 402 (billing) counts as failure", nil, 402, true, 0, 1},
		{"backend 429 (rate limit) counts as failure", nil, 429, true, 0, 1},
		{"unknown status error counts as failure", nil, 0, true, 0, 1},
		{"client-fault 400 counts as neither", nil, 400, true, 0, 0},
		{"client-fault 422 counts as neither", nil, 422, true, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := &fakeHealthTracker{}
			r := &ProxyReceiver{modelHealth: tracker}
			r.trackPromptResult(modelID, tt.transportErr, tt.engineStatus, tt.gotEngineErr)
			if tracker.successes != tt.wantSuccesses {
				t.Errorf("successes = %d, want %d", tracker.successes, tt.wantSuccesses)
			}
			if tracker.failures != tt.wantFailures {
				t.Errorf("failures = %d, want %d", tracker.failures, tt.wantFailures)
			}
		})
	}
}

func TestTrackPromptResultNilTrackerIsNoop(t *testing.T) {
	r := &ProxyReceiver{}
	// must not panic when no tracker is wired (health checks disabled)
	r.trackPromptResult(common.HexToHash("0x01"), errors.New("boom"), 0, false)
}
