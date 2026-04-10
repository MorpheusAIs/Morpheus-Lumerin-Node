package attestation

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

// BackendGoldenSource provides golden register values for LLM backend TEE verification.
// The source of these values is TBD -- could be GHCR OCI attestation, a Secret AI API,
// or per-model configuration. Until defined, use NoopGoldenSource.
type BackendGoldenSource interface {
	FetchGoldenValues(ctx context.Context, modelID string, attestationURL string) (*GoldenValues, error)
}

// NoopGoldenSource always returns nil, skipping golden register comparison.
// Used as a placeholder until the real golden values source is defined.
type NoopGoldenSource struct{}

func (n *NoopGoldenSource) FetchGoldenValues(_ context.Context, _ string, _ string) (*GoldenValues, error) {
	return nil, nil
}

type BackendAttestationStatus string

const (
	StatusPassed  BackendAttestationStatus = "passed"
	StatusFailed  BackendAttestationStatus = "failed"
	StatusUnknown BackendAttestationStatus = "unknown"
)

// BackendAttestationSnapshot holds the cached attestation state for a single model backend.
type BackendAttestationSnapshot struct {
	ModelID           string                   `json:"modelId"`
	AttestationURL    string                   `json:"attestationUrl"`
	CPUQuoteHash      string                   `json:"-"`
	GPUQuoteHash      string                   `json:"-"`
	TLSFingerprint    string                   `json:"-"`
	CPUReportData     string                   `json:"-"`
	GPUReportData     string                   `json:"-"`
	TEEType           TEEType                  `json:"teeType,omitempty"`
	VerifiedAt        time.Time                `json:"verifiedAt"`
	Status            BackendAttestationStatus `json:"status"`
	Error             string                   `json:"error,omitempty"`
	WorkloadStatus    string                   `json:"workloadStatus,omitempty"`
	VMTemplateName    string                   `json:"vmTemplateName,omitempty"`
	ArtifactsVersion  string                   `json:"artifactsVersion,omitempty"`
	DockerComposeHash string                   `json:"dockerComposeHash,omitempty"`
}

// BackendVerifier manages TEE attestation verification of LLM backend endpoints.
// It performs CPU+GPU attestation, caches results per model, and provides
// fast-verify for the per-prompt hot path.
type BackendVerifier struct {
	portalClient      *http.Client
	attestationClient *http.Client
	portalURL         string
	goldenSource      BackendGoldenSource
	nrasVerifier      *NRASVerifier
	artifactRegistry  *ArtifactRegistry
	log               lib.ILogger

	mu    sync.RWMutex
	cache map[string]*BackendAttestationSnapshot
}

func NewBackendVerifier(portalURL string, goldenSource BackendGoldenSource, registry *ArtifactRegistry, log lib.ILogger) *BackendVerifier {
	if portalURL == "" {
		portalURL = DefaultPortalURL
	}
	if goldenSource == nil {
		goldenSource = &NoopGoldenSource{}
	}

	return &BackendVerifier{
		portalClient:      NewPortalHTTPClient(),
		attestationClient: NewAttestationHTTPClient(),
		portalURL:         portalURL,
		goldenSource:      goldenSource,
		nrasVerifier:      NewNRASVerifier(log),
		artifactRegistry:  registry,
		log:               log,
		cache:             make(map[string]*BackendAttestationSnapshot),
	}
}

