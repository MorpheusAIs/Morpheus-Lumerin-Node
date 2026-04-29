package attestation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

const (
	DefaultSevRegistryURL = "https://raw.githubusercontent.com/scrtlabs/secretvm-verify/main/artifacts_registry/sev.json"
)

type SevArtifactRegistry struct {
	registryURL     string
	refreshInterval time.Duration
	client          *http.Client
	log             lib.ILogger

	mu          sync.RWMutex
	entries     []SevArtifactEntry
	lastFetched time.Time
}

func NewSevArtifactRegistry(registryURL string, refreshInterval time.Duration, log lib.ILogger) *SevArtifactRegistry {
	if registryURL == "" {
		registryURL = DefaultSevRegistryURL
	}
	if refreshInterval == 0 {
		refreshInterval = DefaultRegistryRefreshInterval
	}
	return &SevArtifactRegistry{
		registryURL:     registryURL,
		refreshInterval: refreshInterval,
		client:          &http.Client{Timeout: 30 * time.Second},
		log:             log,
	}
}

func (r *SevArtifactRegistry) Start(ctx context.Context) {
	if err := r.fetch(ctx); err != nil {
		r.log.Warnf("initial SEV artifact registry fetch failed: %v", err)
	}

	go func() {
		ticker := time.NewTicker(r.refreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := r.fetch(ctx); err != nil {
					r.log.Warnf("SEV artifact registry refresh failed, keeping stale cache: %v", err)
				}
			}
		}
	}()
}

func (r *SevArtifactRegistry) fetch(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.registryURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch SEV registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SEV registry returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	var entries []SevArtifactEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return fmt.Errorf("parse SEV registry JSON: %w", err)
	}

	r.mu.Lock()
	r.entries = entries
	r.lastFetched = time.Now()
	r.mu.Unlock()

	r.log.Infof("SEV artifact registry loaded %d entries", len(entries))
	return nil
}

func (r *SevArtifactRegistry) FindByVMType(vmType string) []SevArtifactEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []SevArtifactEntry
	for _, e := range r.entries {
		if e.VMType == vmType {
			matched = append(matched, e)
		}
	}
	return matched
}

func (r *SevArtifactRegistry) AllEntries() []SevArtifactEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]SevArtifactEntry, len(r.entries))
	copy(out, r.entries)
	return out
}

func (r *SevArtifactRegistry) IsLoaded() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return !r.lastFetched.IsZero()
}
