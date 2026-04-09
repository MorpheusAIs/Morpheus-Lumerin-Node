package attestation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/verify"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Unit tests for parseAttestationPayload (no network, fast)
// ---------------------------------------------------------------------------

func TestParseAttestationPayload_DirectStatement(t *testing.T) {
	statement := InTotoStatement{
		Type:          "https://in-toto.io/Statement/v0.1",
		PredicateType: teePredicateType,
		Predicate: TEEPredicate{
			TEEImage:       "ghcr.io/morpheusais/morpheus-lumerin-node-tee@sha256:abc123",
			TEEImageDigest: "sha256:abc123",
			ComposeSHA256:  "def456",
			Measurements: TEEMeasurements{
				IntelTDX: &TDXMeasurements{
					RTMR3:           "aabbccdd00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabb",
					SecretVMRelease: "v0.0.25",
				},
			},
		},
	}

	payload, err := json.Marshal(statement)
	require.NoError(t, err)

	values, err := parseAttestationPayload(payload)
	require.NoError(t, err)
	assert.Equal(t, "aabbccdd00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabb", values.RTMR3)
	assert.Empty(t, values.Measurement)
}

func TestParseAttestationPayload_DSSEEnvelope(t *testing.T) {
	statement := InTotoStatement{
		Type:          "https://in-toto.io/Statement/v0.1",
		PredicateType: teePredicateType,
		Predicate: TEEPredicate{
			Measurements: TEEMeasurements{
				IntelTDX: &TDXMeasurements{
					RTMR3: "deadbeef",
				},
			},
		},
	}

	statementBytes, err := json.Marshal(statement)
	require.NoError(t, err)

	envelope := DSSEEnvelope{
		PayloadType: "application/vnd.in-toto+json",
		Payload:     base64.StdEncoding.EncodeToString(statementBytes),
	}

	payload, err := json.Marshal(envelope)
	require.NoError(t, err)

	values, err := parseAttestationPayload(payload)
	require.NoError(t, err)
	assert.Equal(t, "deadbeef", values.RTMR3)
}

func TestParseAttestationPayload_DSSEEnvelope_RawURLEncoding(t *testing.T) {
	statement := InTotoStatement{
		Type:          "https://in-toto.io/Statement/v0.1",
		PredicateType: teePredicateType,
		Predicate: TEEPredicate{
			Measurements: TEEMeasurements{
				AMDSEV: &SEVMeasurements{
					Measurement: "sev-measurement-hash",
				},
			},
		},
	}

	statementBytes, err := json.Marshal(statement)
	require.NoError(t, err)

	envelope := DSSEEnvelope{
		PayloadType: "application/vnd.in-toto+json",
		Payload:     base64.RawURLEncoding.EncodeToString(statementBytes),
	}

	payload, err := json.Marshal(envelope)
	require.NoError(t, err)

	values, err := parseAttestationPayload(payload)
	require.NoError(t, err)
	assert.Empty(t, values.RTMR3)
	assert.Equal(t, "sev-measurement-hash", values.Measurement)
}