// AttestBackend performs full CPU+GPU attestation of a backend LLM TEE endpoint.
// On success the result is cached for fast-verify. On failure the snapshot is
// stored with StatusFailed so the health endpoint can report it.
func (bv *BackendVerifier) AttestBackend(ctx context.Context, modelID string, attestationURL string) error {
	bv.log.Infof("backend attestation: starting full verification for model %s at %s", modelID, attestationURL)

	// 1. Fetch CPU quote
	cpuURL := attestationURL + "/cpu"
	cpuQuote, tlsFingerprint, err := LoadAttestationQuote(ctx, bv.attestationClient, cpuURL)
	if err != nil {
		bv.storeFailure(modelID, attestationURL, fmt.Sprintf("CPU quote fetch failed: %s", err))
		return fmt.Errorf("failed to load CPU attestation quote from %s: %w", cpuURL, err)
	}
	bv.log.Infof("backend attestation: fetched CPU quote from %s, TLS fingerprint: %s", cpuURL, tlsFingerprint)

	if tlsFingerprint == "" {
		bv.storeFailure(modelID, attestationURL, "no TLS certificate from CPU endpoint")
		return fmt.Errorf("no TLS peer certificate received from %s", cpuURL)
	}

	// 2. Verify CPU quote via portal
	cpuResult, err := VerifyQuote(ctx, bv.portalClient, bv.portalURL, cpuQuote)
	if err != nil {
		bv.storeFailure(modelID, attestationURL, fmt.Sprintf("CPU quote portal verification failed: %s", err))
		return fmt.Errorf("CPU attestation quote verification failed: %w", err)
	}
	if !cpuResult.Valid {
		bv.storeFailure(modelID, attestationURL, fmt.Sprintf("CPU attestation invalid: %s", cpuResult.Error))
		return fmt.Errorf("CPU attestation invalid (%s): %s", cpuResult.Type, cpuResult.Error)
	}
	bv.log.Infof("backend attestation: CPU quote valid (type: %s) for model %s", cpuResult.Type, modelID)

	// 3. Verify TLS binding (first half of reportData = TLS cert fingerprint)
	if err := VerifyTLSBinding(tlsFingerprint, cpuResult.ReportData); err != nil {
		bv.storeFailure(modelID, attestationURL, fmt.Sprintf("TLS binding failed: %s", err))
		return fmt.Errorf("CPU TLS binding verification failed: %w", err)
	}
	bv.log.Infof("backend attestation: TLS binding verified for model %s", modelID)

	// 3a. Workload verification (docker-compose vs attestation quote)
	var workloadResult *WorkloadResult
	if bv.artifactRegistry != nil && bv.artifactRegistry.IsLoaded() {
		dockerCompose, composeErr := bv.fetchDockerCompose(ctx, attestationURL)
		if composeErr != nil {
			bv.log.Warnf("backend attestation: could not fetch docker-compose for model %s: %s (workload verification skipped)", modelID, composeErr)
		} else {
			result := VerifyWorkload(bv.artifactRegistry, cpuQuote, dockerCompose, bv.log)
			workloadResult = &result
			switch result.Status {
			case WorkloadAuthentic:
				bv.log.Infof("backend attestation: workload verified for model %s (template=%s, version=%s, env=%s)", modelID, result.TemplateName, result.ArtifactsVer, result.Env)
			case WorkloadAuthenticMismatch:
				bv.storeFailure(modelID, attestationURL, "docker-compose does not match attestation (authentic VM but wrong workload)")
				return fmt.Errorf("workload verification failed for model %s: docker-compose does not match attestation", modelID)
			case WorkloadNotAuthentic:
				bv.storeFailure(modelID, attestationURL, "VM is not an authentic SecretVM (MRTD/RTMR values not in registry)")
				return fmt.Errorf("workload verification failed for model %s: not an authentic SecretVM", modelID)
			}
		}
	} else {
		bv.log.Infof("backend attestation: artifact registry not available, skipping workload verification for model %s", modelID)
	}

	// 4. Fetch GPU attestation data (JSON with nonce, arch, evidence_list)
	gpuURL := attestationURL + "/gpu"
	gpuRawJSON, _, err := LoadAttestationQuote(ctx, bv.attestationClient, gpuURL)
	if err != nil {
		bv.storeFailure(modelID, attestationURL, fmt.Sprintf("GPU quote fetch failed: %s", err))
		return fmt.Errorf("failed to load GPU attestation data from %s: %w", gpuURL, err)
	}
	bv.log.Infof("backend attestation: fetched GPU attestation data from %s", gpuURL)

	// 5. Parse GPU attestation JSON and extract nonce
	gpuData, err := ParseGPUAttestationData(gpuRawJSON)
	if err != nil {
		bv.storeFailure(modelID, attestationURL, fmt.Sprintf("GPU attestation data parse failed: %s", err))
		return fmt.Errorf("failed to parse GPU attestation data: %w", err)
	}
	bv.log.Infof("backend attestation: GPU arch=%s, nonce=%s, evidences=%d", gpuData.Arch, gpuData.Nonce, len(gpuData.EvidenceList))

	// 6. Verify CPU-GPU binding: second half of CPU reportData should be the GPU nonce
	if err := VerifyCPUGPUBinding(cpuResult.ReportData, gpuData.Nonce); err != nil {
		bv.storeFailure(modelID, attestationURL, fmt.Sprintf("CPU-GPU binding failed: %s", err))
		return fmt.Errorf("CPU-GPU binding verification failed: %w", err)
	}
	bv.log.Infof("backend attestation: CPU-GPU binding verified for model %s", modelID)

	// 7. Verify GPU evidence via NVIDIA Remote Attestation Service (NRAS)
	// NRAS may be unreachable from some networks (403/timeout). GPU trust is already
	// established via CPU-GPU nonce binding (step 6), so NRAS failure is non-fatal.
	nrasResult, err := bv.nrasVerifier.VerifyGPU(ctx, gpuData)
	if err != nil {
		bv.log.Warnf("backend attestation: NRAS GPU verification failed for model %s (non-fatal): %s", modelID, err)
	} else {
		bv.log.Infof("backend attestation: NRAS verified GPU for model %s (%d GPU tokens received)", modelID, len(nrasResult.GPUTokens))
	}

	// 8. Golden values comparison (placeholder -- NoopGoldenSource skips this)
	golden, err := bv.goldenSource.FetchGoldenValues(ctx, modelID, attestationURL)
	if err != nil {
		bv.storeFailure(modelID, attestationURL, fmt.Sprintf("golden values fetch failed: %s", err))
		return fmt.Errorf("failed to fetch golden values for model %s: %w", modelID, err)
	}
	if golden != nil {
		if err := CompareRegisters(cpuResult, golden, bv.log); err != nil {
			bv.storeFailure(modelID, attestationURL, fmt.Sprintf("register mismatch: %s", err))
			return err
		}
		bv.log.Infof("backend attestation: golden values match for model %s", modelID)
	} else {
		// bv.log.Infof("backend attestation: golden values comparison skipped (no source configured) for model %s", modelID)
	}

	// 9. Cache the successful attestation
	cpuHash := fmt.Sprintf("%x", sha256.Sum256([]byte(cpuQuote)))
	gpuHash := fmt.Sprintf("%x", sha256.Sum256([]byte(gpuRawJSON)))
	now := time.Now()

	snapshot := &BackendAttestationSnapshot{
		ModelID:        modelID,
		AttestationURL: attestationURL,
		CPUQuoteHash:   cpuHash,
		GPUQuoteHash:   gpuHash,
		TLSFingerprint: tlsFingerprint,
		CPUReportData:  cpuResult.ReportData,
		GPUReportData:  gpuData.Nonce,
		TEEType:        cpuResult.Type,
		VerifiedAt:     now,
		Status:         StatusPassed,
	}

	if workloadResult != nil {
		snapshot.WorkloadStatus = string(workloadResult.Status)
		snapshot.VMTemplateName = workloadResult.TemplateName
		snapshot.ArtifactsVersion = workloadResult.ArtifactsVer
	}

	bv.mu.Lock()
	bv.cache[modelID] = snapshot
	bv.mu.Unlock()

	bv.log.Infof("backend attestation: cached verified snapshot for model %s", modelID)
	return nil
}

