package genericchatstorage

import (
	"encoding/json"
	"reflect"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

// DecisionsRequest is the TypeSafe System One / OpenRouter Decisions core body.
// Unknown request fields are preserved in Extra for upstream compatibility.
// Normative required fields: state + questions. model is overwritten by the adapter.
type DecisionsRequest struct {
	Model     string                     `json:"model,omitempty"`
	State     string                     `json:"state"`
	Questions map[string]json.RawMessage `json:"questions"`
	Extra     map[string]json.RawMessage `json:"-"`
}

func (c *DecisionsRequest) UnmarshalJSON(data []byte) error {
	type known struct {
		Model     string                     `json:"model,omitempty"`
		State     string                     `json:"state"`
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
		State     string                     `json:"state"`
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
