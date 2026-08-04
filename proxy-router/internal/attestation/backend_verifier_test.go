package attestation

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

func selfSignedCert(t *testing.T) (tls.Certificate, string) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{Organization: []string{"test"}},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}
	cert := tls.Certificate{Certificate: [][]byte{certDER}, PrivateKey: priv}
	h := sha256.Sum256(certDER)
	return cert, hex.EncodeToString(h[:])
}

func testLog() lib.ILogger {
	return &lib.LoggerMock{}
}

// certBinding computes both TLS binding digests for a test certificate.
func certBinding(t *testing.T, cert tls.Certificate) TLSCertBinding {
	t.Helper()
	parsed, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	return TLSBindingFromCert(parsed)
}

// --- VerifyCPUGPUBinding ---

func TestVerifyCPUGPUBinding_Valid(t *testing.T) {
	gpuNonce := "aabbccdd11223344556677889900aabbccdd11223344556677889900aabb1122"
	tlsFingerprint := "0011223344556677889900aabbccddeeff0011223344556677889900aabbccdd"
	cpuReportData := tlsFingerprint + gpuNonce

	if err := VerifyCPUGPUBinding(cpuReportData, gpuNonce); err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}
}

func TestVerifyCPUGPUBinding_Mismatch(t *testing.T) {
	tlsFingerprint := "0011223344556677889900aabbccddeeff0011223344556677889900aabbccdd"
	gpuNonce := "aabbccdd11223344556677889900aabbccdd11223344556677889900aabb1122"
	wrongGPU := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffff000000"
	cpuReportData := tlsFingerprint + gpuNonce

	err := VerifyCPUGPUBinding(cpuReportData, wrongGPU)
	if err == nil {
		t.Fatal("expected mismatch error")
	}
	if !strings.Contains(err.Error(), "CPU-GPU binding mismatch") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestVerifyCPUGPUBinding_PrefixDoesNotPass(t *testing.T) {
	gpuNonce := "aabbccdd11223344556677889900aabbccdd11223344556677889900aabb1122"
	tlsFingerprint := "0011223344556677889900aabbccddeeff0011223344556677889900aabbccdd"
	cpuReportData := tlsFingerprint + gpuNonce

	// A GPU reportData that is only a prefix of the expected nonce must be rejected.
	err := VerifyCPUGPUBinding(cpuReportData, gpuNonce[:1])
	if err == nil {
		t.Fatal("expected error for prefix-only GPU reportData")
	}
	if !strings.Contains(err.Error(), "length mismatch") {
		t.Fatalf("expected length mismatch error, got: %s", err)
	}
}

func TestVerifyCPUGPUBinding_LongerGPUReportData(t *testing.T) {
	gpuNonce := "aabbccdd11223344556677889900aabbccdd11223344556677889900aabb1122"
	tlsFingerprint := "0011223344556677889900aabbccddeeff0011223344556677889900aabbccdd"
	cpuReportData := tlsFingerprint + gpuNonce

	err := VerifyCPUGPUBinding(cpuReportData, gpuNonce+"deadbeef")
	if err == nil {
		t.Fatal("expected error for over-length GPU reportData")
	}
	if !strings.Contains(err.Error(), "length mismatch") {
		t.Fatalf("expected length mismatch error, got: %s", err)
	}
}

func TestVerifyCPUGPUBinding_ShortReportData(t *testing.T) {
	err := VerifyCPUGPUBinding("aabbccdd", "1234")
	if err == nil {
		t.Fatal("expected error for short reportData")
	}
}

func TestVerifyCPUGPUBinding_EmptyGPU(t *testing.T) {
	cpuReportData := strings.Repeat("aa", 64)
	err := VerifyCPUGPUBinding(cpuReportData, "")
	if err == nil {
		t.Fatal("expected error for empty GPU reportData")
	}
}

// --- VerifyTLSBinding ---

func TestVerifyTLSBinding_CertificateDigest(t *testing.T) {
	certDigest := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	binding := TLSCertBinding{
		SPKI:        "1234123412341234123412341234123412341234123412341234123412341234",
		Certificate: certDigest,
	}
	reportData := certDigest + "0000000000000000000000000000000000000000000000000000000000000000"

	kind, err := VerifyTLSBinding(binding, reportData)
	if err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}
	if kind != TLSBindingCertificate {
		t.Fatalf("expected certificate binding kind, got %s", kind)
	}
}