// FastVerifyBackend performs a lightweight per-request check.
// Always re-fetches /cpu and compares sha256(quote) + TLS fingerprint against
// the cached attestation snapshot (~50ms). If the quote hash changes, triggers
// full re-attestation. TLS fingerprint mismatch is an immediate error (possible MITM).
func (bv *BackendVerifier) FastVerifyBackend(ctx context.Context, modelID string) error {
	bv.mu.RLock()
	snapshot, exists := bv.cache[modelID]
	bv.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no attestation snapshot for model %s", modelID)
	}

	if snapshot.Status != StatusPassed {
		bv.log.Infof("LLM attestation status is %s for model %s (%s), retrying full attestation", snapshot.Status, modelID, snapshot.Error)
		return bv.AttestBackend(ctx, modelID, snapshot.AttestationURL)
	}

	cpuURL := snapshot.AttestationURL + "/cpu"
	cpuQuote, tlsFingerprint, err := LoadAttestationQuote(ctx, bv.attestationClient, cpuURL)
	if err != nil {
		return fmt.Errorf("LLM fast-verify failed for model %s: %w", modelID, err)
	}

	currentHash := fmt.Sprintf("%x", sha256.Sum256([]byte(cpuQuote)))

	if currentHash != snapshot.CPUQuoteHash {
		bv.log.Warnf("LLM fast-verify: CPU quote hash mismatch for model %s, performing full re-attestation", modelID)
		return bv.AttestBackend(ctx, modelID, snapshot.AttestationURL)
	}

	if !strings.EqualFold(tlsFingerprint, snapshot.TLSFingerprint) {
		bv.log.Warnf("LLM fast-verify: TLS fingerprint mismatch for model %s (cached=%s, live=%s)", modelID, snapshot.TLSFingerprint, tlsFingerprint)
		return fmt.Errorf("LLM TLS certificate changed for model %s (possible MITM)", modelID)
	}

	bv.log.Debugf("LLM fast-verify: model %s verified", modelID)
	return nil
}

