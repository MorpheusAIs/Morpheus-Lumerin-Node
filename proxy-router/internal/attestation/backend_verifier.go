package attestation

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
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
	TLSBinding        TLSCertBinding           `json:"-"`
	TLSBindingKind    TLSBindingKind           `json:"tlsBindingKind,omitempty"`
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
	sevRegistry       *SevArtifactRegistry
	log               lib.ILogger

	mu    sync.RWMutex
	cache map[string]*BackendAttestationSnapshot
}

func NewBackendVerifier(portalURL string, goldenSource BackendGoldenSource, registry *ArtifactRegistry, sevRegistry *SevArtifactRegistry, log lib.ILogger) *BackendVerifier {
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
		sevRegistry:       sevRegistry,
		log:               log,
		cache:             make(map[string]*BackendAttestationSnapshot),
	}
}

// backendAttestationPorts lists the candidate ports probed when resolving the
// Phase 2 backend attestation endpoint: 21434 first (host-net Caddy topology
// used by current SecretAI backends, e.g. jedi/rytn, where attest-rest is
// loopback-only and proxied), then 29343 (attest-rest bound directly, standard
// SecretVMs). Upstream secretvm-verify v0.12.0 probes 29343 first; the order
// is flipped here so the common backend topology resolves without waiting out
// a probe timeout.
var backendAttestationPorts = []string{BackendAttestationPort, AttestationPort}

const attestationProbeTimeout = 5 * time.Second

// ResolveAttestationURL derives the Phase 2 backend attestation base URL from
// a model's apiUrl by probing GET /cpu on each candidate port and returning
// the first endpoint that answers. If none answers, the primary
// (https://<host>:21434) URL is returned so the subsequent attestation
// surfaces a clear fetch error against it.
func (bv *BackendVerifier) ResolveAttestationURL(ctx context.Context, endpoint string) (string, error) {
	var first string
	for _, port := range backendAttestationPorts {
		candidate, err := DeriveAttestationURLWithPort(endpoint, port)
		if err != nil {
			return "", err
		}
		if first == "" {
			first = candidate
		}
		if bv.probeCPU(ctx, candidate) {
			bv.log.Infof("backend attestation: resolved attestation endpoint %s for %s", candidate, endpoint)
			return candidate, nil
		}
	}
	bv.log.Warnf("backend attestation: no attestation port answered /cpu for %s, falling back to %s", endpoint, first)
	return first, nil
}

func (bv *BackendVerifier) probeCPU(ctx context.Context, baseURL string) bool {
	probeCtx, cancel := context.WithTimeout(ctx, attestationProbeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, baseURL+"/cpu", nil)
	if err != nil {
		return false
	}
	resp, err := bv.attestationClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode == http.StatusOK
}