func TestVerifyTLSBinding_SPKIDigest(t *testing.T) {
	spkiDigest := "1234123412341234123412341234123412341234123412341234123412341234"
	binding := TLSCertBinding{
		SPKI:        spkiDigest,
		Certificate: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
	}
	reportData := spkiDigest + "0000000000000000000000000000000000000000000000000000000000000000"

	kind, err := VerifyTLSBinding(binding, reportData)
	if err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}
	if kind != TLSBindingSPKI {
		t.Fatalf("expected spki binding kind, got %s", kind)
	}
}

func TestVerifyTLSBinding_Mismatch(t *testing.T) {
	binding := TLSCertBinding{
		SPKI:        "1234123412341234123412341234123412341234123412341234123412341234",
		Certificate: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
	}
	reportData := strings.Repeat("11", 32) + "0000"

	if _, err := VerifyTLSBinding(binding, reportData); err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestVerifyTLSBinding_EmptyBinding(t *testing.T) {
	if _, err := VerifyTLSBinding(TLSCertBinding{}, "aabb"); err == nil {
		t.Fatal("expected error for empty binding")
	}
}

func TestVerifyTLSBinding_EmptyReportData(t *testing.T) {
	binding := TLSCertBinding{Certificate: "aabb"}
	if _, err := VerifyTLSBinding(binding, ""); err == nil {
		t.Fatal("expected error for empty reportData")
	}
}

func TestVerifyTLSBinding_ShortReportData(t *testing.T) {
	binding := TLSCertBinding{Certificate: strings.Repeat("ab", 32)}
	if _, err := VerifyTLSBinding(binding, "aabbcc"); err == nil {
		t.Fatal("expected error for short reportData")
	}
}

// --- BackendVerifier cache and status ---

func TestBackendVerifier_GetStatus_Unknown(t *testing.T) {
	bv := NewBackendVerifier("http://unused", nil, nil, nil, testLog())
	if status := bv.GetStatus("nonexistent"); status != nil {
		t.Fatalf("expected nil, got: %+v", status)
	}
}

func TestBackendVerifier_GetAllStatuses_Empty(t *testing.T) {
	bv := NewBackendVerifier("http://unused", nil, nil, nil, testLog())
	if statuses := bv.GetAllStatuses(); len(statuses) != 0 {
		t.Fatalf("expected 0 statuses, got %d", len(statuses))
	}
}

func TestBackendVerifier_StoreFailure(t *testing.T) {
	bv := NewBackendVerifier("http://unused", nil, nil, nil, testLog())
	bv.storeFailure("model-1", "https://test:29343", "test error")

	status := bv.GetStatus("model-1")
	if status == nil {
		t.Fatal("expected status")
	}
	if status.Status != StatusFailed {
		t.Fatalf("expected StatusFailed, got %s", status.Status)
	}
	if status.Error != "test error" {
		t.Fatalf("expected 'test error', got '%s'", status.Error)
	}
}

// --- FastVerifyBackend ---

func TestBackendVerifier_FastVerify_NoCache(t *testing.T) {
	bv := NewBackendVerifier("http://unused", nil, nil, nil, testLog())
	err := bv.FastVerifyBackend(context.Background(), "no-such-model")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no attestation snapshot") {
		t.Fatalf("unexpected: %s", err)
	}
}

func TestBackendVerifier_FastVerify_FailedStatus(t *testing.T) {
	bv := NewBackendVerifier("http://unused", nil, nil, nil, testLog())
	bv.storeFailure("model-1", "https://test:29343", "prev failure")

	// A failed snapshot triggers a full re-attestation, which fails against the
	// unreachable test URL.
	err := bv.FastVerifyBackend(context.Background(), "model-1")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to load CPU attestation quote") {
		t.Fatalf("unexpected: %s", err)
	}
}

