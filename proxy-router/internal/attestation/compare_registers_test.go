package attestation

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// perTemplateGolden mirrors what CI publishes for SEV since #722: a
// per_template map and no legacy single measurement.
var perTemplateGolden = &GoldenValues{SEVPerTemplate: map[string]string{
	"small":  "AAAA1111",
	"medium": "BBBB2222",
	"large":  "CCCC3333",
}}

func sevResult(measurement, template string) *AttestationResult {
	return &AttestationResult{Valid: true, Type: TEETypeSEV, Measurement: measurement, SEVTemplate: template}
}

// Regression for the bounty report: with only per_template published the SEV
// branch read the empty legacy field, logged "skipping" and passed any digest.
func TestCompareRegisters_SEVPerTemplateOnly_RejectsUnknownMeasurement(t *testing.T) {
	err := CompareRegisters(sevResult("deadbeef", ""), perTemplateGolden, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "matches none of the published golden values")
	assert.Contains(t, err.Error(), "large=CCCC3333 medium=BBBB2222 small=AAAA1111")
}

func TestCompareRegisters_SEVPerTemplateOnly_AcceptsAnyPublishedWhenTemplateUnknown(t *testing.T) {
	assert.NoError(t, CompareRegisters(sevResult("bbbb2222", ""), perTemplateGolden, nil), "case-insensitive match")
}

func TestCompareRegisters_SEVTemplateKnown_UsesThatEntry(t *testing.T) {
	assert.NoError(t, CompareRegisters(sevResult("aaaa1111", "small"), perTemplateGolden, nil))

	err := CompareRegisters(sevResult("bbbb2222", "small"), perTemplateGolden, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "(small=AAAA1111)")
}

func TestCompareRegisters_SEVTemplateNotPublished(t *testing.T) {
	err := CompareRegisters(sevResult("aaaa1111", "4xlarge"), perTemplateGolden, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `no golden value for VM template "4xlarge"`)
}

func TestCompareRegisters_SEVLegacyMeasurement(t *testing.T) {
	golden := &GoldenValues{Measurement: "aaaa1111"}
	assert.NoError(t, CompareRegisters(sevResult("AAAA1111", ""), golden, nil))
	assert.Error(t, CompareRegisters(sevResult("deadbeef", ""), golden, nil))
}

// A manifest carrying both shapes accepts either, so a non-default VM size is
// not rejected just because a legacy value is also present.
func TestCompareRegisters_SEVLegacyAndPerTemplateTogether(t *testing.T) {
	golden := &GoldenValues{Measurement: "LEGACY99", SEVPerTemplate: perTemplateGolden.SEVPerTemplate}
	assert.NoError(t, CompareRegisters(sevResult("bbbb2222", ""), golden, nil))
	assert.NoError(t, CompareRegisters(sevResult("legacy99", ""), golden, nil))
	assert.NoError(t, CompareRegisters(sevResult("legacy99", "small"), golden, nil))
	assert.NoError(t, CompareRegisters(sevResult("legacy99", "4xlarge"), golden, nil), "unpublished template falls back to legacy")
	assert.Error(t, CompareRegisters(sevResult("bbbb2222", "small"), golden, nil))
}

func TestCompareRegisters_FailsClosedWithoutGoldenForTEEType(t *testing.T) {
	err := CompareRegisters(sevResult("aaaa1111", ""), &GoldenValues{RTMR3: "tdx-only"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "publish no measurement for SEV")

	err = CompareRegisters(&AttestationResult{Type: TEETypeTDX, RTMR3: "rtmr3"}, &GoldenValues{Measurement: "sev-only"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "publish no RTMR3 for TDX")
}

func TestCompareRegisters_RegisterMissingFromQuote(t *testing.T) {
	err := CompareRegisters(sevResult("", ""), perTemplateGolden, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "measurement not present in quote")

	err = CompareRegisters(&AttestationResult{Type: TEETypeTDX}, &GoldenValues{RTMR3: "abc"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "RTMR3 not present in quote")
}

func TestCompareRegisters_TDX(t *testing.T) {
	golden := &GoldenValues{RTMR3: "abc"}
	assert.NoError(t, CompareRegisters(&AttestationResult{Type: TEETypeTDX, RTMR3: "ABC"}, golden, nil))
	assert.Error(t, CompareRegisters(&AttestationResult{Type: TEETypeTDX, RTMR3: "xyz"}, golden, nil))
}

func TestCompareRegisters_NilInputs(t *testing.T) {
	assert.Error(t, CompareRegisters(nil, perTemplateGolden, nil))
	assert.Error(t, CompareRegisters(sevResult("aaaa1111", ""), nil, nil))
	assert.Error(t, CompareRegisters(&AttestationResult{Type: "SGX", RTMR3: "x"}, perTemplateGolden, nil))
}

func TestSevTemplateFromQuote(t *testing.T) {
	raw := make([]byte, 0x100)
	copy(raw[0x010:0x020], "svm-small-sev") // family_id = "<vmType>-<template>-sev"
	assert.Equal(t, "small", SevTemplateFromQuote(base64.StdEncoding.EncodeToString(raw)))

	copy(raw[0x010:0x020], make([]byte, 16))
	copy(raw[0x010:0x020], "svm-huge-sev") // not a SecretVM template
	assert.Equal(t, "", SevTemplateFromQuote(base64.StdEncoding.EncodeToString(raw)))

	assert.Equal(t, "", SevTemplateFromQuote("not base64!!"))
	assert.Equal(t, "", SevTemplateFromQuote(base64.StdEncoding.EncodeToString([]byte("short"))))
}
