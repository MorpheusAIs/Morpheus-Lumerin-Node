package attestation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/verify"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

const (
	defaultImageRepo      = "ghcr.io/morpheusais/morpheus-lumerin-node-tee"
	defaultOIDCIssuer     = "https://token.actions.githubusercontent.com"
	defaultIdentityRegexp = "MorpheusAIs/Morpheus-Lumerin-Node"
	teePredicateType      = "https://morpheusais.github.io/tee-attestation/v1"
	cacheTTL              = 10 * time.Minute

	sigstoreBundleMediaType = "application/vnd.dev.sigstore.bundle.v0.3+json"
)

// InTotoStatement is the in-toto attestation statement envelope.
type InTotoStatement struct {
	Type          string          `json:"_type"`
	PredicateType string          `json:"predicateType"`
	Subject       json.RawMessage `json:"subject"`
	Predicate     TEEPredicate    `json:"predicate"`
}

// TEEPredicate holds the TEE attestation predicate signed by CI/CD.
type TEEPredicate struct {
	TEEImage       string          `json:"tee_image,omitempty"`
	TEEImageDigest string          `json:"tee_image_digest,omitempty"`
	ComposeSHA256  string          `json:"compose_sha256,omitempty"`
	Measurements   TEEMeasurements `json:"measurements"`
	Build          json.RawMessage `json:"build,omitempty"`
	BakedEnv       json.RawMessage `json:"baked_env,omitempty"`
}

type TEEMeasurements struct {
	IntelTDX *TDXMeasurements `json:"intel_tdx,omitempty"`
	AMDSEV   *SEVMeasurements `json:"amd_sev,omitempty"`
}

type TDXMeasurements struct {
	RTMR3           string `json:"rtmr3"`
	SecretVMRelease string `json:"secretvm_release,omitempty"`
}

type SEVMeasurements struct {
	Measurement string `json:"measurement,omitempty"`
}

// GoldenValues contains the expected TEE register values extracted from a
// cosign-verified attestation manifest attached to the GHCR image.
type GoldenValues struct {
	MRTD  string
	RTMR0 string
	RTMR1 string
	RTMR2 string
	RTMR3 string

	// AMD SEV-SNP
	Measurement string
}

// DSSEEnvelope represents a Dead Simple Signing Envelope.
type DSSEEnvelope struct {
	PayloadType string `json:"payloadType"`
	Payload     string `json:"payload"`
}

type cachedEntry struct {
	values    *GoldenValues
	fetchedAt time.Time
}

// GoldenSource verifies cosign attestation manifests from GHCR to extract
// CI/CD-signed golden register values. Attestations are stored as Sigstore
// Bundles in OCI referrers and verified using sigstore-go against the
// Sigstore public good trusted root (Fulcio CA + Rekor transparency log).
type GoldenSource struct {
	imageRepo      string
	oidcIssuer     string
	identityRegexp string
	log            lib.ILogger

	mu    sync.RWMutex
	cache map[string]*cachedEntry
}

func NewGoldenSource(imageRepo string, log lib.ILogger) *GoldenSource {
	if imageRepo == "" {
		imageRepo = defaultImageRepo
	}

	return &GoldenSource{
		imageRepo:      imageRepo,
		oidcIssuer:     defaultOIDCIssuer,
		identityRegexp: defaultIdentityRegexp,
		log:            log,
		cache:          make(map[string]*cachedEntry),
	}
}

// FetchGoldenValues retrieves the expected TEE register values for a release
// by verifying the cosign attestation attached to the GHCR image.
// Results are cached in-memory with a TTL.
func (g *GoldenSource) FetchGoldenValues(ctx context.Context, version string) (*GoldenValues, error) {
	cacheKey := version

	g.mu.RLock()
	entry, ok := g.cache[cacheKey]
	g.mu.RUnlock()

	if ok && time.Since(entry.fetchedAt) < cacheTTL {
		return entry.values, nil
	}

	values, err := g.verifyAndExtract(ctx, version)
	if err != nil {
		if ok {
			g.log.Warnf("failed to refresh golden values for %s, using stale: %s", version, err)
			return entry.values, nil
		}
		return nil, err
	}

	g.mu.Lock()
	g.cache[cacheKey] = &cachedEntry{
		values:    values,
		fetchedAt: time.Now(),
	}
	g.mu.Unlock()

	return values, nil
}