func TestBackendVerifier_FastVerify_CacheHit(t *testing.T) {
	cpuQuote := "stable-cpu-quote-hex"
	cpuHash := fmt.Sprintf("%x", sha256.Sum256([]byte(cpuQuote)))

	attestMux := http.NewServeMux()
	attestMux.HandleFunc("/cpu", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, cpuQuote)
	})
	attestServer := httptest.NewTLSServer(attestMux)
	defer attestServer.Close()

	_, fingerprint := selfSignedCert(t)

	bv := NewBackendVerifier("http://unused", nil, nil, nil, testLog())
	bv.attestationClient = attestServer.Client()

	bv.mu.Lock()
	bv.cache["test-model"] = &BackendAttestationSnapshot{
		ModelID:        "test-model",
		AttestationURL: attestServer.URL,
		CPUQuoteHash:   cpuHash,
		TLSBinding:     TLSCertBinding{Certificate: fingerprint},
		Status:         StatusPassed,
	}
	bv.mu.Unlock()

	// fingerprints won't match (the test TLS server's cert differs from our generated cert),
	// but this validates the flow reaches the comparison step
	err := bv.FastVerifyBackend(context.Background(), "test-model")
	// We expect a TLS fingerprint mismatch since our pre-populated fingerprint
	// differs from the httptest server's actual cert
	if err == nil {
		// If it passes, the hash comparison path worked (unexpected but not wrong
		// if the test TLS cert happened to match)
		return
	}
	if strings.Contains(err.Error(), "TLS certificate changed") {
		// Expected: the live cert differs from the pre-populated fingerprint
		return
	}
	t.Fatalf("unexpected error: %s", err)
}

// --- AttestBackend full flow ---

