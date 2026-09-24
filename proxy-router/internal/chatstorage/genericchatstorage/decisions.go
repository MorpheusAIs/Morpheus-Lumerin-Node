package genericchatstorage

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

// DecisionsRequest is the TypeSafe System One / OpenRouter Decisions core body.
// Unknown request fields are preserved in Extra for upstream compatibility.
// Normative required fields: state + questions. model is overwritten by the adapter.
//
// State is json.RawMessage because TypeSafe accepts string|object|array. A Go
// string would 400 on object/array Unmarshal and silently coerce missing/null
// to "" (billed 200 with empty content). See OpenRouter Decisions state docs.
type DecisionsRequest struct {
	Model     string                     `json:"model,omitempty"`
	State     json.RawMessage            `json:"state"`
	Questions map[string]json.RawMessage `json:"questions"`
	Extra     map[string]json.RawMessage `json:"-"`
}

func (c *DecisionsRequest) UnmarshalJSON(data []byte) error {
	type known struct {
		Model     string                     `json:"model,omitempty"`
		State     json.RawMessage            `json:"state"`
		Questions map[string]json.RawMessage `json:"questions"`
	}
	var k known
	if err := json.Unmarshal(data, &k); err != nil {
		return err
	}
	c.Model = k.Model
	c.State = k.State
	c.Questions = k.Questions

	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}
	lib.StripKnownKeys(all, reflect.TypeOf(k))
	c.Extra = all
	return nil
}

func (c DecisionsRequest) MarshalJSON() ([]byte, error) {
	type known struct {
		Model     string                     `json:"model,omitempty"`
		State     json.RawMessage            `json:"state"`
		Questions map[string]json.RawMessage `json:"questions"`
	}
	b, err := json.Marshal(known{
		Model:     c.Model,
		State:     c.State,
		Questions: c.Questions,
	})
	if err != nil {
		return nil, err
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	for k, v := range c.Extra {
		m[k] = v
	}
	return json.Marshal(m)
}

// Validate ensures state and questions are present and non-empty so we never
// forward state:"" / missing state upstream (which TypeSafe accepts and bills).
func (c *DecisionsRequest) Validate() error {
	if err := validateDecisionsState(c.State); err != nil {
		return err
	}
	if len(c.Questions) == 0 {
		return errors.New("questions is required and must be non-empty")
	}
	return nil
}

func validateDecisionsState(state json.RawMessage) error {
	if len(state) == 0 {
		return errors.New("state is required")
	}
	trimmed := strings.TrimSpace(string(state))
	if trimmed == "" || trimmed == "null" {
		return errors.New("state is required")
	}
	// JSON string: reject "" and whitespace-only.
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(state, &s); err != nil {
			return fmt.Errorf("state: invalid JSON string: %w", err)
		}
		if strings.TrimSpace(s) == "" {
			return errors.New("state must be non-empty")
		}
		return nil
	}
	// object / array / number / bool: present and non-null is enough.
	// Empty {} / [] are unusual but still "present" content for TypeSafe.
	return nil
}

// DecisionsUsage mirrors TypeSafe / OpenRouter Decisions usage fields.
type DecisionsUsage struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	Cost         float64 `json:"cost,omitempty"`
}

// DecisionsResponse is the TypeSafe / OpenRouter Decisions response core.
type DecisionsResponse struct {
	Model   string                     `json:"model"`
	Answers map[string]json.RawMessage `json:"answers"`
	Usage   DecisionsUsage             `json:"usage"`
	Extra   map[string]json.RawMessage `json:"-"`
}

func (c *DecisionsResponse) UnmarshalJSON(data []byte) error {
	type known struct {
		Model   string                     `json:"model"`
		Answers map[string]json.RawMessage `json:"answers"`
		Usage   DecisionsUsage             `json:"usage"`
	}
	var k known
	if err := json.Unmarshal(data, &k); err != nil {
		return err
	}
	c.Model = k.Model
	c.Answers = k.Answers
	c.Usage = k.Usage

	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}
	lib.StripKnownKeys(all, reflect.TypeOf(k))
	c.Extra = all
	return nil
}

func (c DecisionsResponse) MarshalJSON() ([]byte, error) {
	type known struct {
		Model   string                     `json:"model"`
		Answers map[string]json.RawMessage `json:"answers"`
		Usage   DecisionsUsage             `json:"usage"`
	}
	b, err := json.Marshal(known{
		Model:   c.Model,
		Answers: c.Answers,
		Usage:   c.Usage,
	})
	if err != nil {
		return nil, err
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	for k, v := range c.Extra {
		m[k] = v
	}
	return json.Marshal(m)
}
