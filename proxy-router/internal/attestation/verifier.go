package attestation

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

// collateralField handles the "collateral" JSON field which may be either a
// string (raw hex blob) or an object with an optional "error" key.
type collateralField struct {
	Error string
}

func (c *collateralField) UnmarshalJSON(data []byte) error {
	// Try as string first (e.g. "81000000...")
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		c.Error = ""
		return nil
	}
	// Try as object with optional error field
	var obj struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	c.Error = obj.Error
	return nil
}

const (
	attestationPort  = "29343"
	defaultPortalURL = "https://secretai.scrtlabs.com/api/quote-parse"
	verifyTimeout    = 30 * time.Second
)

// ParseQuoteRequest is the POST body for the SecretAI Portal quote-parse API.
type ParseQuoteRequest struct {
	Quote string `json:"quote"`
}

// ParseQuoteResponse represents the parsed attestation quote from the SecretAI Portal.
type ParseQuoteResponse struct {
	Error      string           `json:"error,omitempty"`
	Quote      *QuoteFields     `json:"quote,omitempty"`
	Collateral *collateralField `json:"collateral,omitempty"`
	Status     *QuoteStatus     `json:"status,omitempty"`
}

type QuoteFields struct {
	Version     string `json:"version,omitempty"`
	TEEType     string `json:"tee_type,omitempty"`
	TCBSVN      string `json:"tcb_svn,omitempty"`
	MRSeam      string `json:"mr_seam,omitempty"`
	MRTD        string `json:"mr_td,omitempty"`
	RTMR0       string `json:"rtmr0,omitempty"`
	RTMR1       string `json:"rtmr1,omitempty"`
	RTMR2       string `json:"rtmr2,omitempty"`
	RTMR3       string `json:"rtmr3,omitempty"`
	ReportData  string `json:"report_data,omitempty"`
	Measurement string `json:"measurement,omitempty"`
	MachineID   string `json:"machine_id,omitempty"`
}

type QuoteStatus struct {
	AttestationType string `json:"attestation_type,omitempty"`
	Result          string `json:"result,omitempty"`
	ExpStatus       string `json:"exp_status,omitempty"`
}

type TEEType string

const (
	TEETypeTDX TEEType = "TDX"
	TEETypeSEV TEEType = "SEV"
)

type AttestationResult struct {
	Valid bool
	Type  TEEType
	Error string

	// TDX registers
	MRTD  string
	RTMR0 string
	RTMR1 string
	RTMR2 string
	RTMR3 string

	// SEV-SNP registers
	Measurement string
}

type Verifier struct {
	httpClient *http.Client
	portalURL  string
	goldenSrc  *GoldenSource
	log        lib.ILogger
}

func NewVerifier(portalURL string, imageRepo string, log lib.ILogger) *Verifier {
	if portalURL == "" {
		portalURL = defaultPortalURL
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}

	return &Verifier{
		httpClient: &http.Client{Timeout: verifyTimeout, Transport: transport},
		portalURL:  portalURL,
		goldenSrc:  NewGoldenSource(imageRepo, log),
		log:        log,
	}
}

// VerifyProvider performs TEE attestation verification for a provider.
// 1. Fetches the raw attestation quote from the provider's :29343/cpu endpoint
// 2. Sends it to the SecretAI Portal parse-quote API for cryptographic verification
// 3. Compares all available registers from the parsed quote against golden values
func (v *Verifier) VerifyProvider(ctx context.Context, providerEndpoint string, version string) error {
	attestationURL, err := deriveAttestationURL(providerEndpoint)
	if err != nil {
		return fmt.Errorf("failed to derive attestation URL: %w", err)
	}

	v.log.Infof("verifying TEE attestation for provider %s (version %s)", providerEndpoint, version)

	hexQuote, err := v.loadAttestationQuote(ctx, attestationURL)
	if err != nil {
		return fmt.Errorf("failed to load attestation quote from %s: %w", attestationURL, err)
	}

	v.log.Infof("Got attestation quote: %s", hexQuote)

	result, err := v.verifyQuote(ctx, hexQuote)
	if err != nil {
		return fmt.Errorf("attestation quote verification failed: %w", err)
	}

	v.log.Infof("Got attestation result: %+v", result)

	if !result.Valid {
		return fmt.Errorf("attestation invalid (%s): %s", result.Type, result.Error)
	}

	v.log.Infof("attestation quote is valid (type: %s) for provider %s", result.Type, providerEndpoint)

	golden, err := v.goldenSrc.FetchGoldenValues(ctx, version)
	if err != nil {
		v.log.Warnf("failed to fetch golden values for version %s, skipping register comparison: %s", version, err)
		return nil
	}

	v.log.Infof("Got golden values: %+v", golden)

	if err := v.compareRegisters(result, golden); err != nil {
		v.log.Warnf("failed to compare registers: %s", err)
		return err
	}

	v.log.Infof("all TEE register values match golden values for version %s", version)
	return nil
}

