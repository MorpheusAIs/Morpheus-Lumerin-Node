package attestation

import (
	"encoding/base64"
	"sort"
	"strings"
)

// SevTemplateFromQuote returns the SecretVM template name ("small", "medium",
// ...) encoded in a raw SEV-SNP report's family_id, or "" when the quote does
// not decode or carries no recognised family_id. Phase 1 uses it to pick the
// per-template golden launch digest. The field sits in the hardware-signed
// report, so a provider cannot pick a template it does not run on; and every
// published template holds a digest of the same released image, so choosing
// among them never weakens the check.
func SevTemplateFromQuote(cpuQuoteBase64 string) string {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(cpuQuoteBase64))
	if err != nil || len(raw) < 0x020 {
		return ""
	}
	family := ParseSevFamilyID(raw[0x010:0x020])
	if family == nil {
		return ""
	}
	return family.TemplateName
}

// sevTemplateNames lists the template keys of a per-template golden map in a
// stable order, for error messages.
func sevTemplateNames(perTemplate map[string]string) []string {
	names := make([]string, 0, len(perTemplate))
	for name := range perTemplate {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
