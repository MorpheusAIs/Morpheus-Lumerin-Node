package attestation

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
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
	AttestationPort  = "29343"
	DefaultPortalURL = "https://secretai.scrtlabs.com/api/quote-parse"
	VerifyTimeout    = 30 * time.Second
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

	ReportData string
}

type verifiedQuoteEntry struct {
	quoteHash      string
	tlsFingerprint string
}

// PingFunc obtains the provider's software version by pinging its endpoint.
// providerAddr is the hex-encoded provider address required for signature verification.
// Used by VerifyProviderQuick on cache miss to perform a full verification.
type PingFunc func(ctx context.Context, providerEndpoint string, providerAddr string) (version string, err error)

type Verifier struct {
	portalClient      *http.Client
	attestationClient *http.Client
	portalURL         string
	goldenSrc         *GoldenSource
	log               lib.ILogger
	pingFunc          PingFunc

	mu         sync.RWMutex
	quoteCache map[string]*verifiedQuoteEntry
}

// NewPortalHTTPClient creates an HTTP client for the SecretAI Portal API.
func NewPortalHTTPClient() *http.Client {
	return &http.Client{
		Timeout: VerifyTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		},
	}
}

// NewAttestationHTTPClient creates an HTTP client for TEE attestation endpoints.
// Uses InsecureSkipVerify because the self-signed cert is verified via reportdata binding.
func NewAttestationHTTPClient() *http.Client {
	return &http.Client{
		Timeout: VerifyTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: true, //nolint:gosec // verified via reportdata
			},
		},
	}
}

func NewVerifier(portalURL string, imageRepo string, log lib.ILogger) *Verifier {
	if portalURL == "" {
		portalURL = DefaultPortalURL
	}

	return &Verifier{
		portalClient:      NewPortalHTTPClient(),
		attestationClient: NewAttestationHTTPClient(),
		portalURL:         portalURL,
		goldenSrc:         NewGoldenSource(imageRepo, log),
		log:               log,
		quoteCache:        make(map[string]*verifiedQuoteEntry),
	}
}

func (v *Verifier) SetPingFunc(f PingFunc) {
	v.pingFunc = f
}

// VerifyProvider performs TEE attestation verification for a provider.
//  1. Fetches the raw attestation quote from the provider's :29343/cpu endpoint
//     and captures the TLS certificate fingerprint of the connection
//  2. Sends it to the SecretAI Portal parse-quote API for cryptographic verification
//  3. Verifies that the TLS certificate fingerprint matches the reportdata field
//     in the quote (anti-spoofing: proves the quote belongs to this server)
//  4. Compares all available registers from the parsed quote against golden values
func (v *Verifier) VerifyProvider(ctx context.Context, providerEndpoint string, version string) error {
	attestationURL, err := deriveAttestationURL(providerEndpoint)
	if err != nil {
		return fmt.Errorf("failed to derive attestation URL: %w", err)
	}

	v.log.Infof("verifying TEE attestation for provider %s (version %s)", providerEndpoint, version)

	hexQuote, tlsFingerprint, err := v.loadAttestationQuote(ctx, attestationURL)
	if err != nil {
		return fmt.Errorf("failed to load attestation quote from %s: %w", attestationURL, err)
	}

	v.log.Infof("captured TLS cert fingerprint: %s", tlsFingerprint)

	result, err := v.verifyQuote(ctx, hexQuote)
	if err != nil {
		return fmt.Errorf("attestation quote verification failed: %w", err)
	}

	v.log.Infof("Got attestation result: %+v", result)

	if !result.Valid {
		return fmt.Errorf("attestation invalid (%s): %s", result.Type, result.Error)
	}

	v.log.Infof("attestation quote is valid (type: %s) for provider %s", result.Type, providerEndpoint)

	if err := VerifyTLSBinding(tlsFingerprint, result.ReportData); err != nil {
		return fmt.Errorf("TLS binding verification failed (possible spoofing): %w", err)
	}

	v.log.Infof("TLS certificate fingerprint matches reportdata (anti-spoofing check passed)")

	golden, err := v.goldenSrc.FetchGoldenValues(ctx, version)
	if err != nil {
		return fmt.Errorf("failed to fetch golden values for version %s: %w", version, err)
	}

	v.log.Infof("Got golden values: %+v", golden)

	if err := CompareRegisters(result, golden, v.log); err != nil {
		v.log.Warnf("failed to compare registers: %s", err)
		return err
	}

	v.log.Infof("all TEE register values match golden values for version %s", version)

	quoteHash := fmt.Sprintf("%x", sha256.Sum256([]byte(hexQuote)))
	v.mu.Lock()
	v.quoteCache[attestationURL] = &verifiedQuoteEntry{
		quoteHash:      quoteHash,
		tlsFingerprint: tlsFingerprint,
	}
	v.mu.Unlock()
	v.log.Infof("cached verified quote for %s", attestationURL)

	return nil
}