// verifyAndExtract discovers OCI referrers for the image, fetches Sigstore
// bundles, verifies them cryptographically against the Sigstore trusted root,
// and extracts TEE measurement values from the in-toto predicate.
func (g *GoldenSource) verifyAndExtract(ctx context.Context, version string) (*GoldenValues, error) {
	imageRef := fmt.Sprintf("%s:%s", g.imageRepo, version)
	g.log.Infof("verifying cosign attestation for %s", imageRef)

	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return nil, fmt.Errorf("invalid image reference %s: %w", imageRef, err)
	}

	desc, err := remote.Get(ref, remote.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve image %s: %w", imageRef, err)
	}
	digest := desc.Digest

	// OCI referrers are discovered via tag-based fallback (sha256-<hex>).
	// GHCR does not natively support the Referrers API, so cosign and
	// other tools use this tag schema convention.
	referrerTag := strings.Replace(digest.String(), ":", "-", 1)
	referrerRef, err := name.ParseReference(fmt.Sprintf("%s:%s", ref.Context().String(), referrerTag))
	if err != nil {
		return nil, fmt.Errorf("failed to construct referrer tag: %w", err)
	}

	referrerIndex, err := remote.Index(referrerRef, remote.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("no OCI referrers found for %s: %w", imageRef, err)
	}

	indexManifest, err := referrerIndex.IndexManifest()
	if err != nil {
		return nil, fmt.Errorf("failed to read referrer index manifest: %w", err)
	}

	trustedRoot, err := root.FetchTrustedRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Sigstore trusted root: %w", err)
	}

	certID, err := verify.NewShortCertificateIdentity(g.oidcIssuer, "", "", g.identityRegexp)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate identity: %w", err)
	}

	sev, err := verify.NewSignedEntityVerifier(
		trustedRoot,
		verify.WithSignedCertificateTimestamps(1),
		verify.WithTransparencyLog(1),
		verify.WithObserverTimestamps(1),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create verifier: %w", err)
	}

	for _, refDesc := range indexManifest.Manifests {
		values, err := g.processReferrer(ctx, ref.Context(), refDesc.Digest, sev, certID)
		if err != nil {
			continue
		}
		if values != nil {
			return values, nil
		}
	}

	return nil, fmt.Errorf("no TEE attestation manifest (predicate %s) found for %s", teePredicateType, imageRef)
}

// processReferrer fetches a single OCI referrer manifest, parses its Sigstore
// bundle layer, verifies the bundle, and extracts golden values if the
// predicate matches our TEE attestation type.
func (g *GoldenSource) processReferrer(
	ctx context.Context,
	repo name.Repository,
	dgst v1.Hash,
	sev *verify.Verifier,
	certID verify.CertificateIdentity,
) (*GoldenValues, error) {
	imgRef := repo.Digest(dgst.String())
	img, err := remote.Image(imgRef, remote.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}

	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("layers: %w", err)
	}

	for _, layer := range layers {
		mt, _ := layer.MediaType()
		if string(mt) != sigstoreBundleMediaType {
			continue
		}

		values, err := g.verifyBundleLayer(layer, sev, certID)
		if err != nil {
			return nil, err
		}
		if values != nil {
			return values, nil
		}
	}

	return nil, fmt.Errorf("no sigstore bundle layer found")
}

func (g *GoldenSource) verifyBundleLayer(
	layer v1.Layer,
	sev *verify.Verifier,
	certID verify.CertificateIdentity,
) (*GoldenValues, error) {
	reader, err := layer.Uncompressed()
	if err != nil {
		return nil, fmt.Errorf("uncompress: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	var b bundle.Bundle
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("parse bundle: %w", err)
	}

	_, err = sev.Verify(&b, verify.NewPolicy(
		verify.WithoutArtifactUnsafe(),
		verify.WithCertificateIdentity(certID),
	))
	if err != nil {
		return nil, fmt.Errorf("verify: %w", err)
	}

	envelope, err := b.Envelope()
	if err != nil {
		return nil, fmt.Errorf("envelope: %w", err)
	}

	stmt, err := envelope.Statement()
	if err != nil {
		return nil, fmt.Errorf("statement: %w", err)
	}

	if stmt.GetPredicateType() != teePredicateType {
		return nil, nil
	}

	predicateJSON, err := stmt.GetPredicate().MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal predicate: %w", err)
	}

	var predicate TEEPredicate
	if err := json.Unmarshal(predicateJSON, &predicate); err != nil {
		return nil, fmt.Errorf("parse predicate: %w", err)
	}

	values := &GoldenValues{}
	if m := predicate.Measurements.IntelTDX; m != nil {
		values.RTMR3 = m.RTMR3
	}
	if m := predicate.Measurements.AMDSEV; m != nil {
		values.Measurement = m.Measurement
	}

	g.log.Infof("extracted golden values from verified attestation")
	return values, nil
}

// parseAttestationPayload extracts GoldenValues from a cosign attestation payload.
// The payload may be a DSSE envelope (containing a base64-encoded in-toto statement)
// or a direct in-toto statement.
func parseAttestationPayload(payload []byte) (*GoldenValues, error) {
	var statementBytes []byte

	var envelope DSSEEnvelope
	if err := json.Unmarshal(payload, &envelope); err == nil && envelope.Payload != "" {
		decoded, err := base64.StdEncoding.DecodeString(envelope.Payload)
		if err != nil {
			decoded, err = base64.RawURLEncoding.DecodeString(envelope.Payload)
			if err != nil {
				return nil, fmt.Errorf("failed to decode DSSE payload: %w", err)
			}
		}
		statementBytes = decoded
	} else {
		statementBytes = payload
	}

	var statement InTotoStatement
	if err := json.Unmarshal(statementBytes, &statement); err != nil {
		return nil, fmt.Errorf("failed to parse in-toto statement: %w", err)
	}

	if statement.PredicateType != teePredicateType {
		return nil, fmt.Errorf("wrong predicate type: got %s, want %s", statement.PredicateType, teePredicateType)
	}

	values := &GoldenValues{}
	if m := statement.Predicate.Measurements.IntelTDX; m != nil {
		values.RTMR3 = m.RTMR3
	}
	if m := statement.Predicate.Measurements.AMDSEV; m != nil {
		values.Measurement = m.Measurement
	}

	return values, nil
}
