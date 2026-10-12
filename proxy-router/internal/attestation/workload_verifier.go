package attestation

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

type WorkloadStatus string

const (
	WorkloadAuthentic            WorkloadStatus = "authentic_match"
	WorkloadAuthenticMismatch    WorkloadStatus = "authentic_mismatch"
	WorkloadNotAuthentic         WorkloadStatus = "not_authentic"
	ArtifactRegistryNotAvailable WorkloadStatus = "artifact_registry_not_available"
)

type WorkloadResult struct {
	Status       WorkloadStatus
	TemplateName string
	VMType       string
	ArtifactsVer string
	Env          string
	// RTMR3Scheme is the replay scheme that matched (TDX only).
	RTMR3Scheme RTMR3Scheme
}

// composeCandidates returns the docker-compose byte interpretations to try
// when replaying the measurement. Old attest-rest wraps the /docker-compose
// response in an HTML page (a <pre> block); newer attest-rest serves the raw
// file bytes. The measurement (RTMR3 on TDX, docker_compose_hash on SEV) is
// always over the original file, so both the raw response and the
// HTML-extracted content are tried and a match on either is accepted
// (mirrors scrtlabs/secretvm-verify v0.12.0).
func composeCandidates(compose string) []string {
	extracted := extractPreContent(compose)
	if extracted != compose {
		return []string{compose, extracted}
	}
	return []string{compose}
}

// VerifyTdxWorkload replays RTMR3 from the compose and the registry rootfs for
// every artifact entry matching the quote's MRTD/RTMR0-2. dstackAppID is the
// VM's self-reported app-id from GET /info ("" for pre-dstack images); it
// selects the replay scheme (see rtmr.go). Mirrors secretvm-verify 0.13.0.
func VerifyTdxWorkload(registry *ArtifactRegistry, cpuQuoteHex string, dockerComposeYaml string, dstackAppID string, log lib.ILogger) WorkloadResult {
	fields, err := ParseTdxQuoteFields(cpuQuoteHex)
	if err != nil {
		if log != nil {
			log.Warnf("workload: failed to parse TDX quote: %s", err)
		}
		return WorkloadResult{Status: WorkloadNotAuthentic}
	}

	candidates := registry.FindMatchingArtifacts(fields.MRTD, fields.RTMR0, fields.RTMR1, fields.RTMR2)
	if len(candidates) == 0 {
		if log != nil {
			log.Warnf("workload: no registry entries match MRTD=%s RTMR0=%s", fields.MRTD[:16]+"...", fields.RTMR0[:16]+"...")
		}
		return WorkloadResult{Status: WorkloadNotAuthentic}
	}

	best := registry.PickNewestVersion(candidates)

	appID := NormalizeDstackAppID(dstackAppID)
	scheme := RTMR3SchemeFor(appID)
	composeVariants := composeCandidates(dockerComposeYaml)
	if log != nil {
		composeHash := sha256.Sum256([]byte(dockerComposeYaml))
		log.Infof("workload: compose size=%d bytes, sha256=%s, variants=%d, scheme=%s, app_id=%q, quote RTMR3=%s, candidates=%d",
			len(dockerComposeYaml), hex.EncodeToString(composeHash[:]), len(composeVariants), scheme, appID, fields.RTMR3, len(candidates))
	}

	for i, entry := range candidates {
		for _, variant := range composeVariants {
			expected := CalculateRTMR3WithAppID([]byte(variant), entry.RootfsData, appID)
			if log != nil {
				log.Infof("workload: candidate[%d] template=%s ver=%s rootfs=%s scheme=%s calculated_rtmr3=%s match=%v",
					i, entry.TemplateName, entry.ArtifactsVer, entry.RootfsData, scheme, expected, expected == fields.RTMR3)
			}
			if expected == fields.RTMR3 {
				return WorkloadResult{
					Status:       WorkloadAuthentic,
					TemplateName: entry.TemplateName,
					VMType:       entry.VMType,
					ArtifactsVer: entry.ArtifactsVer,
					Env:          entry.VMType,
					RTMR3Scheme:  scheme,
				}
			}
		}
	}

	return WorkloadResult{
		Status:       WorkloadAuthenticMismatch,
		TemplateName: best.TemplateName,
		VMType:       best.VMType,
		ArtifactsVer: best.ArtifactsVer,
		Env:          best.VMType,
		RTMR3Scheme:  scheme,
	}
}

// VerifyWorkload dispatches on quote type. dstackAppID only affects TDX: the
// SEV-SNP launch digest has no app-id event (dstack KMS on AMD governs key
// release, not the measurement), so the SEV path ignores it.
func VerifyWorkload(registry *ArtifactRegistry, sevRegistry *SevArtifactRegistry, cpuQuoteData string, dockerComposeYaml string, dstackAppID string, log lib.ILogger) WorkloadResult {
	if IsTdxQuote(cpuQuoteData) {
		if registry == nil || !registry.IsLoaded() {
			if log != nil {
				log.Warnf("workload: TDX quote detected but artifact registry not available; cannot verify workload")
			}
			return WorkloadResult{Status: ArtifactRegistryNotAvailable}
		}
		return VerifyTdxWorkload(registry, cpuQuoteData, dockerComposeYaml, dstackAppID, log)
	}
	if sevRegistry == nil || !sevRegistry.IsLoaded() {
		if log != nil {
			log.Warnf("workload: SEV quote detected but SEV registry not available; cannot verify workload")
		}
		return WorkloadResult{Status: ArtifactRegistryNotAvailable}
	}
	return VerifySevWorkload(sevRegistry, cpuQuoteData, dockerComposeYaml, log)
}