// AttestBackend performs full CPU+GPU attestation of a backend LLM TEE endpoint.
// On success the result is cached for fast-verify. On failure the snapshot is
// stored with StatusFailed so the health endpoint can report it.
func (bv *BackendVerifier) AttestBackend(ctx context.Context, modelID string, attestationURL string) error {
	bv.log.Infof("backend attestation: starting full verification for model %s at %s", modelID, attestationURL)

	// 1. Fetch CPU quote
	cpuURL := attestationURL + "/cpu"
	cpuQuote, tlsBinding, err := LoadAttestationQuote(ctx, bv.attestationClient, cpuURL)
	if err != nil {
		bv.storeFailure(modelID, attestationURL, fmt.Sprintf("CPU quote fetch failed: %s", err))
		return fmt.Errorf("failed to load CPU attestation quote from %s: %w", cpuURL, err)
	}
	bv.log.Infof("backend attestation: fetched CPU quote from %s, TLS binding: %s", cpuURL, tlsBinding)

	if tlsBinding.IsZero() {
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

	// 3. Verify TLS binding (first half of reportData = SPKI or full-cert digest)
	bindingKind, err := VerifyTLSBinding(tlsBinding, cpuResult.ReportData)
	if err != nil {
		bv.storeFailure(modelID, attestationURL, fmt.Sprintf("TLS binding failed: %s", err))
		return fmt.Errorf("CPU TLS binding verification failed: %w", err)
	}
	bv.log.Infof("backend attestation: TLS binding verified (%s digest) for model %s", bindingKind, modelID)

	// 3a. Workload verification (docker-compose vs attestation quote).
	dockerCompose, composeErr := bv.fetchDockerCompose(ctx, attestationURL)
	if composeErr != nil {
		bv.storeFailure(modelID, attestationURL, fmt.Sprintf("could not fetch docker-compose: %s", composeErr))
		return fmt.Errorf("workload verification failed for model %s: could not fetch docker-compose: %w", modelID, composeErr)
	}

	workloadResult := VerifyWorkload(bv.artifactRegistry, bv.sevRegistry, cpuQuote, dockerCompose, bv.log)
	switch workloadResult.Status {
	case WorkloadAuthentic:
		bv.log.Infof("backend attestation: workload verified for model %s (template=%s, version=%s, env=%s)", modelID, workloadResult.TemplateName, workloadResult.ArtifactsVer, workloadResult.Env)
	case WorkloadAuthenticMismatch:
		bv.storeFailure(modelID, attestationURL, "docker-compose does not match attestation (authentic VM but wrong workload)")
		return fmt.Errorf("workload verification failed for model %s: docker-compose does not match attestation", modelID)
	case WorkloadNotAuthentic:
		bv.storeFailure(modelID, attestationURL, "VM is not an authentic SecretVM (MRTD/RTMR values not in registry)")
		return fmt.Errorf("workload verification failed for model %s: not an authentic SecretVM", modelID)
	case ArtifactRegistryNotAvailable:
		bv.storeFailure(modelID, attestationURL, "artifact registry not available; cannot verify workload")
		return fmt.Errorf("workload verification unavailable for model %s: artifact registry not loaded", modelID)
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

	// 7. Verify GPU evidence via NVIDIA Remote Attestation Service (NRAS).
	// NRAS validates the NVIDIA-signed GPU evidence and that it was generated over
	// the submitted nonce.
	nrasResult, err := bv.nrasVerifier.VerifyGPU(ctx, gpuData)
	if err != nil {
		bv.storeFailure(modelID, attestationURL, fmt.Sprintf("NRAS GPU verification failed: %s", err))
		return fmt.Errorf("NRAS GPU verification failed for model %s: %w", modelID, err)
	}
	if !nrasResult.OverallResult {
		bv.storeFailure(modelID, attestationURL, "NRAS reported overall attestation result: failed")
		return fmt.Errorf("NRAS GPU attestation failed for model %s (x-nvidia-overall-att-result=false)", modelID)
	}
	// Bind the NRAS-validated evidence back to the CPU quote: the eat_nonce NRAS
	// confirmed the evidence was generated over must equal the nonce we submitted
	// (gpuData.Nonce), which step 6 already proved equals the Intel-signed
	// CPU-embedded nonce.
	submittedNonce := strings.ToLower(strings.TrimSpace(gpuData.Nonce))
	attestedNonce := strings.ToLower(strings.TrimSpace(nrasResult.EATNonce))
	if attestedNonce == "" || attestedNonce != submittedNonce {
		bv.storeFailure(modelID, attestationURL, fmt.Sprintf("NRAS eat_nonce mismatch: submitted=%s attested=%s", submittedNonce, attestedNonce))
		return fmt.Errorf("NRAS eat_nonce does not match CPU-bound nonce for model %s", modelID)
	}
	bv.log.Infof("backend attestation: NRAS verified GPU for model %s (%d GPU tokens, nonce bound)", modelID, len(nrasResult.GPUTokens))

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
		TLSBinding:     tlsBinding,
		TLSBindingKind: bindingKind,
		CPUReportData:  cpuResult.ReportData,
		GPUReportData:  gpuData.Nonce,
		TEEType:        cpuResult.Type,
		VerifiedAt:     now,
		Status:         StatusPassed,
	}

	snapshot.WorkloadStatus = string(workloadResult.Status)
	snapshot.VMTemplateName = workloadResult.TemplateName
	snapshot.ArtifactsVersion = workloadResult.ArtifactsVer

	bv.mu.Lock()
	bv.cache[modelID] = snapshot
	bv.mu.Unlock()

	bv.log.Infof("backend attestation: cached verified snapshot for model %s", modelID)
	return nil
}

// ReattestBackend re-runs full backend attestation for a model regardless of
// the current snapshot status, so a transient failure (e.g. a portal glitch
// during startup attestation) self-heals on the next health sweep and a
// stale pass is re-proven. endpoint is the model's apiUrl, used to resolve
// the attestation URL when no snapshot with a URL exists.
func (bv *BackendVerifier) ReattestBackend(ctx context.Context, modelID string, endpoint string) error {
	bv.mu.RLock()
	snapshot := bv.cache[modelID]
	bv.mu.RUnlock()

	attestationURL := ""
	if snapshot != nil {
		attestationURL = snapshot.AttestationURL
	}
	if attestationURL == "" {
		resolved, err := bv.ResolveAttestationURL(ctx, endpoint)
		if err != nil {
			return fmt.Errorf("cannot resolve attestation URL for model %s: %w", modelID, err)
		}
		attestationURL = resolved
	}

	return bv.AttestBackend(ctx, modelID, attestationURL)
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
	cpuQuote, tlsBinding, err := LoadAttestationQuote(ctx, bv.attestationClient, cpuURL)
	if err != nil {
		return fmt.Errorf("LLM fast-verify failed for model %s: %w", modelID, err)
	}

	currentHash := fmt.Sprintf("%x", sha256.Sum256([]byte(cpuQuote)))

	if currentHash != snapshot.CPUQuoteHash {
		bv.log.Warnf("LLM fast-verify: CPU quote hash mismatch for model %s, performing full re-attestation", modelID)
		return bv.AttestBackend(ctx, modelID, snapshot.AttestationURL)
	}

	if !strings.EqualFold(tlsBinding.Certificate, snapshot.TLSBinding.Certificate) {
		// A renewed certificate keeps the same SPKI when the private key stays
		// inside the TEE; re-attest so the binding is re-proven against the
		// quote instead of hard-failing.
		if tlsBinding.SPKI != "" && strings.EqualFold(tlsBinding.SPKI, snapshot.TLSBinding.SPKI) {
			bv.log.Warnf("LLM fast-verify: TLS certificate rotated (same SPKI) for model %s, performing full re-attestation", modelID)
			return bv.AttestBackend(ctx, modelID, snapshot.AttestationURL)
		}
		bv.log.Warnf("LLM fast-verify: TLS binding mismatch for model %s (cached=%s, live=%s)", modelID, snapshot.TLSBinding, tlsBinding)
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
// certificate binding from the model's attestation snapshot. A peer matching
// either the full-certificate digest or the SPKI digest is accepted, so a
// certificate renewal that keeps the TEE-resident key does not break pinning.
func (bv *BackendVerifier) PinnedHTTPClient(modelID string) (*http.Client, error) {
	bv.mu.RLock()
	snapshot, exists := bv.cache[modelID]
	bv.mu.RUnlock()

	if !exists || snapshot.Status != StatusPassed {
		return nil, fmt.Errorf("no valid attestation for model %s", modelID)
	}

	expected := snapshot.TLSBinding

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: true, //nolint:gosec // verified via attestation binding
				VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
					if len(rawCerts) == 0 {
						return fmt.Errorf("no peer certificate presented")
					}
					certHash := sha256.Sum256(rawCerts[0])
					actualCert := hex.EncodeToString(certHash[:])
					if strings.EqualFold(actualCert, expected.Certificate) {
						return nil
					}
					if cert, err := x509.ParseCertificate(rawCerts[0]); err == nil {
						spkiHash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
						if strings.EqualFold(hex.EncodeToString(spkiHash[:]), expected.SPKI) {
							return nil
						}
					}
					return fmt.Errorf("TLS cert pinning mismatch: expected %s, got cert=%s", expected, actualCert)
				},
			},
		},
	}, nil
}

