package attestation

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"strings"
)

// CalculateRTMR3 computes the expected RTMR3 value from a docker-compose.yaml
// content and a rootfs_data hash (hex string from the artifact registry).
func CalculateRTMR3(dockerComposeBytes []byte, rootfsDataHex string) string {
	composeHash := sha256.Sum256(dockerComposeBytes)
	log := []string{
		hex.EncodeToString(composeHash[:]),
		strings.ToLower(strings.TrimPrefix(rootfsDataHex, "0x")),
	}
	return replayRtmr(log)
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
