package attestation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dstackFixtureCompose is scrtlabs/secretvm-verify test-data/dstack_docker_compose.yaml
// byte-for-byte (137 bytes). Its SHA-256 is the compose-hash event payload.
const dstackFixtureCompose = "version: '3'\n\nservices:\n  nginx:\n    image: nginx:latest\n    ports:\n      - \"8080:80\"\n    restart: always\n    env_file:\n      - usr/.env\n"

const dstackFixtureComposeSHA256 = "77f748eb66c0a2466ff7d4c3dc0f66aeab291b8ceaefed4222c38cf0690b5c70"

// Ground-truth vectors captured by SCRT Labs from real SecretVMs running
// small/v0.0.34 (secretvm-verify 0.13.0 node/src/workload.test.ts). The
// expected RTMR3 is the exact value in each machine's TDX quote.
var dstackVectors = []struct {
	name   string
	rootfs string
	appID  string
	rtmr3  string
}{
	{
		name:   "prod (beige-lobster)",
		rootfs: "6a441f52179d66d4b82d8113a99892b1287a01ded04b2f47019750f824071cc0",
		appID:  "6c4ed58ae19edcc3b90a7d7bd9087aee951d1ecc",
		rtmr3:  "bd88eccd83e76e65edc863ad771edd4d1bf388c5c655979774aa476dfedf611df4b05e78cb9dacbfe655dd0ed4a11313",
	},
	{
		name:   "dev (gray-tyrannosaurus)",
		rootfs: "b28b457b468cd058be0732208d2edc79e41f42fd4a0d25353aacacf016b6dc4e",
		appID:  "020e04828a82f7e5466c47279120e1a8d087848a",
		rtmr3:  "7f3c3d4872dcea04729f51573cd30787b02efda396b8264d87082b6f9c8716423738f0038aeecb3e98f2c439664bbc08",
	},
}

func TestDstackFixtureComposeHash(t *testing.T) {
	sum := sha256.Sum256([]byte(dstackFixtureCompose))
	assert.Equal(t, dstackFixtureComposeSHA256, hex.EncodeToString(sum[:]))
}

func TestCalculateRTMR3WithAppID_GroundTruth(t *testing.T) {
	for _, v := range dstackVectors {
		t.Run(v.name, func(t *testing.T) {
			got := CalculateRTMR3WithAppID([]byte(dstackFixtureCompose), v.rootfs, v.appID)
			assert.Equal(t, v.rtmr3, got, "dstack replay must reproduce the live quote's RTMR3")

			// Hash-level form (what the golden-manifest path uses) agrees.
			fromHashes := CalculateRTMR3FromHashes(dstackFixtureComposeSHA256, v.rootfs, v.appID)
			assert.Equal(t, v.rtmr3, fromHashes)

			// "sha256:"-prefixed compose digest (manifest top-level form) is accepted.
			assert.Equal(t, v.rtmr3, CalculateRTMR3FromHashes("sha256:"+dstackFixtureComposeSHA256, v.rootfs, v.appID))
		})
	}
}

func TestCalculateRTMR3WithAppID_SchemesDiffer(t *testing.T) {
	compose := []byte("version: '3'\nservices: {}\n")
	rootfs := "ab" + repeatHex("ab", 31)
	appID := "E418296d0E99734599a4138774E6b85e058A64FE"

	withID := CalculateRTMR3WithAppID(compose, rootfs, appID)
	without := CalculateRTMR3WithAppID(compose, rootfs, "")
	assert.NotEqual(t, withID, without)
	assert.Equal(t, CalculateRTMR3(compose, rootfs), without, "legacy wrapper is the attest-tool scheme")

	// Empty / whitespace app-id falls back to attest-tool.
	assert.Equal(t, without, CalculateRTMR3WithAppID(compose, rootfs, "   "))

	// Case and 0x prefix are normalized.
	assert.Equal(t, withID, CalculateRTMR3WithAppID(compose, rootfs, "0x"+appID))
	assert.Equal(t, withID, CalculateRTMR3WithAppID(compose, rootfs, "0X"+appID))

	assert.Equal(t, RTMR3SchemeDstack, RTMR3SchemeFor(appID))
	assert.Equal(t, RTMR3SchemeAttestTool, RTMR3SchemeFor(""))
	assert.Equal(t, RTMR3SchemeAttestTool, RTMR3SchemeFor("not-hex"), "non-hex app-id is treated as absent")
}

