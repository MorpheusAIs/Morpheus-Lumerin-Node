package attestation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

const (
	nrasBaseURL       = "https://nras.attestation.nvidia.com"
	nrasGPUAttestPath = "/v4/attest/gpu"
)

// GPUEvidence represents a single GPU's attestation evidence from the SecretVM /gpu endpoint.
type GPUEvidence struct {
	Evidence    string `json:"evidence"`
	Certificate string `json:"certificate"`
}

// GPUAttestationData is the JSON structure returned by the SecretVM /gpu endpoint.
type GPUAttestationData struct {
	Nonce        string        `json:"nonce"`
	Arch         string        `json:"arch"`
	EvidenceList []GPUEvidence `json:"evidence_list"`
}

// NRASResult holds the parsed response from NRAS.
type NRASResult struct {
	OverallToken string
	GPUTokens    map[string]string
}

// NRASVerifier verifies GPU attestation evidence via NVIDIA Remote Attestation Service.
// It sends the GPU evidence collected from the SecretVM /gpu endpoint to NRAS,
// which validates the NVIDIA certificate chain and evidence signature,
// and returns signed Entity Attestation Token (EAT) JWTs.
type NRASVerifier struct {
	client  *http.Client
	baseURL string
	log     lib.ILogger
}

func NewNRASVerifier(log lib.ILogger) *NRASVerifier {
	return &NRASVerifier{
		client:  NewPortalHTTPClient(),
		baseURL: nrasBaseURL,
		log:     log,
	}
}

// VerifyGPU sends GPU attestation evidence to NRAS for verification.
// The gpuData is the raw JSON response from the SecretVM /gpu endpoint.
// Returns the overall JWT token and per-GPU tokens on success.
func (nv *NRASVerifier) VerifyGPU(ctx context.Context, gpuData *GPUAttestationData) (*NRASResult, error) {
	if len(gpuData.EvidenceList) == 0 {
		return nil, fmt.Errorf("GPU attestation data contains no evidence entries")
	}

	// Send the GPU attestation data as-is to NRAS (same as secretvm-verify JS library).
	// GPUAttestationData JSON tags match what NRAS expects: nonce, arch, evidence_list.
	body, err := json.Marshal(gpuData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal NRAS request: %w", err)
	}

	attestURL := nv.baseURL + nrasGPUAttestPath
	noncePreview := gpuData.Nonce
	if len(noncePreview) > 16 {
		noncePreview = noncePreview[:16] + "..."
	}
	nv.log.Infof("NRAS: sending GPU evidence to %s (arch=%s, nonce=%s, gpus=%d)",
		attestURL, gpuData.Arch, noncePreview, len(gpuData.EvidenceList))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, attestURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create NRAS request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := nv.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("NRAS request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read NRAS response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NRAS returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return parseNRASResponse(respBody)
}

// parseNRASResponse parses the NRAS response format:
// [ ["JWT", "<overall-token>"], { "GPU-0": "<token>", ... } ]
func parseNRASResponse(body []byte) (*NRASResult, error) {
	var raw []json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse NRAS response as array: %w", err)
	}

	if len(raw) < 2 {
		return nil, fmt.Errorf("NRAS response has fewer than 2 elements (got %d)", len(raw))
	}

	// First element: ["JWT", "<overall-token>"]
	var overallPair []string
	if err := json.Unmarshal(raw[0], &overallPair); err != nil {
		return nil, fmt.Errorf("failed to parse NRAS overall token pair: %w", err)
	}
	if len(overallPair) < 2 {
		return nil, fmt.Errorf("NRAS overall token pair has fewer than 2 elements")
	}

	// Second element: { "GPU-0": "<token>", ... }
	var gpuTokens map[string]string
	if err := json.Unmarshal(raw[1], &gpuTokens); err != nil {
		return nil, fmt.Errorf("failed to parse NRAS GPU tokens: %w", err)
	}

	return &NRASResult{
		OverallToken: overallPair[1],
		GPUTokens:    gpuTokens,
	}, nil
}

// ParseGPUAttestationData parses the raw JSON response from the SecretVM /gpu endpoint.
func ParseGPUAttestationData(rawJSON string) (*GPUAttestationData, error) {
	var data GPUAttestationData
	if err := json.Unmarshal([]byte(rawJSON), &data); err != nil {
		return nil, fmt.Errorf("failed to parse GPU attestation JSON: %w", err)
	}
	if data.Nonce == "" {
		return nil, fmt.Errorf("GPU attestation data missing nonce field")
	}
	if data.Arch == "" {
		return nil, fmt.Errorf("GPU attestation data missing arch field")
	}
	if len(data.EvidenceList) == 0 {
		return nil, fmt.Errorf("GPU attestation data has empty evidence_list")
	}
	return &data, nil
}
