package attestation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

func TestParseGPUAttestationData_Valid(t *testing.T) {
	raw := `{
		"nonce": "a768bb3a38c48d3636b9e8250a3b12da05f557eca0b484bf9d8fa5719c0341bd",
		"arch": "HOPPER",
		"evidence_list": [
			{"certificate": "LS0tLS1CRUdJTg==", "evidence": "EeAB/6do"}
		]
	}`

	data, err := ParseGPUAttestationData(raw)
	if err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}
	if data.Nonce != "a768bb3a38c48d3636b9e8250a3b12da05f557eca0b484bf9d8fa5719c0341bd" {
		t.Fatalf("unexpected nonce: %s", data.Nonce)
	}
	if data.Arch != "HOPPER" {
		t.Fatalf("unexpected arch: %s", data.Arch)
	}
	if len(data.EvidenceList) != 1 {
		t.Fatalf("expected 1 evidence, got %d", len(data.EvidenceList))
	}
}

func TestParseGPUAttestationData_MissingNonce(t *testing.T) {
	raw := `{"arch": "HOPPER", "evidence_list": [{"certificate": "x", "evidence": "y"}]}`
	_, err := ParseGPUAttestationData(raw)
	if err == nil {
		t.Fatal("expected error for missing nonce")
	}
}

func TestParseGPUAttestationData_MissingArch(t *testing.T) {
	raw := `{"nonce": "abc", "evidence_list": [{"certificate": "x", "evidence": "y"}]}`
	_, err := ParseGPUAttestationData(raw)
	if err == nil {
		t.Fatal("expected error for missing arch")
	}
}

func TestParseGPUAttestationData_EmptyEvidenceList(t *testing.T) {
	raw := `{"nonce": "abc", "arch": "HOPPER", "evidence_list": []}`
	_, err := ParseGPUAttestationData(raw)
	if err == nil {
		t.Fatal("expected error for empty evidence_list")
	}
}

func TestParseGPUAttestationData_InvalidJSON(t *testing.T) {
	_, err := ParseGPUAttestationData("not json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseNRASResponse_Valid(t *testing.T) {
	resp := `[["JWT", "eyOverall"], {"GPU-0": "eyGPU0", "GPU-1": "eyGPU1"}]`
	result, err := parseNRASResponse([]byte(resp))
	if err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}
	if result.OverallToken != "eyOverall" {
		t.Fatalf("unexpected overall token: %s", result.OverallToken)
	}
	if len(result.GPUTokens) != 2 {
		t.Fatalf("expected 2 GPU tokens, got %d", len(result.GPUTokens))
	}
	if result.GPUTokens["GPU-0"] != "eyGPU0" {
		t.Fatalf("unexpected GPU-0 token: %s", result.GPUTokens["GPU-0"])
	}
}

func TestParseNRASResponse_InvalidFormat(t *testing.T) {
	_, err := parseNRASResponse([]byte(`"not an array"`))
	if err == nil {
		t.Fatal("expected error for non-array response")
	}
}

func TestParseNRASResponse_TooFewElements(t *testing.T) {
	_, err := parseNRASResponse([]byte(`[["JWT","tok"]]`))
	if err == nil {
		t.Fatal("expected error for too few elements")
	}
}

func TestNRASVerifier_VerifyGPU_Success(t *testing.T) {
	nrasMux := http.NewServeMux()
	nrasMux.HandleFunc("/v4/attest/gpu", func(w http.ResponseWriter, r *http.Request) {
		var req GPUAttestationData
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if req.Arch != "HOPPER" {
			t.Errorf("expected HOPPER, got %s", req.Arch)
		}
		if req.Nonce != "testnonce123" {
			t.Errorf("expected testnonce123, got %s", req.Nonce)
		}
		if len(req.EvidenceList) != 1 {
			t.Errorf("expected 1 evidence, got %d", len(req.EvidenceList))
		}

		resp := []json.RawMessage{
			[]byte(`["JWT", "eyOverallToken"]`),
			[]byte(`{"GPU-0": "eyGPU0Token"}`),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(nrasMux)
	defer server.Close()

	nv := NewNRASVerifier(&lib.LoggerMock{})
	nv.baseURL = server.URL

	gpuData := &GPUAttestationData{
		Nonce: "testnonce123",
		Arch:  "HOPPER",
		EvidenceList: []GPUEvidence{
			{Certificate: "dGVzdA==", Evidence: "dGVzdA=="},
		},
	}

	result, err := nv.VerifyGPU(context.Background(), gpuData)
	if err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}
	if result.OverallToken != "eyOverallToken" {
		t.Fatalf("unexpected overall token: %s", result.OverallToken)
	}
	if result.GPUTokens["GPU-0"] != "eyGPU0Token" {
		t.Fatalf("unexpected GPU-0 token: %s", result.GPUTokens["GPU-0"])
	}
}

func TestNRASVerifier_VerifyGPU_ServerError(t *testing.T) {
	nrasMux := http.NewServeMux()
	nrasMux.HandleFunc("/v4/attest/gpu", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	})
	server := httptest.NewServer(nrasMux)
	defer server.Close()

	nv := NewNRASVerifier(&lib.LoggerMock{})
	nv.baseURL = server.URL

	gpuData := &GPUAttestationData{
		Nonce:        "testnonce",
		Arch:         "HOPPER",
		EvidenceList: []GPUEvidence{{Certificate: "x", Evidence: "y"}},
	}

	_, err := nv.VerifyGPU(context.Background(), gpuData)
	if err == nil {
		t.Fatal("expected error for server error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Fatalf("expected status 500 in error, got: %s", err)
	}
}

func TestNRASVerifier_VerifyGPU_EmptyEvidence(t *testing.T) {
	nv := NewNRASVerifier(&lib.LoggerMock{})

	gpuData := &GPUAttestationData{
		Nonce:        "testnonce",
		Arch:         "HOPPER",
		EvidenceList: []GPUEvidence{},
	}

	_, err := nv.VerifyGPU(context.Background(), gpuData)
	if err == nil {
		t.Fatal("expected error for empty evidence")
	}
}
