// Package modelsfile loads a provider's models-config.json for diagnostics
// commands, accepting both the V2 ({"models": [...]}) and the legacy
// (modelId -> config map) layouts that config.ModelConfigLoader accepts.
package modelsfile

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
)

// Entry pairs a model ID with its backend config.
type Entry struct {
	ID  string
	Cfg config.ModelConfig
}

// Load reads and parses the file, returning entries sorted by model ID.
func Load(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var probe map[string]json.RawMessage
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, fmt.Errorf("invalid models config format: %w", err)
	}

	var entries []Entry
	if probe["models"] != nil {
		var v2 config.ModelConfigsV2
		if err := json.Unmarshal(data, &v2); err != nil {
			return nil, fmt.Errorf("invalid models config V2 format: %w", err)
		}
		for _, m := range v2.Models {
			entries = append(entries, Entry{ID: m.ID, Cfg: m.ModelConfig})
		}
	} else {
		var legacy config.ModelConfigs
		if err := json.Unmarshal(data, &legacy); err != nil {
			return nil, fmt.Errorf("invalid models config: %w", err)
		}
		for id, cfg := range legacy {
			entries = append(entries, Entry{ID: id, Cfg: cfg})
		}
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	return entries, nil
}

// Filter keeps entries whose model name or ID contains needle
// (case-insensitive); an empty needle keeps everything.
func Filter(entries []Entry, needle string) []Entry {
	if needle == "" {
		return entries
	}
	n := strings.ToLower(needle)
	var out []Entry
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e.Cfg.ModelName), n) || strings.Contains(strings.ToLower(e.ID), n) {
			out = append(out, e)
		}
	}
	return out
}