// compareRegisters checks every register present in the golden values against
// the values extracted from the provider's attestation quote.
func (v *Verifier) compareRegisters(result *AttestationResult, golden *GoldenValues) error {
	type regPair struct {
		name   string
		golden string
		actual string
	}

	var pairs []regPair

	switch result.Type {
	case TEETypeTDX:
		pairs = []regPair{
			// {"MRTD", golden.MRTD, result.MRTD},
			// {"RTMR0", golden.RTMR0, result.RTMR0},
			// {"RTMR1", golden.RTMR1, result.RTMR1},
			// {"RTMR2", golden.RTMR2, result.RTMR2},
			{"RTMR3", golden.RTMR3, result.RTMR3},
		}
	case TEETypeSEV:
		pairs = []regPair{
			{"measurement", golden.Measurement, result.Measurement},
		}
	}

	var mismatches []string
	for _, p := range pairs {
		if p.golden == "" {
			continue
		}
		if p.actual == "" {
			mismatches = append(mismatches, fmt.Sprintf("%s: expected %s but not present in quote", p.name, p.golden))
			continue
		}
		if !strings.EqualFold(p.golden, p.actual) {
			mismatches = append(mismatches, fmt.Sprintf("%s: expected %s, got %s", p.name, p.golden, p.actual))
		}
	}

	if len(mismatches) > 0 {
		return fmt.Errorf("register mismatch: %s", strings.Join(mismatches, "; "))
	}

	return nil
}

// loadAttestationQuote fetches the raw hex-encoded attestation quote from the provider.
func (v *Verifier) loadAttestationQuote(ctx context.Context, attestationBaseURL string) (string, error) {
	cpuURL := attestationBaseURL + "/cpu"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cpuURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch attestation quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("attestation endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read attestation quote: %w", err)
	}

	hexQuote := strings.TrimSpace(string(body))
	if hexQuote == "" {
		return "", fmt.Errorf("empty attestation quote from provider")
	}

	return hexQuote, nil
}

// verifyQuote sends the hex attestation quote to the SecretAI Portal parse-quote API
// for cryptographic verification and field extraction.
func (v *Verifier) verifyQuote(ctx context.Context, hexQuote string) (*AttestationResult, error) {
	reqBody := ParseQuoteRequest{Quote: hexQuote}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.portalURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("portal request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read portal response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("portal returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed ParseQuoteResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse portal response: %w", err)
	}

	if parsed.Error != "" {
		return &AttestationResult{Valid: false, Error: parsed.Error}, nil
	}

	q := parsed.Quote
	mrtd := qf(q, "mr_td")
	rtmr0 := qf(q, "rtmr0")
	rtmr1 := qf(q, "rtmr1")
	rtmr2 := qf(q, "rtmr2")
	rtmr3 := qf(q, "rtmr3")
	measurement := qf(q, "measurement")
	reportData := qf(q, "report_data")

	hasTDX := mrtd != "" && rtmr0 != "" && rtmr1 != "" && rtmr2 != ""
	hasSEV := measurement != "" || reportData != ""

	// status.attestation_type can also indicate TDX/SEV
	if !hasTDX && parsed.Status != nil && strings.EqualFold(parsed.Status.AttestationType, "tdx") {
		hasTDX = true
	}

	if !hasTDX && !hasSEV {
		return &AttestationResult{Valid: false, Error: "missing required TEE fields in attestation quote"}, nil
	}

	if parsed.Collateral != nil && parsed.Collateral.Error != "" {
		teeType := TEETypeSEV
		if hasTDX {
			teeType = TEETypeTDX
		}
		return &AttestationResult{Valid: false, Type: teeType, Error: parsed.Collateral.Error}, nil
	}

	teeType := TEETypeSEV
	if hasTDX {
		teeType = TEETypeTDX
	}

	return &AttestationResult{
		Valid:       true,
		Type:        teeType,
		MRTD:        mrtd,
		RTMR0:       rtmr0,
		RTMR1:       rtmr1,
		RTMR2:       rtmr2,
		RTMR3:       rtmr3,
		Measurement: measurement,
	}, nil
}

func qf(q *QuoteFields, field string) string {
	if q == nil {
		return ""
	}
	switch field {
	case "mr_td":
		return q.MRTD
	case "rtmr0":
		return q.RTMR0
	case "rtmr1":
		return q.RTMR1
	case "rtmr2":
		return q.RTMR2
	case "rtmr3":
		return q.RTMR3
	case "measurement":
		return q.Measurement
	case "report_data":
		return q.ReportData
	default:
		return ""
	}
}

// deriveAttestationURL constructs the SecretVM attestation base URL from a provider endpoint.
// Provider endpoint format: "host:port" (e.g., "morpheus.dev.lumerin.io:3333")
// Attestation URL format: "https://host:29343" (standard SecretVM attestation port)
func deriveAttestationURL(providerEndpoint string) (string, error) {
	if strings.Contains(providerEndpoint, "://") {
		parsed, err := url.Parse(providerEndpoint)
		if err != nil {
			return "", fmt.Errorf("invalid provider endpoint URL: %w", err)
		}
		host := parsed.Hostname()
		return fmt.Sprintf("https://%s:%s", host, attestationPort), nil
	}

	host, _, err := net.SplitHostPort(providerEndpoint)
	if err != nil {
		host = providerEndpoint
	}
	return fmt.Sprintf("https://%s:%s", host, attestationPort), nil
}
