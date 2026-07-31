package attestation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseQuoteResponse_NumericVersionAndTeeType(t *testing.T) {
	// Live SecretAI Portal shape (TDX): version/tee_type are JSON numbers.
	raw := []byte(`{
		"status": {"attestation_type": "tdx", "result": "ok"},
		"quote": {
			"attestation_type": "tdx",
			"version": 4,
			"tee_type": 129,
			"mr_td": "abcd",
			"rtmr0": "1111",
			"rtmr1": "2222",
			"rtmr2": "3333",
			"rtmr3": "4444",
			"report_data": "deadbeef"
		}
	}`)

	var parsed ParseQuoteResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("unmarshal portal response: %v", err)
	}
	if parsed.Quote == nil {
		t.Fatal("expected quote")
	}
	if got := parsed.Quote.Version.String(); got != "4" {
		t.Fatalf("Version = %q, want 4", got)
	}
	if got := parsed.Quote.TEEType.String(); got != "129" {
		t.Fatalf("TEEType = %q, want 129", got)
	}
	if parsed.Quote.MRTD != "abcd" {
		t.Fatalf("MRTD = %q, want abcd", parsed.Quote.MRTD)
	}
}

func TestParseQuoteResponse_StringVersionStillAccepted(t *testing.T) {
	raw := []byte(`{"quote":{"version":"4","tee_type":"TDX","mr_td":"a","rtmr0":"b","rtmr1":"c","rtmr2":"d","report_data":"e"}}`)
	var parsed ParseQuoteResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed.Quote.Version.String() != "4" || parsed.Quote.TEEType.String() != "TDX" {
		t.Fatalf("got version=%q tee_type=%q", parsed.Quote.Version, parsed.Quote.TEEType)
	}
}

func TestParseQuoteResponse_SEVNumericVersion(t *testing.T) {
	raw := []byte(`{
		"status": {"attestation_type": "sev", "verify": "ok"},
		"quote": {
			"attestation_type": "sev",
			"version": 3,
			"report_data": "aa bb",
			"measurement": "ff00"
		}
	}`)
	var parsed ParseQuoteResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("unmarshal SEV portal response: %v", err)
	}
	if parsed.Quote.Version.String() != "3" {
		t.Fatalf("Version = %q, want 3", parsed.Quote.Version)
	}
	if parsed.Quote.Measurement != "ff00" {
		t.Fatalf("Measurement = %q", parsed.Quote.Measurement)
	}
}

func TestVerifyQuote_PortalNumericFields(t *testing.T) {
	portal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": {"attestation_type": "tdx", "result": "ok"},
			"quote": {
				"version": 4,
				"tee_type": 129,
				"mr_td": "aaaa",
				"rtmr0": "bbbb",
				"rtmr1": "cccc",
				"rtmr2": "dddd",
				"rtmr3": "eeee",
				"report_data": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
			}
		}`))
	}))
	defer portal.Close()

	result, err := VerifyQuote(context.Background(), portal.Client(), portal.URL, "0400020081000000deadbeef")
	if err != nil {
		t.Fatalf("VerifyQuote: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid, got error %q", result.Error)
	}
	if result.Type != TEETypeTDX {
		t.Fatalf("type = %s, want TDX", result.Type)
	}
	if result.MRTD != "aaaa" || result.ReportData == "" {
		t.Fatalf("unexpected fields: %+v", result)
	}
}