func TestParseAttestationPayload_WrongPredicateType(t *testing.T) {
	statement := InTotoStatement{
		Type:          "https://in-toto.io/Statement/v0.1",
		PredicateType: "https://slsa.dev/provenance/v0.2",
		Predicate:     TEEPredicate{},
	}

	payload, err := json.Marshal(statement)
	require.NoError(t, err)

	_, err = parseAttestationPayload(payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wrong predicate type")
}

func TestParseAttestationPayload_InvalidJSON(t *testing.T) {
	_, err := parseAttestationPayload([]byte("not json"))
	assert.Error(t, err)
}

func TestParseAttestationPayload_EmptyMeasurements(t *testing.T) {
	statement := InTotoStatement{
		Type:          "https://in-toto.io/Statement/v0.1",
		PredicateType: teePredicateType,
		Predicate: TEEPredicate{
			Measurements: TEEMeasurements{},
		},
	}

	payload, err := json.Marshal(statement)
	require.NoError(t, err)

	values, err := parseAttestationPayload(payload)
	require.NoError(t, err)
	assert.Empty(t, values.RTMR3)
	assert.Empty(t, values.Measurement)
}

// ---------------------------------------------------------------------------
// Unit tests for GoldenSource caching
// ---------------------------------------------------------------------------

func TestGoldenSource_CacheHit(t *testing.T) {
	log := lib.NewTestLogger()
	gs := NewGoldenSource("ghcr.io/example/test", log)

	expected := &GoldenValues{RTMR3: "cached-value"}
	gs.cache["v1.0.0"] = &cachedEntry{
		values:    expected,
		fetchedAt: time.Now(),
	}

	values, err := gs.FetchGoldenValues(context.Background(), "v1.0.0")
	require.NoError(t, err)
	assert.Equal(t, "cached-value", values.RTMR3)
}

func TestGoldenSource_CacheExpired_NoNetwork(t *testing.T) {
	log := lib.NewTestLogger()
	gs := NewGoldenSource("ghcr.io/nonexistent/image", log)

	stale := &GoldenValues{RTMR3: "stale-value"}
	gs.cache["v1.0.0"] = &cachedEntry{
		values:    stale,
		fetchedAt: time.Now().Add(-1 * time.Hour),
	}

	values, err := gs.FetchGoldenValues(context.Background(), "v1.0.0")
	require.NoError(t, err)
	assert.Equal(t, "stale-value", values.RTMR3, "should fall back to stale cache on network error")
}

func TestGoldenSource_DefaultImageRepo(t *testing.T) {
	log := lib.NewTestLogger()
	gs := NewGoldenSource("", log)
	assert.Equal(t, defaultImageRepo, gs.imageRepo)
}

// ---------------------------------------------------------------------------
// Integration test: verify real cosign attestation from GHCR
//
// Run with:
//   TEE_TEST_IMAGE=ghcr.io/morpheusais/morpheus-lumerin-node-tee \
//   TEE_TEST_VERSION=v5.14.6 \
//   go test -v -run TestGoldenSource_RealImage -count=1 ./internal/attestation/
//
// Set TEE_TEST_EXPECTED_RTMR3 to also assert the extracted value.
// ---------------------------------------------------------------------------

func TestGoldenSource_RealImage(t *testing.T) {
	imageRepo := os.Getenv("TEE_TEST_IMAGE")
	version := os.Getenv("TEE_TEST_VERSION")
	if imageRepo == "" || version == "" {
		t.Skip("TEE_TEST_IMAGE and TEE_TEST_VERSION not set; skipping real-image integration test")
	}

	log := lib.NewTestLogger()
	gs := NewGoldenSource(imageRepo, log)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	values, err := gs.FetchGoldenValues(ctx, version)
	require.NoError(t, err, "cosign attestation verification should succeed")

	t.Logf("--- Golden values from %s:%s ---", imageRepo, version)
	t.Logf("RTMR3:       %s", values.RTMR3)
	t.Logf("MRTD:        %s", values.MRTD)
	t.Logf("RTMR0:       %s", values.RTMR0)
	t.Logf("RTMR1:       %s", values.RTMR1)
	t.Logf("RTMR2:       %s", values.RTMR2)
	t.Logf("Measurement: %s", values.Measurement)

	if expectedRTMR3 := os.Getenv("TEE_TEST_EXPECTED_RTMR3"); expectedRTMR3 != "" {
		assert.Equal(t, expectedRTMR3, values.RTMR3, "RTMR3 should match expected value")
	}

	// Verify caching works on the second call
	values2, err := gs.FetchGoldenValues(ctx, version)
	require.NoError(t, err)
	assert.Equal(t, values.RTMR3, values2.RTMR3, "cached result should match")
}

// ---------------------------------------------------------------------------
// Integration test: verify real image and compare against live provider quote
//
// Run with:
//   TEE_TEST_IMAGE=ghcr.io/morpheusais/morpheus-lumerin-node-tee \
//   TEE_TEST_VERSION=v5.14.6 \
//   TEE_TEST_PROVIDER_URL=morpheus.example.com:3333 \
//   TEE_PORTAL_URL=https://secretai.scrtlabs.com/api/parse-quote \
//   go test -v -run TestFullVerification -count=1 ./internal/attestation/
// ---------------------------------------------------------------------------

func TestFullVerification(t *testing.T) {
	imageRepo := os.Getenv("TEE_TEST_IMAGE")
	version := os.Getenv("TEE_TEST_VERSION")
	providerURL := os.Getenv("TEE_TEST_PROVIDER_URL")
	if imageRepo == "" || version == "" || providerURL == "" {
		t.Skip("TEE_TEST_IMAGE, TEE_TEST_VERSION, and TEE_TEST_PROVIDER_URL not set; skipping full verification test")
	}

	portalURL := os.Getenv("TEE_PORTAL_URL")
	log := lib.NewTestLogger()
	verifier := NewVerifier(portalURL, imageRepo, log)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	err := verifier.VerifyProvider(ctx, providerURL, version)
	if err != nil {
		t.Fatalf("full TEE verification failed: %s", err)
	}
	t.Logf("Full TEE verification PASSED for provider %s (version %s)", providerURL, version)
}

// ---------------------------------------------------------------------------
// Standalone: fetch and print attestation predicate (like cosign verify-attestation | jq)
//
// Run with:
//   TEE_TEST_IMAGE=ghcr.io/morpheusais/morpheus-lumerin-node-tee \
//   TEE_TEST_VERSION=v5.14.6 \
//   go test -v -run TestDumpAttestationPredicate -count=1 ./internal/attestation/
// ---------------------------------------------------------------------------

func TestDumpAttestationPredicate(t *testing.T) {
	imageRepo := os.Getenv("TEE_TEST_IMAGE")
	version := os.Getenv("TEE_TEST_VERSION")
	if imageRepo == "" || version == "" {
		t.Skip("TEE_TEST_IMAGE and TEE_TEST_VERSION not set; skipping attestation dump test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	imageRef := fmt.Sprintf("%s:%s", imageRepo, version)
	t.Logf("Fetching and verifying cosign attestations for %s ...", imageRef)

	ref, err := name.ParseReference(imageRef)
	require.NoError(t, err)

	desc, err := remote.Get(ref, remote.WithContext(ctx))
	require.NoError(t, err)

	referrerTag := strings.Replace(desc.Digest.String(), ":", "-", 1)
	referrerRef, err := name.ParseReference(fmt.Sprintf("%s:%s", ref.Context().String(), referrerTag))
	require.NoError(t, err)

	referrerIndex, err := remote.Index(referrerRef, remote.WithContext(ctx))
	require.NoError(t, err)

	indexManifest, err := referrerIndex.IndexManifest()
	require.NoError(t, err)

	trustedRoot, err := root.FetchTrustedRoot()
	require.NoError(t, err)

	certID, err := verify.NewShortCertificateIdentity(defaultOIDCIssuer, "", "", defaultIdentityRegexp)
	require.NoError(t, err)

	sev, err := verify.NewSignedEntityVerifier(
		trustedRoot,
		verify.WithSignedCertificateTimestamps(1),
		verify.WithTransparencyLog(1),
		verify.WithObserverTimestamps(1),
	)
	require.NoError(t, err)

	t.Logf("Total referrers found: %d", len(indexManifest.Manifests))

	for i, refDesc := range indexManifest.Manifests {
		t.Logf("[%d] digest=%s artifactType=%s", i, refDesc.Digest, refDesc.ArtifactType)

		imgRef := ref.Context().Digest(refDesc.Digest.String())
		img, err := remote.Image(imgRef, remote.WithContext(ctx))
		if err != nil {
			t.Logf("[%d] fetch error: %s", i, err)
			continue
		}

		dumpBundleLayers(t, i, img, sev, certID)
	}
}

func dumpBundleLayers(t *testing.T, idx int, img v1.Image, sev *verify.Verifier, certID verify.CertificateIdentity) {
	t.Helper()

	layers, err := img.Layers()
	if err != nil {
		t.Logf("[%d] layers error: %s", idx, err)
		return
	}

	for _, layer := range layers {
		mt, _ := layer.MediaType()
		if string(mt) != sigstoreBundleMediaType {
			continue
		}

		reader, err := layer.Uncompressed()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			continue
		}

		var b bundle.Bundle
		if err := json.Unmarshal(data, &b); err != nil {
			t.Logf("[%d] bundle parse error: %s", idx, err)
			continue
		}

		_, err = sev.Verify(&b, verify.NewPolicy(
			verify.WithoutArtifactUnsafe(),
			verify.WithCertificateIdentity(certID),
		))
		if err != nil {
			t.Logf("[%d] verification failed: %s", idx, err)
			continue
		}
		t.Logf("[%d] VERIFIED", idx)

		envelope, err := b.Envelope()
		if err != nil {
			continue
		}
		stmt, err := envelope.Statement()
		if err != nil {
			continue
		}

		t.Logf("[%d] predicateType: %s", idx, stmt.GetPredicateType())
		predJSON, _ := stmt.GetPredicate().MarshalJSON()
		var pretty json.RawMessage
		if json.Unmarshal(predJSON, &pretty) == nil {
			out, _ := json.MarshalIndent(pretty, "", "  ")
			t.Logf("[%d] predicate:\n%s", idx, string(out))
		}
	}
}

