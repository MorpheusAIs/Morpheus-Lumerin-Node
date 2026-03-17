package attestation

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

// Intel TDX Quote v4 binary layout offsets and sizes.
// Reference: Intel TDX DCAP Quoting Library API, Quote Format v4.
const (
	quoteHeaderSize   = 48
	tdReportBodySize  = 584
	minTDXQuoteSize   = quoteHeaderSize + tdReportBodySize
	tdxTEEType uint32 = 0x81
)

// Quote header field offsets (from start of quote).
const (
	offVersion    = 0  // uint16 LE
	offAttKeyType = 2  // uint16 LE
	offTEEType    = 4  // uint32 LE
	offReserved   = 8  // 4 bytes (QE SVN + PCE SVN in some specs)
	offVendorID   = 12 // 16 bytes
	offUserData   = 28 // 20 bytes
)

// TD Report Body field offsets (relative to start of body at quoteHeaderSize).
const (
	rbOffTCBSVN       = 0   // 16 bytes
	rbOffMRSeam       = 16  // 48 bytes
	rbOffMRSignerSeam = 64  // 48 bytes
	rbOffSeamAttrs    = 112 // 8 bytes
	rbOffTDAttrs      = 120 // 8 bytes
	rbOffXFAM         = 128 // 8 bytes
	rbOffMRTD         = 136 // 48 bytes
	rbOffMRConfigID   = 184 // 48 bytes
	rbOffMROwner      = 232 // 48 bytes
	rbOffMROwnerCfg   = 280 // 48 bytes
	rbOffRTMR0        = 328 // 48 bytes
	rbOffRTMR1        = 376 // 48 bytes
	rbOffRTMR2        = 424 // 48 bytes
	rbOffRTMR3        = 472 // 48 bytes
	rbOffReportData   = 520 // 64 bytes
)

// TDXQuote holds parsed fields from an Intel TDX Quote v4 binary.
type TDXQuote struct {
	Version      uint16
	AttKeyType   uint16
	TEEType      uint32
	TCBSVN       string
	MRSeam       string
	MRSignerSeam string
	TDAttributes string
	XFAM         string
	MRTD         string
	MRConfigID   string
	MROwner      string
	MROwnerCfg   string
	RTMR0        string
	RTMR1        string
	RTMR2        string
	RTMR3        string
	ReportData   string
}

// parseTDXQuoteHex decodes a hex-encoded TDX attestation quote and extracts
// the TD Report Body fields. This provides a local fallback for extracting
// register values without calling the SecretAI Portal API.
//
// Note: this only parses the binary structure — it does NOT perform
// cryptographic verification of the quote signature against Intel's PCK
// certificate chain. The SecretAI Portal API should be preferred when
// available for full verification.
func parseTDXQuoteHex(hexQuote string) (*TDXQuote, error) {
	hexQuote = strings.TrimSpace(hexQuote)
	raw, err := hex.DecodeString(hexQuote)
	if err != nil {
		return nil, fmt.Errorf("invalid hex encoding: %w", err)
	}

	if len(raw) < minTDXQuoteSize {
		return nil, fmt.Errorf("quote too short: %d bytes, need at least %d", len(raw), minTDXQuoteSize)
	}

	version := binary.LittleEndian.Uint16(raw[offVersion:])
	if version != 4 {
		return nil, fmt.Errorf("unsupported quote version %d (expected 4)", version)
	}

	teeType := binary.LittleEndian.Uint32(raw[offTEEType:])
	if teeType != tdxTEEType {
		return nil, fmt.Errorf("not a TDX quote: tee_type=0x%08x (expected 0x%08x)", teeType, tdxTEEType)
	}

	rb := raw[quoteHeaderSize:] // TD Report Body

	return &TDXQuote{
		Version:      version,
		AttKeyType:   binary.LittleEndian.Uint16(raw[offAttKeyType:]),
		TEEType:      teeType,
		TCBSVN:       hex.EncodeToString(rb[rbOffTCBSVN : rbOffTCBSVN+16]),
		MRSeam:       hex.EncodeToString(rb[rbOffMRSeam : rbOffMRSeam+48]),
		MRSignerSeam: hex.EncodeToString(rb[rbOffMRSignerSeam : rbOffMRSignerSeam+48]),
		TDAttributes: hex.EncodeToString(rb[rbOffTDAttrs : rbOffTDAttrs+8]),
		XFAM:         hex.EncodeToString(rb[rbOffXFAM : rbOffXFAM+8]),
		MRTD:         hex.EncodeToString(rb[rbOffMRTD : rbOffMRTD+48]),
		MRConfigID:   hex.EncodeToString(rb[rbOffMRConfigID : rbOffMRConfigID+48]),
		MROwner:      hex.EncodeToString(rb[rbOffMROwner : rbOffMROwner+48]),
		MROwnerCfg:   hex.EncodeToString(rb[rbOffMROwnerCfg : rbOffMROwnerCfg+48]),
		RTMR0:        hex.EncodeToString(rb[rbOffRTMR0 : rbOffRTMR0+48]),
		RTMR1:        hex.EncodeToString(rb[rbOffRTMR1 : rbOffRTMR1+48]),
		RTMR2:        hex.EncodeToString(rb[rbOffRTMR2 : rbOffRTMR2+48]),
		RTMR3:        hex.EncodeToString(rb[rbOffRTMR3 : rbOffRTMR3+48]),
		ReportData:   hex.EncodeToString(rb[rbOffReportData : rbOffReportData+64]),
	}, nil
}

// toAttestationResult converts a parsed TDX quote into the standard
// AttestationResult used by the verifier.
func (q *TDXQuote) toAttestationResult() *AttestationResult {
	return &AttestationResult{
		Valid: true,
		Type:  TEETypeTDX,
		MRTD:  q.MRTD,
		RTMR0: q.RTMR0,
		RTMR1: q.RTMR1,
		RTMR2: q.RTMR2,
		RTMR3: q.RTMR3,
	}
}