// VerifyCPUGPUBinding checks that the CPU and GPU attestation quotes are bound together.
// The second half (bytes 32-63, hex chars 64-127) of the CPU reportData should match
// the GPU attestation's reportData (which serves as the GPU nonce).
func VerifyCPUGPUBinding(cpuReportData string, gpuReportData string) error {
	cpuReportData = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(cpuReportData), " ", ""))
	gpuReportData = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(gpuReportData), " ", ""))

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

	// Require an exact match. A prefix/shorter-length comparison would let a GPU
	// nonce that only matches the first N chars of the CPU-embedded nonce pass,
	// collapsing the binding's entropy and enabling nonce-collision/replay.
	if len(gpuNonceFromCPU) != len(gpuReportData) {
		return fmt.Errorf("CPU-GPU binding length mismatch: cpu_reportdata_suffix=%d chars, gpu_reportdata=%d chars",
			len(gpuNonceFromCPU), len(gpuReportData))
	}
	if subtle.ConstantTimeCompare([]byte(gpuNonceFromCPU), []byte(gpuReportData)) != 1 {
		return fmt.Errorf("CPU-GPU binding mismatch: cpu_reportdata_suffix=%s, gpu_reportdata=%s",
			gpuNonceFromCPU, gpuReportData)
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

	// Return the exact response bytes. Old attest-rest wraps the YAML in an
	// HTML page while newer attest-rest serves the raw file; workload
	// verification tries both interpretations (see composeCandidates).
	return string(body), nil
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
