package attestation

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"strings"
)

const (
	ldSize       = 48 // SHA-384 digest size
	vmsaGPA      = uint64(0xFFFFFFFFF000)
	vcpuSigEPYC  = 0x00800f12
	guestFeatures = 0x1
	bspEIP       = 0xfffffff0
)

// VcpuMap maps SecretVM template names to vCPU counts.
var VcpuMap = map[string]int{
	"small":   1,
	"medium":  2,
	"large":   4,
	"2xlarge": 8,
	"4xlarge": 16,
}

// SevOvmfSection represents one OVMF section in the SEV registry.
type SevOvmfSection struct {
	GPA         uint64 `json:"gpa"`
	Size        int    `json:"size"`
	SectionType int    `json:"section_type"`
}

// SevArtifactEntry represents one entry in the SEV artifact registry JSON.
type SevArtifactEntry struct {
	VMType            string           `json:"vm_type"`
	ArtifactsVer      string           `json:"artifacts_ver"`
	KernelHash        string           `json:"kernel_hash"`
	InitrdHash        string           `json:"initrd_hash"`
	VcpuType          string           `json:"vcpu_type"`
	RootfsHash        string           `json:"rootfs_hash"`
	OvmfHash          string           `json:"ovmf_hash"`
	SevHashesTableGPA uint64           `json:"sev_hashes_table_gpa"`
	SevEsResetEIP     uint32           `json:"sev_es_reset_eip"`
	OvmfSections      []SevOvmfSection `json:"ovmf_sections"`
	CmdlineExtra      string           `json:"cmdline_extra,omitempty"`
}

// SevFamilyID holds parsed SEV family_id fields from the attestation report.
type SevFamilyID struct {
	VMType       string
	TemplateName string
	Vcpus        int
}

// ParseSevFamilyID parses the 16-byte family_id field from a SEV-SNP report.
// Expects format "<vmType>-<templateName>-sev" (e.g. "prod-small-sev").
func ParseSevFamilyID(familyIDBytes []byte) *SevFamilyID {
	s := strings.TrimRight(string(familyIDBytes[:16]), "\x00#")
	if !strings.HasSuffix(s, "-sev") {
		return nil
	}
	core := s[:len(s)-4] // strip "-sev"
	idx := strings.Index(core, "-")
	if idx < 0 {
		return nil
	}
	vmType := core[:idx]
	templateName := core[idx+1:]
	vcpus, ok := VcpuMap[templateName]
	if !ok {
		return nil
	}
	return &SevFamilyID{VMType: vmType, TemplateName: templateName, Vcpus: vcpus}
}

// gctxUpdate performs the core GCTX page-update: SHA-384 of the 0x70-byte PAGE_INFO struct.
// See AMD SNP spec Section 8.17.2 Table 67.
func gctxUpdate(ld []byte, pageType byte, gpa uint64, contents []byte) []byte {
	var buf [0x70]byte
	copy(buf[0:48], ld)
	copy(buf[48:96], contents)
	binary.LittleEndian.PutUint16(buf[96:98], 0x70) // page_info_len
	buf[98] = pageType
	// buf[99..103] = 0 (imi, vmpl perms, reserved) -- already zero
	binary.LittleEndian.PutUint64(buf[104:112], gpa)
	h := sha512.Sum384(buf[:])
	return h[:]
}

func sha384(data []byte) []byte {
	h := sha512.Sum384(data)
	return h[:]
}

func gctxUpdateNormalPages(ld []byte, startGPA uint64, data []byte) []byte {
	for offset := 0; offset < len(data); offset += 4096 {
		end := offset + 4096
		if end > len(data) {
			end = len(data)
		}
		page := data[offset:end]
		ld = gctxUpdate(ld, 0x01, startGPA+uint64(offset), sha384(page))
	}
	return ld
}

func gctxUpdateVmsaPage(ld []byte, data []byte) []byte {
	return gctxUpdate(ld, 0x02, vmsaGPA, sha384(data))
}

func gctxUpdateZeroPages(ld []byte, gpa uint64, size int) []byte {
	zeros := make([]byte, ldSize)
	for offset := 0; offset < size; offset += 4096 {
		ld = gctxUpdate(ld, 0x03, gpa+uint64(offset), zeros)
	}
	return ld
}

