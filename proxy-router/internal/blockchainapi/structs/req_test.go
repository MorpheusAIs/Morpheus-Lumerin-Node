package structs

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestOpenSessionWithFailoverOmitProvider(t *testing.T) {
	var req OpenSessionWithFailover
	body := `{"sessionDuration":3600,"failover":false,"omitProvider":"0xAbC1234567890abcdef1234567890abcdef12345"}`
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	want := common.HexToAddress("0xAbC1234567890abcdef1234567890abcdef12345")
	if req.OmitProvider.Address != want {
		t.Errorf("OmitProvider = %s, want %s", req.OmitProvider.Address.Hex(), want.Hex())
	}
}

func TestOpenSessionWithFailoverOmitProviderDefaultsToZero(t *testing.T) {
	// Older gateways don't send the field; bid selection must not skip anyone.
	var req OpenSessionWithFailover
	body := `{"sessionDuration":3600,"failover":true}`
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if req.OmitProvider.Address != (common.Address{}) {
		t.Errorf("OmitProvider = %s, want zero address", req.OmitProvider.Address.Hex())
	}
}