func repeatHex(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

func TestExtractDstackAppID(t *testing.T) {
	assert.Equal(t, "e418296d0e99734599a4138774e6b85e058a64fe",
		ExtractDstackAppID([]byte(`{"dstack_app_id":"E418296d0E99734599a4138774E6b85e058A64FE"}`)))
	assert.Equal(t, "dead", ExtractDstackAppID([]byte(`{"dstack_app_id":"0xDEAD"}`)))
	assert.Equal(t, "", ExtractDstackAppID([]byte(`{"dstack_app_id":""}`)))
	assert.Equal(t, "", ExtractDstackAppID([]byte(`{}`)))
	assert.Equal(t, "", ExtractDstackAppID([]byte(`<html>not json</html>`)))
	assert.Equal(t, "", ExtractDstackAppID([]byte(`{"dstack_app_id":"zz-not-hex"}`)))
}

func TestFetchDstackAppID(t *testing.T) {
	t.Run("newer image serves /info", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/info" {
				_, _ = w.Write([]byte(`{"dstack_app_id":"6c4ed58ae19edcc3b90a7d7bd9087aee951d1ecc"}`))
				return
			}
			http.NotFound(w, r)
		}))
		defer srv.Close()
		assert.Equal(t, "6c4ed58ae19edcc3b90a7d7bd9087aee951d1ecc", FetchDstackAppID(context.Background(), srv.Client(), srv.URL))
	})

	t.Run("older image has no /info", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}))
		defer srv.Close()
		assert.Equal(t, "", FetchDstackAppID(context.Background(), srv.Client(), srv.URL))
	})

	t.Run("unreachable host", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		url := srv.URL
		srv.Close()
		assert.Equal(t, "", FetchDstackAppID(context.Background(), http.DefaultClient, url))
	})

	t.Run("nil client", func(t *testing.T) {
		assert.Equal(t, "", FetchDstackAppID(context.Background(), nil, "https://example.invalid"))
	})
}