func gctxUpdateSecretsPage(ld []byte, gpa uint64) []byte {
	zeros := make([]byte, ldSize)
	return gctxUpdate(ld, 0x05, gpa, zeros)
}

func gctxUpdateCpuidPage(ld []byte, gpa uint64) []byte {
	zeros := make([]byte, ldSize)
	return gctxUpdate(ld, 0x06, gpa, zeros)
}

// uuidToLE converts a UUID string to little-endian byte encoding per RFC 4122.
func uuidToLE(guid string) []byte {
	hexStr := strings.ReplaceAll(guid, "-", "")
	b, _ := hex.DecodeString(hexStr)
	le := make([]byte, len(b))
	copy(le, b)
	// group1: bytes 0-3 (swap)
	le[0], le[1], le[2], le[3] = b[3], b[2], b[1], b[0]
	// group2: bytes 4-5 (swap)
	le[4], le[5] = b[5], b[4]
	// group3: bytes 6-7 (swap)
	le[6], le[7] = b[7], b[6]
	// groups 4+5 remain big-endian
	return le
}

const (
	sevHashTableHeaderGUID = "9438d606-4f22-4cc9-b479-a793d411fd21"
	sevKernelEntryGUID     = "4de79437-abd2-427f-b835-d5b172d2045b"
	sevInitrdEntryGUID     = "44baf731-3a2f-4bd7-9af1-41e29169781d"
	sevCmdlineEntryGUID    = "97d02dd8-bd20-4c94-aa78-e7714d36ab2a"
)

// sevHashTableEntry builds a 50-byte SevHashTableEntry: guid(16) + length(u16 LE) + hash(32).
func sevHashTableEntry(guidStr string, hash []byte) []byte {
	entry := make([]byte, 50)
	copy(entry[0:16], uuidToLE(guidStr))
	binary.LittleEndian.PutUint16(entry[16:18], 50)
	copy(entry[18:50], hash)
	return entry
}

// buildHashesPage constructs the QEMU SEV kernel hashes page.
func buildHashesPage(kernelHashHex, initrdHashHex, cmdline string, offsetInPage int) []byte {
	kernelHash, _ := hex.DecodeString(kernelHashHex)

	var initrdHash []byte
	if initrdHashHex != "" {
		initrdHash, _ = hex.DecodeString(initrdHashHex)
	} else {
		h := sha256.Sum256(nil)
		initrdHash = h[:]
	}

	var cmdlineBytes []byte
	if cmdline != "" {
		cmdlineBytes = append([]byte(cmdline), 0)
	} else {
		cmdlineBytes = []byte{0}
	}
	cmdlineHashArr := sha256.Sum256(cmdlineBytes)
	cmdlineHash := cmdlineHashArr[:]

	// SevHashTable: guid(16) + length(u16) + cmdline(50) + initrd(50) + kernel(50) = 168 bytes
	ht := make([]byte, 168)
	copy(ht[0:16], uuidToLE(sevHashTableHeaderGUID))
	binary.LittleEndian.PutUint16(ht[16:18], 168)
	copy(ht[18:68], sevHashTableEntry(sevCmdlineEntryGUID, cmdlineHash))
	copy(ht[68:118], sevHashTableEntry(sevInitrdEntryGUID, initrdHash))
	copy(ht[118:168], sevHashTableEntry(sevKernelEntryGUID, kernelHash))

	// Pad to 16-byte alignment: 168 % 16 = 8 -> 8 padding bytes -> 176 bytes
	padded := make([]byte, 176)
	copy(padded, ht)

	page := make([]byte, 4096)
	copy(page[offsetInPage:], padded)
	return page
}

