package attestation

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	tdxQuoteHeaderSize    = 48
	tdxTdReportBodySize   = 584
	tdxMinQuoteSize       = tdxQuoteHeaderSize + tdxTdReportBodySize
	tdxExpectedVersion    = 4
	tdxExpectedTeeType    = 0x81
)

type TdxQuoteFields struct {
	MRTD       string
	RTMR0      string
	RTMR1      string
	RTMR2      string
	RTMR3      string
	ReportData string
}

func ParseTdxQuoteFields(quoteHex string) (*TdxQuoteFields, error) {
	raw, err := hex.DecodeString(strings.TrimSpace(quoteHex))
	if err != nil {
		return nil, fmt.Errorf("invalid hex encoding: %w", err)
	}

	if len(raw) < tdxMinQuoteSize {
		return nil, fmt.Errorf("quote too short: got %d bytes, need at least %d", len(raw), tdxMinQuoteSize)
	}

	version := binary.LittleEndian.Uint16(raw[0:2])
	if version != tdxExpectedVersion {
		return nil, fmt.Errorf("unexpected quote version: got %d, expected %d", version, tdxExpectedVersion)
	}

	teeType := binary.LittleEndian.Uint32(raw[4:8])
	if teeType != tdxExpectedTeeType {
		return nil, fmt.Errorf("unexpected TEE type: got 0x%x, expected 0x%x", teeType, tdxExpectedTeeType)
	}

	return &TdxQuoteFields{
		MRTD:       hex.EncodeToString(raw[184:232]),
		RTMR0:      hex.EncodeToString(raw[376:424]),
		RTMR1:      hex.EncodeToString(raw[424:472]),
		RTMR2:      hex.EncodeToString(raw[472:520]),
		RTMR3:      hex.EncodeToString(raw[520:568]),
		ReportData: hex.EncodeToString(raw[568:632]),
	}, nil
}

func IsTdxQuote(quoteHex string) bool {
	raw, err := hex.DecodeString(strings.TrimSpace(quoteHex))
	if err != nil || len(raw) < 8 {
		return false
	}
	version := binary.LittleEndian.Uint16(raw[0:2])
	teeType := binary.LittleEndian.Uint32(raw[4:8])
	return version == tdxExpectedVersion && teeType == tdxExpectedTeeType
}