// TestCompareRegisters_TDX_Dstack covers the golden-manifest path for the two
// RTMR3 schemes and the fail-closed cases between them.
func TestCompareRegisters_TDX_Dstack(t *testing.T) {
	v := dstackVectors[0]
	legacyRTMR3 := CalculateRTMR3FromHashes(dstackFixtureComposeSHA256, v.rootfs, "")

	fullGolden := &GoldenValues{
		RTMR3:            legacyRTMR3,
		TDXComposeSHA256: dstackFixtureComposeSHA256,
		TDXRootfsSHA256:  v.rootfs,
	}
	legacyOnlyGolden := &GoldenValues{RTMR3: legacyRTMR3}

	t.Run("pre-dstack VM matches published RTMR3", func(t *testing.T) {
		res := &AttestationResult{Type: TEETypeTDX, RTMR3: legacyRTMR3}
		assert.NoError(t, CompareRegisters(res, fullGolden, nil))
		assert.NoError(t, CompareRegisters(res, legacyOnlyGolden, nil))
	})

	t.Run("dstack VM matches replay from compose+rootfs+app-id", func(t *testing.T) {
		res := &AttestationResult{Type: TEETypeTDX, RTMR3: v.rtmr3, DstackAppID: v.appID}
		assert.NoError(t, CompareRegisters(res, fullGolden, nil))
	})

	t.Run("dstack VM against legacy-only manifest fails closed", func(t *testing.T) {
		res := &AttestationResult{Type: TEETypeTDX, RTMR3: v.rtmr3, DstackAppID: v.appID}
		err := CompareRegisters(res, legacyOnlyGolden, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "publishes no compose/rootfs digests")
	})

	t.Run("dstack VM with wrong app-id mismatches", func(t *testing.T) {
		res := &AttestationResult{Type: TEETypeTDX, RTMR3: v.rtmr3, DstackAppID: dstackVectors[1].appID}
		err := CompareRegisters(res, fullGolden, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "scheme=dstack")
	})

	t.Run("dstack-measured quote with no reported app-id mismatches", func(t *testing.T) {
		res := &AttestationResult{Type: TEETypeTDX, RTMR3: v.rtmr3}
		err := CompareRegisters(res, fullGolden, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reported no dstack_app_id")
	})

	t.Run("manifest with digests but no legacy rtmr3 still verifies dstack VM", func(t *testing.T) {
		g := &GoldenValues{TDXComposeSHA256: dstackFixtureComposeSHA256, TDXRootfsSHA256: v.rootfs}
		res := &AttestationResult{Type: TEETypeTDX, RTMR3: v.rtmr3, DstackAppID: v.appID}
		assert.NoError(t, CompareRegisters(res, g, nil))
	})

	t.Run("manifest with nothing for TDX fails", func(t *testing.T) {
		res := &AttestationResult{Type: TEETypeTDX, RTMR3: v.rtmr3, DstackAppID: v.appID}
		err := CompareRegisters(res, &GoldenValues{Measurement: "sev-only"}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "publish no RTMR3 for TDX")
	})
}

func TestGoldenValues_ExpectedTDXRTMR3(t *testing.T) {
	v := dstackVectors[1]
	g := &GoldenValues{
		RTMR3:            "legacy",
		TDXComposeSHA256: dstackFixtureComposeSHA256,
		TDXRootfsSHA256:  v.rootfs,
	}
	got, scheme := g.ExpectedTDXRTMR3("")
	assert.Equal(t, "legacy", got)
	assert.Equal(t, RTMR3SchemeAttestTool, scheme)

	got, scheme = g.ExpectedTDXRTMR3(v.appID)
	assert.Equal(t, v.rtmr3, got)
	assert.Equal(t, RTMR3SchemeDstack, scheme)

	got, scheme = (&GoldenValues{RTMR3: "legacy"}).ExpectedTDXRTMR3(v.appID)
	assert.Equal(t, "", got)
	assert.Equal(t, RTMR3SchemeDstack, scheme)
}

func TestGoldenFromPredicate_TDXDigests(t *testing.T) {
	t.Run("intel_tdx carries digests", func(t *testing.T) {
		p := &TEEPredicate{
			ComposeSHA256: "sha256:ignored-when-block-has-it",
			Measurements: TEEMeasurements{IntelTDX: &TDXMeasurements{
				RTMR3:         "abc",
				RTMR3Scheme:   "attest-tool",
				ComposeSHA256: "sha256:" + dstackFixtureComposeSHA256,
				RootfsSHA256:  "0xBFDE9932EB037397FAD42F50677C5D1045287369CE3B768560ACC0C1A5224F4F",
			}},
		}
		g := goldenFromPredicate(p)
		assert.Equal(t, "abc", g.RTMR3)
		assert.Equal(t, dstackFixtureComposeSHA256, g.TDXComposeSHA256)
		assert.Equal(t, "bfde9932eb037397fad42f50677c5d1045287369ce3b768560acc0c1a5224f4f", g.TDXRootfsSHA256)
		assert.True(t, g.SupportsDstackRTMR3())
	})

	t.Run("older manifest: top-level compose + intel_tdx rootfs", func(t *testing.T) {
		p := &TEEPredicate{
			ComposeSHA256: "sha256:" + dstackFixtureComposeSHA256,
			Measurements: TEEMeasurements{IntelTDX: &TDXMeasurements{
				RTMR3:        "abc",
				RootfsSHA256: dstackVectors[0].rootfs,
			}},
		}
		g := goldenFromPredicate(p)
		assert.Equal(t, dstackFixtureComposeSHA256, g.TDXComposeSHA256)
		assert.True(t, g.SupportsDstackRTMR3())
	})

	t.Run("legacy manifest without digests", func(t *testing.T) {
		p := &TEEPredicate{Measurements: TEEMeasurements{IntelTDX: &TDXMeasurements{RTMR3: "abc"}}}
		g := goldenFromPredicate(p)
		assert.Equal(t, "abc", g.RTMR3)
		assert.False(t, g.SupportsDstackRTMR3())
	})
}