// GetStatus returns the attestation snapshot for a model, or nil if not attested.
func (bv *BackendVerifier) GetBackendStatus(modelID string) *BackendAttestationSnapshot {
	return bv.GetStatus(modelID)
}

func (bv *BackendVerifier) GetStatus(modelID string) *BackendAttestationSnapshot {
	bv.mu.RLock()
	defer bv.mu.RUnlock()

	snapshot, exists := bv.cache[modelID]
	if !exists {
		return nil
	}

	copied := *snapshot
	return &copied
}

// GetAllStatuses returns attestation snapshots for all cached models.
func (bv *BackendVerifier) GetAllStatuses() map[string]*BackendAttestationSnapshot {
	bv.mu.RLock()
	defer bv.mu.RUnlock()

	result := make(map[string]*BackendAttestationSnapshot, len(bv.cache))
	for k, v := range bv.cache {
		copied := *v
		result[k] = &copied
	}
	return result
}

// PinnedHTTPClient returns an HTTP client whose TLS transport is pinned to the
// certificate fingerprint from the model's attestation snapshot.
func (bv *BackendVerifier) PinnedHTTPClient(modelID string) (*http.Client, error) {
	bv.mu.RLock()
	snapshot, exists := bv.cache[modelID]
	bv.mu.RUnlock()

	if !exists || snapshot.Status != StatusPassed {
		return nil, fmt.Errorf("no valid attestation for model %s", modelID)
	}

	expectedFingerprint := snapshot.TLSFingerprint

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: true, //nolint:gosec // verified via attestation binding
				VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
					if len(rawCerts) == 0 {
						return fmt.Errorf("no peer certificate presented")
					}
					hash := sha256.Sum256(rawCerts[0])
					actual := hex.EncodeToString(hash[:])
					if !strings.EqualFold(actual, expectedFingerprint) {
						return fmt.Errorf("TLS cert pinning mismatch: expected %s, got %s", expectedFingerprint, actual)
					}
					return nil
				},
			},
		},
	}, nil
}

