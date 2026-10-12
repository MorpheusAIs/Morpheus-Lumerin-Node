package attestation

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"strings"
)

// SecretVM measures the workload into RTMR3 with one of two schemes. Which
// one a VM uses is decided by its init: images whose KMS is dstack/gramine
// (v0.0.34+ on the portal) expose a non-empty dstack app-id on GET /info and
// extend RTMR3 with dstack event digests; older images extend the raw hashes
// (attest-tool). Mirrors scrtlabs/secretvm-verify 0.13.0 node/src/rtmr.ts.
//
// attest-tool (no app-id):
//  1. sha256(docker-compose bytes)
//  2. rootfs_data
//
// dstack-util (app-id present):
//  1. digest("app-id",        app-id bytes)
//  2. digest("compose-hash",  sha256(docker-compose bytes))
//  3. digest("os-image-hash", rootfs_data bytes)
//
// where digest(name, payload) = sha384(LE32(0x08000001) || ":" || name || ":" || payload).
// Both schemes then fold each entry with mr = sha384(mr || pad48(entry)) from
// 48 zero bytes.

// dstackEventType is the fixed, non-TCG event type dstack-util uses for RTMR3
// app events (see dstack cc-eventlog). It is always serialized little-endian
// because the measuring host is x86_64 TDX.
const dstackEventType uint32 = 0x08000001

// RTMR3Scheme names the replay scheme that reproduced an RTMR3 value.
type RTMR3Scheme string

const (
	RTMR3SchemeAttestTool RTMR3Scheme = "attest-tool"
	RTMR3SchemeDstack     RTMR3Scheme = "dstack"
)

// NormalizeDstackAppID lowercases a dstack app-id and strips a 0x prefix and
// surrounding whitespace. Returns "" for an empty or non-hex value so callers
// can treat it as "no app-id" and fall back to the attest-tool scheme.
func NormalizeDstackAppID(appID string) string {
	s := strings.ToLower(strings.TrimSpace(appID))
	s = strings.TrimPrefix(s, "0x")
	if s == "" {
		return ""
	}
	if _, err := hex.DecodeString(s); err != nil {
		return ""
	}
	return s
}

// DstackEventDigest computes the dstack-util event digest that gets extended
// into RTMR3: sha384(LE32(event_type) || ":" || name || ":" || payload).
func DstackEventDigest(eventName string, payload []byte) string {
	var et [4]byte
	binary.LittleEndian.PutUint32(et[:], dstackEventType)

	h := sha512.New384()
	h.Write(et[:])
	h.Write([]byte(":"))
	h.Write([]byte(eventName))
	h.Write([]byte(":"))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// CalculateRTMR3 computes the expected RTMR3 value from docker-compose bytes
// and a rootfs_data hash using the attest-tool scheme (no dstack app-id).
// Kept for the pre-dstack SecretVM fleet.
func CalculateRTMR3(dockerComposeBytes []byte, rootfsDataHex string) string {
	return CalculateRTMR3WithAppID(dockerComposeBytes, rootfsDataHex, "")
}

// CalculateRTMR3WithAppID computes the expected RTMR3 for docker-compose bytes,
// a rootfs_data hash, and an optional dstack app-id. An empty app-id selects
// the attest-tool scheme; a non-empty one selects the dstack scheme.
func CalculateRTMR3WithAppID(dockerComposeBytes []byte, rootfsDataHex string, dstackAppID string) string {
	composeHash := sha256.Sum256(dockerComposeBytes)
	return CalculateRTMR3FromHashes(hex.EncodeToString(composeHash[:]), rootfsDataHex, dstackAppID)
}

// CalculateRTMR3FromHashes is the hash-level form of CalculateRTMR3WithAppID.
// It takes the compose SHA-256 (hex) instead of the compose bytes so a
// verifier that only holds the published compose digest (the signed golden
// manifest) can still reproduce the per-VM dstack RTMR3 once it learns the
// VM's app-id.
func CalculateRTMR3FromHashes(composeSHA256Hex string, rootfsDataHex string, dstackAppID string) string {
	composeHex := normalizeDigestHex(composeSHA256Hex)
	rootfsHex := normalizeDigestHex(rootfsDataHex)
	appID := NormalizeDstackAppID(dstackAppID)

	if appID == "" {
		return replayRtmr([]string{composeHex, rootfsHex})
	}

	composeBytes, _ := hex.DecodeString(composeHex)
	rootfsBytes, _ := hex.DecodeString(rootfsHex)
	appIDBytes, _ := hex.DecodeString(appID)
	return replayRtmr([]string{
		DstackEventDigest("app-id", appIDBytes),
		DstackEventDigest("compose-hash", composeBytes),
		DstackEventDigest("os-image-hash", rootfsBytes),
	})
}

// RTMR3SchemeFor reports which scheme an app-id selects.
func RTMR3SchemeFor(dstackAppID string) RTMR3Scheme {
	if NormalizeDstackAppID(dstackAppID) == "" {
		return RTMR3SchemeAttestTool
	}
	return RTMR3SchemeDstack
}

// normalizeDigestHex is normalizeHex plus stripping a "sha256:" prefix, the
// form the signed manifest uses for its top-level compose digest.
func normalizeDigestHex(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "sha256:")
	return normalizeHex(s)
}

// replayRtmr replays the RTMR extend operation over a sequence of hex-encoded entries.
// Starting from 48 zero bytes, for each entry:
//   - Decode from hex, right-pad to 48 bytes with zeros
//   - Compute SHA-384(current_mr || padded_entry)
//   - Take first 48 bytes as new mr
//
// Returns the final mr as a lowercase hex string.
func replayRtmr(log []string) string {
	var mr [48]byte

	for _, entry := range log {
		entryBytes, _ := hex.DecodeString(entry)

		var padded [48]byte
		copy(padded[:], entryBytes)

		var buf [96]byte
		copy(buf[:48], mr[:])
		copy(buf[48:], padded[:])

		mr = sha512.Sum384(buf[:])
	}

	return hex.EncodeToString(mr[:])
}
