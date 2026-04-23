package attestation

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

const (
	DefaultRegistryURL             = "https://raw.githubusercontent.com/scrtlabs/secretvm-verify/main/artifacts_registry/tdx.csv"
	DefaultRegistryRefreshInterval = 1 * time.Hour
)

type TdxArtifactEntry struct {
	TemplateName string
	VMType       string
	ArtifactsVer string
	MRTD         string
	RTMR0        string
	RTMR1        string
	RTMR2        string
	RootfsData   string
	HostID       string
}

type ArtifactRegistry struct {
	registryURL     string
	refreshInterval time.Duration
	client          *http.Client
	log             lib.ILogger

	mu          sync.RWMutex
	entries     []TdxArtifactEntry
	lastFetched time.Time
}

func NewArtifactRegistry(registryURL string, refreshInterval time.Duration, log lib.ILogger) *ArtifactRegistry {
	if registryURL == "" {
		registryURL = DefaultRegistryURL
	}
	if refreshInterval == 0 {
		refreshInterval = DefaultRegistryRefreshInterval
	}
	return &ArtifactRegistry{
		registryURL:     registryURL,
		refreshInterval: refreshInterval,
		client:          &http.Client{Timeout: 30 * time.Second},
		log:             log,
	}
}

func (r *ArtifactRegistry) Start(ctx context.Context) {
	if err := r.fetch(ctx); err != nil {
		r.log.Warnf("initial artifact registry fetch failed: %v", err)
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
					r.log.Warnf("artifact registry refresh failed, keeping stale cache: %v", err)
				}
			}
		}
	}()
}

func (r *ArtifactRegistry) fetch(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.registryURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	entries := parseTdxRegistryCSV(string(body))

	r.mu.Lock()
	r.entries = entries
	r.lastFetched = time.Now()
	r.mu.Unlock()

	r.log.Infof("artifact registry loaded %d entries", len(entries))
	return nil
}

func parseTdxRegistryCSV(content string) []TdxArtifactEntry {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) < 2 {
		return nil
	}

	var entries []TdxArtifactEntry
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) < 9 {
			continue
		}
		entries = append(entries, TdxArtifactEntry{
			TemplateName: strings.TrimSpace(fields[0]),
			VMType:       strings.TrimSpace(fields[1]),
			ArtifactsVer: strings.TrimSpace(fields[2]),
			MRTD:         normalizeHex(fields[3]),
			RTMR0:        normalizeHex(fields[4]),
			RTMR1:        normalizeHex(fields[5]),
			RTMR2:        normalizeHex(fields[6]),
			RootfsData:   normalizeHex(fields[7]),
			HostID:       normalizeHex(fields[8]),
		})
	}
	return entries
}

func normalizeHex(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.TrimPrefix(s, "0x")
	return s
}

func (r *ArtifactRegistry) FindMatchingArtifacts(mrtd, rtmr0, rtmr1, rtmr2 string) []TdxArtifactEntry {
	mrtd = normalizeHex(mrtd)
	rtmr0 = normalizeHex(rtmr0)
	rtmr1 = normalizeHex(rtmr1)
	rtmr2 = normalizeHex(rtmr2)

	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []TdxArtifactEntry
	for _, e := range r.entries {
		if e.MRTD == mrtd && e.RTMR0 == rtmr0 && e.RTMR1 == rtmr1 && e.RTMR2 == rtmr2 {
			matched = append(matched, e)
		}
	}
	return matched
}

func (r *ArtifactRegistry) PickNewestVersion(entries []TdxArtifactEntry) *TdxArtifactEntry {
	if len(entries) == 0 {
		return nil
	}

	best := 0
	for i := 1; i < len(entries); i++ {
		if compareSemver(entries[i].ArtifactsVer, entries[best].ArtifactsVer) > 0 {
			best = i
		}
	}
	result := entries[best]
	return &result
}

// compareSemver returns >0 if a > b, <0 if a < b, 0 if equal.
// Release versions (no pre-release) sort after pre-release.
func compareSemver(a, b string) int {
	aMajor, aMinor, aPatch, aPre := parseSemver(a)
	bMajor, bMinor, bPatch, bPre := parseSemver(b)

	if aMajor != bMajor {
		return aMajor - bMajor
	}
	if aMinor != bMinor {
		return aMinor - bMinor
	}
	if aPatch != bPatch {
		return aPatch - bPatch
	}

	// "" (release) > any non-empty pre-release string
	if aPre == "" && bPre == "" {
		return 0
	}
	if aPre == "" {
		return 1
	}
	if bPre == "" {
		return -1
	}
	if aPre < bPre {
		return -1
	}
	if aPre > bPre {
		return 1
	}
	return 0
}

func parseSemver(v string) (major, minor, patch int, pre string) {
	v = strings.TrimPrefix(v, "v")

	preParts := strings.SplitN(v, "-", 2)
	core := preParts[0]
	if len(preParts) > 1 {
		pre = preParts[1]
	}

	parts := strings.Split(core, ".")
	if len(parts) >= 1 {
		major, _ = strconv.Atoi(parts[0])
	}
	if len(parts) >= 2 {
		minor, _ = strconv.Atoi(parts[1])
	}
	if len(parts) >= 3 {
		patch, _ = strconv.Atoi(parts[2])
	}
	return
}

func (r *ArtifactRegistry) IsLoaded() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return !r.lastFetched.IsZero()
}