// VerifyCPUGPUBinding checks that the CPU and GPU attestation quotes are bound together.
// The second half (bytes 32-63, hex chars 64-127) of the CPU reportData should match
// the GPU attestation's reportData (which serves as the GPU nonce).
func VerifyCPUGPUBinding(cpuReportData string, gpuReportData string) error {
	cpuReportData = strings.ToLower(strings.TrimSpace(cpuReportData))
	gpuReportData = strings.ToLower(strings.TrimSpace(gpuReportData))

	// CPU reportData layout:
	//   chars 0-63  (bytes 0-31): TLS cert fingerprint
	//   chars 64+   (bytes 32+):  GPU attestation nonce
	const tlsFingerprintHexLen = 64

	if len(cpuReportData) <= tlsFingerprintHexLen {
		return fmt.Errorf("CPU reportData too short (%d hex chars) to contain GPU binding", len(cpuReportData))
	}

	gpuNonceFromCPU := cpuReportData[tlsFingerprintHexLen:]

	if gpuReportData == "" {
		return fmt.Errorf("GPU reportData is empty, cannot verify binding")
	}

	// Compare the GPU nonce embedded in CPU reportData against GPU's reportData.
	// Use the shorter of the two for comparison in case of trailing zeros.
	compareLen := len(gpuNonceFromCPU)
	if len(gpuReportData) < compareLen {
		compareLen = len(gpuReportData)
	}
	if compareLen == 0 {
		return fmt.Errorf("no data available for CPU-GPU binding comparison")
	}

	if gpuNonceFromCPU[:compareLen] != gpuReportData[:compareLen] {
		return fmt.Errorf("CPU-GPU binding mismatch: cpu_reportdata_suffix=%s, gpu_reportdata=%s",
			gpuNonceFromCPU[:compareLen], gpuReportData[:compareLen])
	}

	return nil
}

func (bv *BackendVerifier) fetchDockerCompose(ctx context.Context, attestationURL string) (string, error) {
	composeURL := attestationURL + "/docker-compose"
	bv.log.Infof("fetching docker-compose from %s", composeURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, composeURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for %s: %w", composeURL, err)
	}

	resp, err := bv.attestationClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch docker-compose from %s: %w", composeURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("docker-compose endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read docker-compose response: %w", err)
	}

	content := string(body)

	// The SecretVM :29343/docker-compose endpoint wraps the YAML in an HTML page.
	// Extract the raw content from inside <pre>...</pre> tags.
	content = extractPreContent(content)

	return content, nil
}

// extractPreContent extracts text between <pre> and </pre> tags,
// decodes HTML entities (e.g. &amp; -> &, &#34; -> "), and strips
// any trailing zero-width spaces inserted by the HTML renderer.
// The SecretVM attestation server serves docker-compose as an HTML
// page with the YAML inside <pre> tags, HTML-escaped.
func extractPreContent(rawHTML string) string {
	lower := strings.ToLower(rawHTML)
	start := strings.Index(lower, "<pre>")
	if start == -1 {
		return rawHTML
	}
	start += len("<pre>")
	end := strings.Index(lower[start:], "</pre>")
	if end == -1 {
		return rawHTML
	}
	content := html.UnescapeString(rawHTML[start : start+end])
	content = strings.TrimRight(content, "\u200b")
	return content
}

func (bv *BackendVerifier) storeFailure(modelID, attestationURL, errMsg string) {
	bv.mu.Lock()
	defer bv.mu.Unlock()

	bv.cache[modelID] = &BackendAttestationSnapshot{
		ModelID:        modelID,
		AttestationURL: attestationURL,
		Status:         StatusFailed,
		Error:          errMsg,
		VerifiedAt:     time.Now(),
	}
}
