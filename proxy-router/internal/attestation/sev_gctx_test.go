package attestation

import (
	"testing"
)

func TestParseSevFamilyID_Valid(t *testing.T) {
	// "prod-small-sev" padded to 16 bytes with nulls
	input := make([]byte, 16)
	copy(input, "prod-small-sev")

	result := ParseSevFamilyID(input)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.VMType != "prod" {
		t.Fatalf("expected VMType=prod, got %s", result.VMType)
	}
	if result.TemplateName != "small" {
		t.Fatalf("expected TemplateName=small, got %s", result.TemplateName)
	}
	if result.Vcpus != 1 {
		t.Fatalf("expected Vcpus=1, got %d", result.Vcpus)
	}
}

func TestParseSevFamilyID_Large(t *testing.T) {
	input := make([]byte, 16)
	copy(input, "dev-large-sev")

	result := ParseSevFamilyID(input)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.VMType != "dev" {
		t.Fatalf("expected VMType=dev, got %s", result.VMType)
	}
	if result.TemplateName != "large" {
		t.Fatalf("expected TemplateName=large, got %s", result.TemplateName)
	}
	if result.Vcpus != 4 {
		t.Fatalf("expected Vcpus=4, got %d", result.Vcpus)
	}
}

func TestParseSevFamilyID_NoSevSuffix(t *testing.T) {
	input := make([]byte, 16)
	copy(input, "prod-small")

	result := ParseSevFamilyID(input)
	if result != nil {
		t.Fatal("expected nil for missing -sev suffix")
	}
}

func TestParseSevFamilyID_NoDash(t *testing.T) {
	input := make([]byte, 16)
	copy(input, "prodsmall-sev")

	result := ParseSevFamilyID(input)
	if result != nil {
		t.Fatal("expected nil for missing dash in core")
	}
}

func TestParseSevFamilyID_UnknownTemplate(t *testing.T) {
	input := make([]byte, 16)
	copy(input, "prod-huge-sev")

	result := ParseSevFamilyID(input)
	if result != nil {
		t.Fatal("expected nil for unknown template name")
	}
}

func TestParseSevFamilyID_AllZeros(t *testing.T) {
	input := make([]byte, 16)
	result := ParseSevFamilyID(input)
	if result != nil {
		t.Fatal("expected nil for all-zero input")
	}
}

func TestCalcSevMeasurement_Deterministic(t *testing.T) {
	entry := &SevArtifactEntry{
		KernelHash: "98c41a86a1ba6a9a9d772ae0b028835091b4930f79ea509b595d2080d7df90c2",
		InitrdHash: "5f99893492640368a4324e51377296e3ebf4989598297fdea36943da3317aa7a",
		OvmfHash:   "c581d3eaebf2941beb1f757de97497279b953a6999921cab05f9ed5268f9c0505d741f4021b5a3995c9893851cde190e",
		SevHashesTableGPA: 8457216,
		SevEsResetEIP:     8433668,
		RootfsHash: "fc0a5cc3e9e7e1f72dde8d48a12ae592327ef0e8a9e78991af27b6aea52ac47e",
		OvmfSections: []SevOvmfSection{
			{GPA: 8388608, Size: 36864, SectionType: 1},
			{GPA: 8429568, Size: 12288, SectionType: 1},
			{GPA: 8441856, Size: 4096, SectionType: 2},
			{GPA: 8445952, Size: 4096, SectionType: 3},
			{GPA: 8450048, Size: 4096, SectionType: 4},
			{GPA: 8458240, Size: 61440, SectionType: 1},
			{GPA: 8454144, Size: 4096, SectionType: 16},
		},
	}

	cmdline := "console=ttyS0 loglevel=7 docker_compose_hash=abcd1234 rootfs_hash=" + entry.RootfsHash
	result1 := CalcSevMeasurement(entry, 1, cmdline)
	result2 := CalcSevMeasurement(entry, 1, cmdline)

	if result1 != result2 {
		t.Fatalf("expected deterministic results, got %s and %s", result1, result2)
	}
	if len(result1) != 96 { // SHA-384 = 48 bytes = 96 hex chars
		t.Fatalf("expected 96 hex chars, got %d", len(result1))
	}
}

func TestCalcSevMeasurement_DifferentVcpus(t *testing.T) {
	entry := &SevArtifactEntry{
		KernelHash: "98c41a86a1ba6a9a9d772ae0b028835091b4930f79ea509b595d2080d7df90c2",
		InitrdHash: "5f99893492640368a4324e51377296e3ebf4989598297fdea36943da3317aa7a",
		OvmfHash:   "c581d3eaebf2941beb1f757de97497279b953a6999921cab05f9ed5268f9c0505d741f4021b5a3995c9893851cde190e",
		SevHashesTableGPA: 8457216,
		SevEsResetEIP:     8433668,
		RootfsHash: "fc0a5cc3e9e7e1f72dde8d48a12ae592327ef0e8a9e78991af27b6aea52ac47e",
		OvmfSections: []SevOvmfSection{
			{GPA: 8388608, Size: 36864, SectionType: 1},
			{GPA: 8429568, Size: 12288, SectionType: 1},
			{GPA: 8441856, Size: 4096, SectionType: 2},
			{GPA: 8445952, Size: 4096, SectionType: 3},
			{GPA: 8450048, Size: 4096, SectionType: 4},
			{GPA: 8458240, Size: 61440, SectionType: 1},
			{GPA: 8454144, Size: 4096, SectionType: 16},
		},
	}

	cmdline := "console=ttyS0 loglevel=7 docker_compose_hash=test rootfs_hash=" + entry.RootfsHash
	result1 := CalcSevMeasurement(entry, 1, cmdline)
	result2 := CalcSevMeasurement(entry, 2, cmdline)

	if result1 == result2 {
		t.Fatal("expected different measurements for different vcpu counts")
	}
}

func TestCalcSevMeasurement_DifferentCmdline(t *testing.T) {
	entry := &SevArtifactEntry{
		KernelHash: "98c41a86a1ba6a9a9d772ae0b028835091b4930f79ea509b595d2080d7df90c2",
		InitrdHash: "5f99893492640368a4324e51377296e3ebf4989598297fdea36943da3317aa7a",
		OvmfHash:   "c581d3eaebf2941beb1f757de97497279b953a6999921cab05f9ed5268f9c0505d741f4021b5a3995c9893851cde190e",
		SevHashesTableGPA: 8457216,
		SevEsResetEIP:     8433668,
		RootfsHash: "fc0a5cc3e9e7e1f72dde8d48a12ae592327ef0e8a9e78991af27b6aea52ac47e",
		OvmfSections: []SevOvmfSection{
			{GPA: 8388608, Size: 36864, SectionType: 1},
			{GPA: 8429568, Size: 12288, SectionType: 1},
			{GPA: 8441856, Size: 4096, SectionType: 2},
			{GPA: 8445952, Size: 4096, SectionType: 3},
			{GPA: 8450048, Size: 4096, SectionType: 4},
			{GPA: 8458240, Size: 61440, SectionType: 1},
			{GPA: 8454144, Size: 4096, SectionType: 16},
		},
	}

	result1 := CalcSevMeasurement(entry, 1, "console=ttyS0 loglevel=7 docker_compose_hash=aaaa rootfs_hash="+entry.RootfsHash)
	result2 := CalcSevMeasurement(entry, 1, "console=ttyS0 loglevel=7 docker_compose_hash=bbbb rootfs_hash="+entry.RootfsHash)

	if result1 == result2 {
		t.Fatal("expected different measurements for different cmdlines")
	}
}