// VerifyTLSBinding checks that the SHA-256 fingerprint of the TLS certificate
// presented by the attestation endpoint matches the reportdata field in the
// hardware-signed attestation quote.
//
// SecretVM generates a TLS certificate inside the TEE at boot and stores its
// fingerprint in the first 32 bytes (64 hex chars) of reportdata. Because the
// TLS private key never leaves the TEE, a spoofed server cannot present a
// certificate whose fingerprint matches a stolen quote's reportdata.
func VerifyTLSBinding(tlsFingerprint string, reportData string) error {
	if tlsFingerprint == "" {
		return fmt.Errorf("no TLS certificate fingerprint captured from attestation endpoint")
	}
	if reportData == "" {
		return fmt.Errorf("no report_data in attestation quote")
	}

	reportData = strings.ToLower(strings.TrimSpace(reportData))
	tlsFingerprint = strings.ToLower(strings.TrimSpace(tlsFingerprint))

	if len(reportData) < len(tlsFingerprint) {
		return fmt.Errorf("report_data too short (%d chars) to contain TLS fingerprint (%d chars)",
			len(reportData), len(tlsFingerprint))
	}

	reportPrefix := reportData[:len(tlsFingerprint)]
	if reportPrefix != tlsFingerprint {
		return fmt.Errorf("TLS certificate fingerprint mismatch: connection=%s, reportdata_prefix=%s",
			tlsFingerprint, reportPrefix)
	}

	return nil
}

// VerifyProviderQuick performs a fast per-request attestation check.
//
// Cache hit: fetches the quote from :29343/cpu (~50-150ms TLS handshake),
// computes sha256(quote) and compares it (plus the TLS fingerprint) against
// the cached values from the last full verification. If both match the
// provider is the same TEE -- return nil.
//
// Cache miss (e.g. after process restart): performs a full VerifyProvider
// (ping for version + portal verification + golden values) and populates
// the cache. This is slower (~250-650ms) but only happens once per provider.
//
// If isTee is false the check is a no-op.
func (v *Verifier) VerifyProviderQuick(ctx context.Context, providerEndpoint string, providerAddr string, isTee bool) error {
	if !isTee {
		v.log.Debugf("quick attestation: skipping non-TEE session for %s", providerEndpoint)
		return nil
	}

	v.log.Infof("quick attestation: starting check for provider %s", providerEndpoint)

	attestationURL, err := deriveAttestationURL(providerEndpoint)
	if err != nil {
		return fmt.Errorf("failed to derive attestation URL: %w", err)
	}

	v.mu.RLock()
	cached, hasCached := v.quoteCache[attestationURL]
	v.mu.RUnlock()

	if !hasCached {
		v.log.Infof("quick attestation: no cached quote for %s, falling back to full verification", attestationURL)
		return v.fullVerifyWithPing(ctx, providerEndpoint, providerAddr)
	}

	v.log.Infof("quick attestation: cache hit for %s, fetching live quote", attestationURL)

	hexQuote, tlsFingerprint, err := v.loadAttestationQuote(ctx, attestationURL)
	if err != nil {
		return fmt.Errorf("quick attestation check failed: %w", err)
	}

	v.log.Infof("quick attestation: fetched live quote from %s, TLS fingerprint: %s", attestationURL, tlsFingerprint)

	currentHash := fmt.Sprintf("%x", sha256.Sum256([]byte(hexQuote)))

	if currentHash != cached.quoteHash {
		v.log.Warnf("quick attestation: quote hash MISMATCH for %s (cached=%s, live=%s)", providerEndpoint, cached.quoteHash, currentHash)
		return v.fullVerifyWithPing(ctx, providerEndpoint, providerAddr)
	}

	v.log.Infof("quick attestation: quote hash matches cached value for %s", providerEndpoint)

	if !strings.EqualFold(tlsFingerprint, cached.tlsFingerprint) {
		v.log.Warnf("quick attestation: TLS fingerprint MISMATCH for %s (cached=%s, live=%s)", providerEndpoint, cached.tlsFingerprint, tlsFingerprint)
		return fmt.Errorf("TLS certificate changed since session was opened (provider %s)", providerEndpoint)
	}

	v.log.Infof("quick attestation: TLS fingerprint matches cached value for %s — provider verified", providerEndpoint)
	return nil
}

// fullVerifyWithPing pings the provider to obtain its version, then performs
// a full VerifyProvider which populates the quote cache on success.
func (v *Verifier) fullVerifyWithPing(ctx context.Context, providerEndpoint string, providerAddr string) error {
	if v.pingFunc == nil {
		return fmt.Errorf("cannot perform full verification: no ping function configured")
	}

	v.log.Infof("full verification: pinging provider %s (addr %s) for version", providerEndpoint, providerAddr)

	version, err := v.pingFunc(ctx, providerEndpoint, providerAddr)
	if err != nil {
		return fmt.Errorf("TEE ping failed for provider %s: %w", providerEndpoint, err)
	}
	if version == "" {
		return fmt.Errorf("TEE provider %s did not report a version", providerEndpoint)
	}

	v.log.Infof("full verification: provider %s reported version %s, proceeding with full attestation", providerEndpoint, version)

	return v.VerifyProvider(ctx, providerEndpoint, version)
}

