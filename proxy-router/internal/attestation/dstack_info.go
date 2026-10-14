package attestation

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// dstackInfoTimeout bounds the best-effort GET /info. Older SecretVM images
// have no /info endpoint at all, so a slow failure must not stall attestation.
const dstackInfoTimeout = 10 * time.Second

// ExtractDstackAppID pulls `dstack_app_id` out of a SecretVM's GET /info JSON
// body, normalized to lowercase hex without a 0x prefix. Returns "" when the
// field is missing, empty, non-hex, or the body is not JSON — callers treat
// "" as "no app-id" and replay RTMR3 with the attest-tool scheme.
func ExtractDstackAppID(infoBody []byte) string {
	var parsed struct {
		DstackAppID string `json:"dstack_app_id"`
	}
	if err := json.Unmarshal(infoBody, &parsed); err != nil {
		return ""
	}
	return NormalizeDstackAppID(parsed.DstackAppID)
}

// FetchDstackAppID performs a best-effort GET {attestationBaseURL}/info and
// returns the VM's dstack app-id, or "" on any failure (no endpoint, non-200,
// bad body). The value is self-reported by the VM and is only trustworthy as
// an input to an RTMR3 replay that reproduces the hardware-signed quote: a
// wrong or forged id simply fails the register comparison.
func FetchDstackAppID(ctx context.Context, client *http.Client, attestationBaseURL string) string {
	if client == nil {
		return ""
	}
	reqCtx, cancel := context.WithTimeout(ctx, dstackInfoTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, attestationBaseURL+"/info", nil)
	if err != nil {
		return ""
	}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return ""
	}
	return ExtractDstackAppID(body)
}
