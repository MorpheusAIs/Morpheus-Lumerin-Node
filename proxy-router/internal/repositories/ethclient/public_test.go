package ethclient

import "testing"

func TestPublicBaseRPCURLsExcludeRetiredEndpoints(t *testing.T) {
	urls, err := GetPublicRPCURLs(8453)
	if err != nil {
		t.Fatalf("GetPublicRPCURLs() error = %v", err)
	}

	for _, endpoint := range urls {
		if endpoint == "https://base.lava.build" {
			t.Fatal("retired Base Lava endpoint must not be included in the fallback RPC pool")
		}
	}
}