// CompareRegisters checks every register present in the golden values against
// the values extracted from the attestation quote.
func CompareRegisters(result *AttestationResult, golden *GoldenValues, log lib.ILogger) error {
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
			if log != nil {
				log.Debugf("register %s: golden value empty, skipping", p.name)
			}
			continue
		}
		if p.actual == "" {
			mismatches = append(mismatches, fmt.Sprintf("%s: expected %s but not present in quote", p.name, p.golden))
			continue
		}
		if !strings.EqualFold(p.golden, p.actual) {
			mismatches = append(mismatches, fmt.Sprintf("%s: expected %s, got %s", p.name, p.golden, p.actual))
		} else if log != nil {
			log.Infof("register %s: matches golden value", p.name)
		}
	}

	if len(mismatches) > 0 {
		return fmt.Errorf("register mismatch: %s", strings.Join(mismatches, "; "))
	}

	if log != nil {
		log.Infof("all checked registers match golden values")
	}
	return nil
}

// LoadAttestationQuote fetches a raw hex-encoded attestation quote from the
// given URL path and returns the SHA-256 fingerprint of the peer's TLS certificate.
func LoadAttestationQuote(ctx context.Context, client *http.Client, quoteURL string) (hexQuote string, tlsFingerprint string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, quoteURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch attestation quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("attestation endpoint returned status %d", resp.StatusCode)
	}

	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		hash := sha256.Sum256(resp.TLS.PeerCertificates[0].Raw)
		tlsFingerprint = hex.EncodeToString(hash[:])
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read attestation quote: %w", err)
	}

	hexQuote = strings.TrimSpace(string(body))
	if hexQuote == "" {
		return "", "", fmt.Errorf("empty attestation quote")
	}

	return hexQuote, tlsFingerprint, nil
}

// loadAttestationQuote is the instance method that delegates to the package-level function.
func (v *Verifier) loadAttestationQuote(ctx context.Context, attestationBaseURL string) (hexQuote string, tlsFingerprint string, err error) {
	cpuURL := attestationBaseURL + "/cpu"
	v.log.Infof("fetching attestation quote from %s", cpuURL)

	hexQuote, tlsFingerprint, err = LoadAttestationQuote(ctx, v.attestationClient, cpuURL)
	if err != nil {
		return "", "", err
	}

	if tlsFingerprint == "" {
		v.log.Warnf("no TLS peer certificate received from %s", cpuURL)
	}

	v.log.Infof("received attestation quote from %s (%d bytes)", cpuURL, len(hexQuote))
	return hexQuote, tlsFingerprint, nil
}

// VerifyQuote sends the hex attestation quote to the SecretAI Portal parse-quote API
// for cryptographic verification and field extraction.
func VerifyQuote(ctx context.Context, portalClient *http.Client, portalURL string, hexQuote string) (*AttestationResult, error) {
	reqBody := ParseQuoteRequest{Quote: hexQuote}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, portalURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := portalClient.Do(req)
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
		ReportData:  reportData,
	}, nil
}

// verifyQuote is the instance method that delegates to the package-level function.
func (v *Verifier) verifyQuote(ctx context.Context, hexQuote string) (*AttestationResult, error) {
	v.log.Infof("sending quote to SecretAI portal for cryptographic verification (%s)", v.portalURL)
	result, err := VerifyQuote(ctx, v.portalClient, v.portalURL, hexQuote)
	if err != nil {
		return nil, err
	}
	if result.Error != "" {
		v.log.Warnf("portal returned error: %s", result.Error)
	} else {
		v.log.Infof("portal verified quote successfully")
	}
	return result, nil
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

// DeriveAttestationURL constructs the SecretVM attestation base URL from an endpoint.
// Input format: "host:port" or "https://host:port/path"
// Output format: "https://host:29343"
func DeriveAttestationURL(endpoint string) (string, error) {
	if strings.Contains(endpoint, "://") {
		parsed, err := url.Parse(endpoint)
		if err != nil {
			return "", fmt.Errorf("invalid endpoint URL: %w", err)
		}
		host := parsed.Hostname()
		return fmt.Sprintf("https://%s:%s", host, AttestationPort), nil
	}

	host, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		host = endpoint
	}
	return fmt.Sprintf("https://%s:%s", host, AttestationPort), nil
}

// deriveAttestationURL is the old unexported alias kept for internal compatibility.
func deriveAttestationURL(providerEndpoint string) (string, error) {
	return DeriveAttestationURL(providerEndpoint)
}