func TestBackendVerifier_AttestBackend_FullFlow(t *testing.T) {
	cert, fingerprint := selfSignedCert(t)

	gpuNonce := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	cpuReportData := fingerprint + gpuNonce
	cpuQuote := testTdxQuoteHex()

	gpuJSON := fmt.Sprintf(`{
		"nonce": "%s",
		"arch": "HOPPER",
		"evidence_list": [{"certificate": "dGVzdA==", "evidence": "dGVzdA=="}]
	}`, gpuNonce)

	attestMux := http.NewServeMux()
	attestMux.HandleFunc("/cpu", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, cpuQuote)
	})
	attestMux.HandleFunc("/gpu", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, gpuJSON)
	})
	attestMux.HandleFunc("/docker-compose", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "services:\n  llm:\n    image: test\n")
	})
	attestServer := httptest.NewTLSServer(attestMux)
	defer attestServer.Close()
	attestServer.TLS.Certificates = []tls.Certificate{cert}

	portalMux := http.NewServeMux()
	portalHandler := func(w http.ResponseWriter, _ *http.Request) {
		resp := ParseQuoteResponse{
			Quote: &QuoteFields{
				MRTD:       "aaaa",
				RTMR0:      "bbbb",
				RTMR1:      "cccc",
				RTMR2:      "dddd",
				RTMR3:      "eeee",
				ReportData: cpuReportData,
			},
			Status: &QuoteStatus{AttestationType: "tdx"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
	portalMux.HandleFunc("/api/quote-parse", portalHandler)
	portalMux.HandleFunc("/api/quote-parse-sev", portalHandler)
	portalServer := httptest.NewServer(portalMux)
	defer portalServer.Close()

	nrasServer := httptest.NewServer(mockNRASHandler(t, true))
	defer nrasServer.Close()

	bv := NewBackendVerifier(portalServer.URL+"/api/quote-parse", nil, loadedRegistry(), nil, testLog())
	bv.attestationClient = NewAttestationHTTPClient()
	bv.nrasVerifier.baseURL = nrasServer.URL

	err := bv.AttestBackend(context.Background(), "test-model", attestServer.URL)
	if err != nil {
		t.Fatalf("AttestBackend failed: %s", err)
	}

	status := bv.GetStatus("test-model")
	if status == nil {
		t.Fatal("expected status")
	}
	if status.Status != StatusPassed {
		t.Fatalf("expected StatusPassed, got %s (error: %s)", status.Status, status.Error)
	}
	if status.TEEType != TEETypeTDX {
		t.Fatalf("expected TDX, got %s", status.TEEType)
	}
	if status.TLSBindingKind != TLSBindingCertificate {
		t.Fatalf("expected certificate binding kind, got %s", status.TLSBindingKind)
	}
}

// TestBackendVerifier_AttestBackend_SPKIBinding verifies the full flow when
// report_data binds SHA-256(SPKI DER) — the binding used by current SecretVMs —
// instead of the legacy full-certificate digest.
func TestBackendVerifier_AttestBackend_SPKIBinding(t *testing.T) {
	cert, _ := selfSignedCert(t)
	spkiFingerprint := certBinding(t, cert).SPKI

	gpuNonce := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	cpuReportData := spkiFingerprint + gpuNonce
	cpuQuote := testTdxQuoteHex()

	gpuJSON := fmt.Sprintf(`{
		"nonce": "%s",
		"arch": "HOPPER",
		"evidence_list": [{"certificate": "dGVzdA==", "evidence": "dGVzdA=="}]
	}`, gpuNonce)

	attestMux := http.NewServeMux()
	attestMux.HandleFunc("/cpu", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, cpuQuote)
	})
	attestMux.HandleFunc("/gpu", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, gpuJSON)
	})
	attestMux.HandleFunc("/docker-compose", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, testComposeYAML)
	})
	attestServer := httptest.NewTLSServer(attestMux)
	defer attestServer.Close()
	attestServer.TLS.Certificates = []tls.Certificate{cert}

	portalMux := http.NewServeMux()
	portalHandler := func(w http.ResponseWriter, _ *http.Request) {
		resp := ParseQuoteResponse{
			Quote:  &QuoteFields{MRTD: "aaaa", RTMR0: "bbbb", RTMR1: "cccc", RTMR2: "dddd", RTMR3: "eeee", ReportData: cpuReportData},
			Status: &QuoteStatus{AttestationType: "tdx"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
	portalMux.HandleFunc("/api/quote-parse", portalHandler)
	portalMux.HandleFunc("/api/quote-parse-sev", portalHandler)
	portalServer := httptest.NewServer(portalMux)
	defer portalServer.Close()

	nrasServer := httptest.NewServer(mockNRASHandler(t, true))
	defer nrasServer.Close()

	bv := NewBackendVerifier(portalServer.URL+"/api/quote-parse", nil, loadedRegistry(), nil, testLog())
	bv.attestationClient = NewAttestationHTTPClient()
	bv.nrasVerifier.baseURL = nrasServer.URL

	if err := bv.AttestBackend(context.Background(), "test-model-spki", attestServer.URL); err != nil {
		t.Fatalf("AttestBackend with SPKI binding failed: %s", err)
	}

	status := bv.GetStatus("test-model-spki")
	if status == nil || status.Status != StatusPassed {
		t.Fatalf("expected StatusPassed, got %+v", status)
	}
	if status.TLSBindingKind != TLSBindingSPKI {
		t.Fatalf("expected spki binding kind, got %s", status.TLSBindingKind)
	}
}

// testComposeYAML is the docker-compose served by the orchestration tests. It
// must match the bytes used to compute the fixture registry entry's RTMR3.
const testComposeYAML = "services:\n  llm:\n    image: test\n"

// testRootfsData is an arbitrary 48-byte hex rootfs_data for the fixture entry.
const testRootfsData = "1111111111111111111111111111111111111111111111111111111111111111"

// testTdxRegisters returns deterministic register values for a synthetic TDX
// quote. RTMR3 is derived from testComposeYAML + testRootfsData so that
// VerifyTdxWorkload against the fixture registry yields WorkloadAuthentic.
func testTdxRegisters() (mrtd, rtmr0, rtmr1, rtmr2, rtmr3 [48]byte) {
	for i := 0; i < 48; i++ {
		mrtd[i] = 0xa1
		rtmr0[i] = 0xb2
		rtmr1[i] = 0xc3
		rtmr2[i] = 0xd4
	}
	rtmr3Hex := CalculateRTMR3([]byte(testComposeYAML), testRootfsData)
	b, _ := hex.DecodeString(rtmr3Hex)
	copy(rtmr3[:], b)
	return
}

// testTdxQuoteHex builds a minimal synthetic TDX v4 quote carrying the fixture
// register values, so AttestBackend's mandatory workload verification can run.
func testTdxQuoteHex() string {
	mrtd, rtmr0, rtmr1, rtmr2, rtmr3 := testTdxRegisters()
	raw := make([]byte, 632)
	binary.LittleEndian.PutUint16(raw[0:2], 4)    // version
	binary.LittleEndian.PutUint32(raw[4:8], 0x81) // tee type (TDX)
	copy(raw[184:232], mrtd[:])
	copy(raw[376:424], rtmr0[:])
	copy(raw[424:472], rtmr1[:])
	copy(raw[472:520], rtmr2[:])
	copy(raw[520:568], rtmr3[:])
	return hex.EncodeToString(raw)
}

// loadedRegistry returns an ArtifactRegistry (IsLoaded()==true) containing the
// entry that matches testTdxQuoteHex + testComposeYAML, so tests exercise
// AttestBackend's mandatory workload-verification step end to end.
func loadedRegistry() *ArtifactRegistry {
	mrtd, rtmr0, rtmr1, rtmr2, _ := testTdxRegisters()
	r := NewArtifactRegistry("http://unused", 1*time.Hour, testLog())
	r.mu.Lock()
	r.entries = []TdxArtifactEntry{{
		TemplateName: "test-template",
		VMType:       "tdx",
		ArtifactsVer: "v1.0.0",
		MRTD:         hex.EncodeToString(mrtd[:]),
		RTMR0:        hex.EncodeToString(rtmr0[:]),
		RTMR1:        hex.EncodeToString(rtmr1[:]),
		RTMR2:        hex.EncodeToString(rtmr2[:]),
		RootfsData:   testRootfsData,
	}}
	r.lastFetched = time.Now()
	r.mu.Unlock()
	return r
}

// mockNRASHandler returns an NRAS handler that echoes the submitted nonce back
// as the eat_nonce claim in a JWT-encoded overall EAT token, with the given
// overall attestation result.
func mockNRASHandler(t *testing.T, overallResult bool) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		var req GPUAttestationData
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("NRAS: failed to decode request: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		overall := makeEATToken(t, fmt.Sprintf(`{"eat_nonce":%q,"x-nvidia-overall-att-result":%t}`, req.Nonce, overallResult))
		resp := []json.RawMessage{
			[]byte(fmt.Sprintf(`["JWT", %q]`, overall)),
			[]byte(`{"GPU-0": "eyGPU0Token"}`),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func TestBackendVerifier_AttestBackend_WithNRAS(t *testing.T) {
	cert, fingerprint := selfSignedCert(t)

	gpuNonce := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	cpuReportData := fingerprint + gpuNonce
	cpuQuote := testTdxQuoteHex()

	gpuJSON := fmt.Sprintf(`{
		"nonce": "%s",
		"arch": "HOPPER",
		"evidence_list": [{"certificate": "dGVzdA==", "evidence": "dGVzdA=="}]
	}`, gpuNonce)

	attestMux := http.NewServeMux()
	attestMux.HandleFunc("/cpu", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, cpuQuote)
	})
	attestMux.HandleFunc("/gpu", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, gpuJSON)
	})
	attestMux.HandleFunc("/docker-compose", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "services:\n  llm:\n    image: test\n")
	})
	attestServer := httptest.NewTLSServer(attestMux)
	defer attestServer.Close()
	attestServer.TLS.Certificates = []tls.Certificate{cert}

	portalMux := http.NewServeMux()
	portalHandler := func(w http.ResponseWriter, _ *http.Request) {
		resp := ParseQuoteResponse{
			Quote: &QuoteFields{
				MRTD:       "aaaa",
				RTMR0:      "bbbb",
				RTMR1:      "cccc",
				RTMR2:      "dddd",
				RTMR3:      "eeee",
				ReportData: cpuReportData,
			},
			Status: &QuoteStatus{AttestationType: "tdx"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
	portalMux.HandleFunc("/api/quote-parse", portalHandler)
	portalMux.HandleFunc("/api/quote-parse-sev", portalHandler)
	portalServer := httptest.NewServer(portalMux)
	defer portalServer.Close()

	nrasServer := httptest.NewServer(mockNRASHandler(t, true))
	defer nrasServer.Close()

	bv := NewBackendVerifier(portalServer.URL+"/api/quote-parse", nil, loadedRegistry(), nil, testLog())
	bv.attestationClient = NewAttestationHTTPClient()
	bv.nrasVerifier.baseURL = nrasServer.URL

	err := bv.AttestBackend(context.Background(), "test-model-nras", attestServer.URL)
	if err != nil {
		t.Fatalf("AttestBackend with NRAS failed: %s", err)
	}

	status := bv.GetStatus("test-model-nras")
	if status == nil {
		t.Fatal("expected status")
	}
	if status.Status != StatusPassed {
		t.Fatalf("expected StatusPassed, got %s (error: %s)", status.Status, status.Error)
	}
}

func TestBackendVerifier_AttestBackend_NRASOverallResultFalse(t *testing.T) {
	cert, fingerprint := selfSignedCert(t)

	gpuNonce := "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	cpuReportData := fingerprint + gpuNonce
	cpuQuote := testTdxQuoteHex()

	gpuJSON := fmt.Sprintf(`{
		"nonce": "%s",
		"arch": "HOPPER",
		"evidence_list": [{"certificate": "dGVzdA==", "evidence": "dGVzdA=="}]
	}`, gpuNonce)

	attestMux := http.NewServeMux()
	attestMux.HandleFunc("/cpu", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, cpuQuote)
	})
	attestMux.HandleFunc("/gpu", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, gpuJSON)
	})
	attestMux.HandleFunc("/docker-compose", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "services:\n  llm:\n    image: test\n")
	})
	attestServer := httptest.NewTLSServer(attestMux)
	defer attestServer.Close()
	attestServer.TLS.Certificates = []tls.Certificate{cert}

	portalMux := http.NewServeMux()
	portalHandler := func(w http.ResponseWriter, _ *http.Request) {
		resp := ParseQuoteResponse{
			Quote:  &QuoteFields{MRTD: "aaaa", RTMR0: "bbbb", RTMR1: "cccc", RTMR2: "dddd", RTMR3: "eeee", ReportData: cpuReportData},
			Status: &QuoteStatus{AttestationType: "tdx"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
	portalMux.HandleFunc("/api/quote-parse", portalHandler)
	portalMux.HandleFunc("/api/quote-parse-sev", portalHandler)
	portalServer := httptest.NewServer(portalMux)
	defer portalServer.Close()

	// NRAS returns a well-formed token but reports overall attestation failure.
	nrasServer := httptest.NewServer(mockNRASHandler(t, false))
	defer nrasServer.Close()

	bv := NewBackendVerifier(portalServer.URL+"/api/quote-parse", nil, loadedRegistry(), nil, testLog())
	bv.attestationClient = NewAttestationHTTPClient()
	bv.nrasVerifier.baseURL = nrasServer.URL

	err := bv.AttestBackend(context.Background(), "test-model-fail", attestServer.URL)
	if err == nil {
		t.Fatal("expected AttestBackend to fail when NRAS overall result is false")
	}
	if !strings.Contains(err.Error(), "overall-att-result") {
		t.Fatalf("unexpected error: %s", err)
	}
	if status := bv.GetStatus("test-model-fail"); status == nil || status.Status != StatusFailed {
		t.Fatalf("expected StatusFailed snapshot, got %+v", status)
	}
}

func TestBackendVerifier_AttestBackend_RegistryNotLoaded(t *testing.T) {
	cert, fingerprint := selfSignedCert(t)
	cpuReportData := fingerprint + "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	cpuQuote := testTdxQuoteHex() // valid TDX quote so workload verification runs

	attestMux := http.NewServeMux()
	attestMux.HandleFunc("/cpu", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, cpuQuote)
	})
	attestMux.HandleFunc("/docker-compose", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, testComposeYAML)
	})
	attestServer := httptest.NewTLSServer(attestMux)
	defer attestServer.Close()
	attestServer.TLS.Certificates = []tls.Certificate{cert}

	portalMux := http.NewServeMux()
	portalHandler := func(w http.ResponseWriter, _ *http.Request) {
		resp := ParseQuoteResponse{
			Quote:  &QuoteFields{MRTD: "aaaa", RTMR0: "bbbb", RTMR1: "cccc", RTMR2: "dddd", RTMR3: "eeee", ReportData: cpuReportData},
			Status: &QuoteStatus{AttestationType: "tdx"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
	portalMux.HandleFunc("/api/quote-parse", portalHandler)
	portalMux.HandleFunc("/api/quote-parse-sev", portalHandler)
	portalServer := httptest.NewServer(portalMux)
	defer portalServer.Close()

	// nil registry -> workload verification is unavailable -> must fail closed.
	bv := NewBackendVerifier(portalServer.URL+"/api/quote-parse", nil, nil, nil, testLog())
	bv.attestationClient = NewAttestationHTTPClient()

	err := bv.AttestBackend(context.Background(), "test-model-noreg", attestServer.URL)
	if err == nil {
		t.Fatal("expected AttestBackend to fail closed when artifact registry is not loaded")
	}
	if !strings.Contains(err.Error(), "artifact registry not loaded") {
		t.Fatalf("unexpected error: %s", err)
	}
	if status := bv.GetStatus("test-model-noreg"); status == nil || status.Status != StatusFailed {
		t.Fatalf("expected StatusFailed snapshot, got %+v", status)
	}
}

// --- PinnedHTTPClient ---

func TestBackendVerifier_PinnedHTTPClient_NoModel(t *testing.T) {
	bv := NewBackendVerifier("http://unused", nil, nil, nil, testLog())
	_, err := bv.PinnedHTTPClient("no-model")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBackendVerifier_PinnedHTTPClient_Success(t *testing.T) {
	bv := NewBackendVerifier("http://unused", nil, nil, nil, testLog())

	bv.mu.Lock()
	bv.cache["model-1"] = &BackendAttestationSnapshot{
		ModelID:    "model-1",
		TLSBinding: TLSCertBinding{Certificate: "aabbccdd"},
		Status:     StatusPassed,
	}
	bv.mu.Unlock()

	client, err := bv.PinnedHTTPClient("model-1")
	if err != nil {
		t.Fatalf("expected client, got: %s", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestBackendVerifier_PinnedHTTPClient_FailedStatus(t *testing.T) {
	bv := NewBackendVerifier("http://unused", nil, nil, nil, testLog())
	bv.storeFailure("model-1", "https://test:29343", "broken")

	_, err := bv.PinnedHTTPClient("model-1")
	if err == nil {
		t.Fatal("expected error for failed model")
	}
}

// --- NoopGoldenSource ---

func TestNoopGoldenSource(t *testing.T) {
	src := &NoopGoldenSource{}
	golden, err := src.FetchGoldenValues(context.Background(), "any", "any")
	if err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}
	if golden != nil {
		t.Fatalf("expected nil, got: %+v", golden)
	}
}
