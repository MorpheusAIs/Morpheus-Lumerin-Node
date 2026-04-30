package mobile

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

func toJSON(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func bigStr(b *big.Int) string {
	if b == nil {
		return "0"
	}
	return b.String()
}

func bigIntStr(b *lib.BigInt) string {
	if b == nil {
		return "0"
	}
	return b.String()
}

// parseEthNodeURLs splits a config string into deduplicated RPC URLs (order preserved).
func parseEthNodeURLs(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ';' || r == '|' || r == '\n' || r == '\t'
	})
	seen := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		u := strings.TrimSpace(p)
		if u == "" {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}
