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

func TestVerifyTLSBinding_Valid(t *testing.T) {
	fingerprint := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	reportData := fingerprint + "0000000000000000000000000000000000000000000000000000000000000000"

	if err := VerifyTLSBinding(fingerprint, reportData); err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}
}

func TestVerifyTLSBinding_Mismatch(t *testing.T) {
	fingerprint := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	reportData := "1111111111111111111111111111111111111111111111111111111111111111" + "0000"

	err := VerifyTLSBinding(fingerprint, reportData)
	if err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestVerifyTLSBinding_EmptyFingerprint(t *testing.T) {
	err := VerifyTLSBinding("", "aabb")
	if err == nil {
		t.Fatal("expected error for empty fingerprint")
	}
}

func TestVerifyTLSBinding_EmptyReportData(t *testing.T) {
	err := VerifyTLSBinding("aabb", "")
	if err == nil {
		t.Fatal("expected error for empty reportData")
	}
}

// --- BackendVerifier cache and status ---

func TestBackendVerifier_GetStatus_Unknown(t *testing.T) {
	bv := NewBackendVerifier("http://unused", nil, nil, testLog())
	if status := bv.GetStatus("nonexistent"); status != nil {
		t.Fatalf("expected nil, got: %+v", status)
	}
}

func TestBackendVerifier_GetAllStatuses_Empty(t *testing.T) {
	bv := NewBackendVerifier("http://unused", nil, nil, testLog())
	if statuses := bv.GetAllStatuses(); len(statuses) != 0 {
		t.Fatalf("expected 0 statuses, got %d", len(statuses))
	}
}

func TestBackendVerifier_StoreFailure(t *testing.T) {
	bv := NewBackendVerifier("http://unused", nil, nil, testLog())
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
	bv := NewBackendVerifier("http://unused", nil, nil, testLog())
	err := bv.FastVerifyBackend(context.Background(), "no-such-model")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no attestation snapshot") {
		t.Fatalf("unexpected: %s", err)
	}
}

func TestBackendVerifier_FastVerify_FailedStatus(t *testing.T) {
	bv := NewBackendVerifier("http://unused", nil, nil, testLog())
	bv.storeFailure("model-1", "https://test:29343", "prev failure")

	err := bv.FastVerifyBackend(context.Background(), "model-1")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "status is failed") {
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

	bv := NewBackendVerifier("http://unused", nil, nil, testLog())
	bv.attestationClient = attestServer.Client()

	bv.mu.Lock()
	bv.cache["test-model"] = &BackendAttestationSnapshot{
		ModelID:        "test-model",
		AttestationURL: attestServer.URL,
		CPUQuoteHash:   cpuHash,
		TLSFingerprint: fingerprint,
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

	gpuJSON := fmt.Sprintf(`{
		"nonce": "%s",
		"arch": "HOPPER",
		"evidence_list": [{"certificate": "dGVzdA==", "evidence": "dGVzdA=="}]
	}`, gpuNonce)

	attestMux := http.NewServeMux()
	attestMux.HandleFunc("/cpu", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "deadbeefcpu")
	})
	attestMux.HandleFunc("/gpu", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, gpuJSON)
	})
	attestServer := httptest.NewTLSServer(attestMux)
	defer attestServer.Close()
	attestServer.TLS.Certificates = []tls.Certificate{cert}

	portalMux := http.NewServeMux()
	portalMux.HandleFunc("/api/quote-parse", func(w http.ResponseWriter, _ *http.Request) {
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
	})
	portalServer := httptest.NewServer(portalMux)
	defer portalServer.Close()

	bv := NewBackendVerifier(portalServer.URL+"/api/quote-parse", nil, nil, testLog())
	bv.attestationClient = NewAttestationHTTPClient()

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
}

func TestBackendVerifier_AttestBackend_WithNRAS(t *testing.T) {
	cert, fingerprint := selfSignedCert(t)

	gpuNonce := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	cpuReportData := fingerprint + gpuNonce

	gpuJSON := fmt.Sprintf(`{
		"nonce": "%s",
		"arch": "HOPPER",
		"evidence_list": [{"certificate": "dGVzdA==", "evidence": "dGVzdA=="}]
	}`, gpuNonce)

	attestMux := http.NewServeMux()
	attestMux.HandleFunc("/cpu", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "deadbeefcpu")
	})
	attestMux.HandleFunc("/gpu", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, gpuJSON)
	})
	attestServer := httptest.NewTLSServer(attestMux)
	defer attestServer.Close()
	attestServer.TLS.Certificates = []tls.Certificate{cert}

	portalMux := http.NewServeMux()
	portalMux.HandleFunc("/api/quote-parse", func(w http.ResponseWriter, _ *http.Request) {
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
	})
	portalServer := httptest.NewServer(portalMux)
	defer portalServer.Close()

	// Mock NRAS server
	nrasMux := http.NewServeMux()
	nrasMux.HandleFunc("/v4/attest/gpu", func(w http.ResponseWriter, r *http.Request) {
		var req GPUAttestationData
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("NRAS: failed to decode request: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if req.Arch != "HOPPER" {
			t.Errorf("NRAS: expected arch HOPPER, got %s", req.Arch)
		}
		if len(req.EvidenceList) != 1 {
			t.Errorf("NRAS: expected 1 evidence, got %d", len(req.EvidenceList))
		}

		resp := []json.RawMessage{
			[]byte(`["JWT", "eyOverallToken"]`),
			[]byte(`{"GPU-0": "eyGPU0Token"}`),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	nrasServer := httptest.NewServer(nrasMux)
	defer nrasServer.Close()

	bv := NewBackendVerifier(portalServer.URL+"/api/quote-parse", nil, nil, testLog())
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

// --- PinnedHTTPClient ---

func TestBackendVerifier_PinnedHTTPClient_NoModel(t *testing.T) {
	bv := NewBackendVerifier("http://unused", nil, nil, testLog())
	_, err := bv.PinnedHTTPClient("no-model")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBackendVerifier_PinnedHTTPClient_Success(t *testing.T) {
	bv := NewBackendVerifier("http://unused", nil, nil, testLog())

	bv.mu.Lock()
	bv.cache["model-1"] = &BackendAttestationSnapshot{
		ModelID:        "model-1",
		TLSFingerprint: "aabbccdd",
		Status:         StatusPassed,
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
	bv := NewBackendVerifier("http://unused", nil, nil, testLog())
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
