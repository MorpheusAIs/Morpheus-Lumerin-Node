package attestation

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

// VerifySevWorkload verifies that an AMD SEV-SNP attestation quote was produced
// by a known SecretVM running the given docker-compose YAML.
//
// Port of verifySevWorkload from secretvm-verify/workload.ts.
func VerifySevWorkload(registry *SevArtifactRegistry, cpuQuoteBase64 string, dockerComposeYaml string, log lib.ILogger) WorkloadResult {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(cpuQuoteBase64))
	if err != nil {
		if log != nil {
			log.Warnf("sev workload: failed to base64-decode quote: %s", err)
		}
		return WorkloadResult{Status: WorkloadNotAuthentic}
	}

	// Need at least 0x090 + 48 = 192 bytes for the measurement field
	if len(raw) < 0x090+48 {
		if log != nil {
			log.Warnf("sev workload: quote too short (%d bytes)", len(raw))
		}
		return WorkloadResult{Status: WorkloadNotAuthentic}
	}

	quoteMeasurement := hex.EncodeToString(raw[0x090 : 0x090+48])

	// Parse family_id (0x010..0x020) and image_id (0x020..0x030)
	var familyIDBytes []byte
	var imageID string
	if len(raw) >= 0x030 {
		familyIDBytes = raw[0x010:0x020]
		imageID = strings.TrimRight(string(raw[0x020:0x030]), "\x00#")
	}

	composeHashArr := sha256.Sum256([]byte(dockerComposeYaml))
	composeHash := hex.EncodeToString(composeHashArr[:])

	family := ParseSevFamilyID(familyIDBytes)

	if log != nil {
		familyStr := "<nil>"
		if family != nil {
			familyStr = fmt.Sprintf("%s/%s/%d", family.VMType, family.TemplateName, family.Vcpus)
		}
		log.Infof("sev workload: measurement=%s..., family=%s, imageId=%s, composeHash=%s...",
			quoteMeasurement[:16], familyStr, imageID, composeHash[:16])
	}

	buildCmdline := func(entry *SevArtifactEntry) string {
		prefix := "console=ttyS0 loglevel=7"
		if entry.CmdlineExtra != "" {
			prefix += " " + entry.CmdlineExtra
		}
		return fmt.Sprintf("%s docker_compose_hash=%s rootfs_hash=%s", prefix, composeHash, entry.RootfsHash)
	}

	tryEntry := func(entry *SevArtifactEntry, vcpus int) bool {
		cmdline := buildCmdline(entry)
		computed := CalcSevMeasurement(entry, vcpus, cmdline)
		return computed == quoteMeasurement
	}

	if family == nil {
		// family_id not set -- brute-force all registry entries and vcpu counts
		if log != nil {
			log.Infof("sev workload: family_id not set, brute-forcing all entries x vcpu counts")
		}
		allEntries := registry.AllEntries()
		for _, entry := range allEntries {
			for templateName, vcpus := range VcpuMap {
				if tryEntry(&entry, vcpus) {
					return WorkloadResult{
						Status:       WorkloadAuthentic,
						TemplateName: templateName,
						VMType:       entry.VMType,
						ArtifactsVer: entry.ArtifactsVer,
						Env:          entry.VMType,
					}
				}
			}
		}
		return WorkloadResult{Status: WorkloadNotAuthentic}
	}

	// family_id is valid -- filter by vmType
	candidates := registry.FindByVMType(family.VMType)
	if log != nil {
		log.Infof("sev workload: found %d candidates for vm_type=%s", len(candidates), family.VMType)
	}

	// Try version-specific entries first
	var versionEntries []SevArtifactEntry
	if imageID != "" {
		for _, e := range candidates {
			if e.ArtifactsVer == imageID {
				versionEntries = append(versionEntries, e)
			}
		}
	}

	for i := range versionEntries {
		if tryEntry(&versionEntries[i], family.Vcpus) {
			return WorkloadResult{
				Status:       WorkloadAuthentic,
				TemplateName: family.TemplateName,
				VMType:       family.VMType,
				ArtifactsVer: versionEntries[i].ArtifactsVer,
				Env:          family.VMType,
			}
		}
	}

	// Fallback: try other entries for this vmType
	for i := range candidates {
		if imageID != "" && candidates[i].ArtifactsVer == imageID {
			continue // already tried above
		}
		if tryEntry(&candidates[i], family.Vcpus) {
			return WorkloadResult{
				Status:       WorkloadAuthentic,
				TemplateName: family.TemplateName,
				VMType:       family.VMType,
				ArtifactsVer: candidates[i].ArtifactsVer,
				Env:          family.VMType,
			}
		}
	}

	// No compose match. If version entries exist, the VM is authentic but
	// the provided compose doesn't match the measurement.
	if len(versionEntries) > 0 {
		return WorkloadResult{
			Status:       WorkloadAuthenticMismatch,
			TemplateName: family.TemplateName,
			VMType:       family.VMType,
			ArtifactsVer: imageID,
			Env:          family.VMType,
		}
	}

	return WorkloadResult{Status: WorkloadNotAuthentic}
}
