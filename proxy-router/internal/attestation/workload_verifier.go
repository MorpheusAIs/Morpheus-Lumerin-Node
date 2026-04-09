package attestation

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

type WorkloadStatus string

const (
	WorkloadAuthentic         WorkloadStatus = "authentic_match"
	WorkloadAuthenticMismatch WorkloadStatus = "authentic_mismatch"
	WorkloadNotAuthentic      WorkloadStatus = "not_authentic"
)

type WorkloadResult struct {
	Status       WorkloadStatus
	TemplateName string
	VMType       string
	ArtifactsVer string
	Env          string
}

func VerifyTdxWorkload(registry *ArtifactRegistry, cpuQuoteHex string, dockerComposeYaml string, log lib.ILogger) WorkloadResult {
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

	composeBytes := []byte(dockerComposeYaml)
	composeHash := sha256.Sum256(composeBytes)
	if log != nil {
		log.Infof("workload: compose size=%d bytes, sha256=%s, quote RTMR3=%s, candidates=%d",
			len(composeBytes), hex.EncodeToString(composeHash[:]), fields.RTMR3, len(candidates))
	}

	for i, entry := range candidates {
		expected := CalculateRTMR3(composeBytes, entry.RootfsData)
		if log != nil {
			log.Infof("workload: candidate[%d] template=%s ver=%s rootfs=%s calculated_rtmr3=%s match=%v",
				i, entry.TemplateName, entry.ArtifactsVer, entry.RootfsData, expected, expected == fields.RTMR3)
		}
		if expected == fields.RTMR3 {
			return WorkloadResult{
				Status:       WorkloadAuthentic,
				TemplateName: entry.TemplateName,
				VMType:       entry.VMType,
				ArtifactsVer: entry.ArtifactsVer,
				Env:          entry.VMType,
			}
		}
	}

	return WorkloadResult{
		Status:       WorkloadAuthenticMismatch,
		TemplateName: best.TemplateName,
		VMType:       best.VMType,
		ArtifactsVer: best.ArtifactsVer,
		Env:          best.VMType,
	}
}

func VerifyWorkload(registry *ArtifactRegistry, cpuQuoteData string, dockerComposeYaml string, log lib.ILogger) WorkloadResult {
	if IsTdxQuote(cpuQuoteData) {
		return VerifyTdxWorkload(registry, cpuQuoteData, dockerComposeYaml, log)
	}
	return WorkloadResult{Status: WorkloadNotAuthentic}
}