// buildVmsaPage constructs a VMSA page for QEMU SEV-SNP mode.
func buildVmsaPage(eip uint32, vcpuSig uint32, guestFeat uint64) []byte {
	page := make([]byte, 4096)

	vmcbSeg := func(off int, sel uint16, attr uint16, lim uint32, base uint64) {
		binary.LittleEndian.PutUint16(page[off:], sel)
		binary.LittleEndian.PutUint16(page[off+2:], attr)
		binary.LittleEndian.PutUint32(page[off+4:], lim)
		binary.LittleEndian.PutUint64(page[off+8:], base)
	}

	csBase := uint64(eip & 0xffff0000)
	rip := uint64(eip & 0x0000ffff)

	vmcbSeg(0x000, 0, 0x0093, 0xffff, 0)       // es
	vmcbSeg(0x010, 0xf000, 0x009b, 0xffff, csBase) // cs
	vmcbSeg(0x020, 0, 0x0093, 0xffff, 0)       // ss
	vmcbSeg(0x030, 0, 0x0093, 0xffff, 0)       // ds
	vmcbSeg(0x040, 0, 0x0093, 0xffff, 0)       // fs
	vmcbSeg(0x050, 0, 0x0093, 0xffff, 0)       // gs
	vmcbSeg(0x060, 0, 0x0000, 0xffff, 0)       // gdtr
	vmcbSeg(0x070, 0, 0x0082, 0xffff, 0)       // ldtr
	vmcbSeg(0x080, 0, 0x0000, 0xffff, 0)       // idtr
	vmcbSeg(0x090, 0, 0x008b, 0xffff, 0)       // tr

	binary.LittleEndian.PutUint64(page[0x0d0:], 0x1000)         // efer (SVME)
	binary.LittleEndian.PutUint64(page[0x148:], 0x40)           // cr4 (MCE)
	binary.LittleEndian.PutUint64(page[0x158:], 0x10)           // cr0 (PE)
	binary.LittleEndian.PutUint64(page[0x160:], 0x400)          // dr7
	binary.LittleEndian.PutUint64(page[0x168:], 0xffff0ff0)     // dr6
	binary.LittleEndian.PutUint64(page[0x170:], 0x2)            // rflags
	binary.LittleEndian.PutUint64(page[0x178:], rip)            // rip
	binary.LittleEndian.PutUint64(page[0x268:], 0x0007040600070406) // g_pat
	binary.LittleEndian.PutUint64(page[0x310:], uint64(vcpuSig)) // rdx (CPUID sig)
	binary.LittleEndian.PutUint64(page[0x3b0:], guestFeat)      // sev_features
	binary.LittleEndian.PutUint64(page[0x3e8:], 0x1)            // xcr0
	binary.LittleEndian.PutUint32(page[0x408:], 0x1f80)         // mxcsr
	binary.LittleEndian.PutUint16(page[0x410:], 0x037f)         // x87_fcw

	return page
}

// CalcSevMeasurement computes the expected SEV-SNP launch digest for a given
// registry entry, vcpu count, and kernel cmdline.
func CalcSevMeasurement(entry *SevArtifactEntry, vcpus int, cmdline string) string {
	ld, _ := hex.DecodeString(entry.OvmfHash)

	offsetInPage := int(entry.SevHashesTableGPA & 0xfff)
	hashesPage := buildHashesPage(entry.KernelHash, entry.InitrdHash, cmdline, offsetInPage)

	for _, sec := range entry.OvmfSections {
		gpa := sec.GPA
		switch sec.SectionType {
		case 1: // SNP_SEC_MEM
			ld = gctxUpdateZeroPages(ld, gpa, sec.Size)
		case 2: // SNP_SECRETS
			ld = gctxUpdateSecretsPage(ld, gpa)
		case 3: // CPUID
			ld = gctxUpdateCpuidPage(ld, gpa)
		case 4: // SVSM_CAA
			ld = gctxUpdateZeroPages(ld, gpa, sec.Size)
		case 0x10: // SNP_KERNEL_HASHES
			ld = gctxUpdateNormalPages(ld, gpa, hashesPage)
		}
	}

	apEIP := entry.SevEsResetEIP
	for i := 0; i < vcpus; i++ {
		eip := uint32(bspEIP)
		if i != 0 {
			eip = apEIP
		}
		vmsa := buildVmsaPage(eip, vcpuSigEPYC, guestFeatures)
		ld = gctxUpdateVmsaPage(ld, vmsa)
	}

	return hex.EncodeToString(ld)
}
