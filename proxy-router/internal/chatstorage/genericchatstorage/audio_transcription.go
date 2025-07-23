package genericchatstorage

import (
	"encoding/json"
	"reflect"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/sashabaranov/go-openai"
)

type AudioTranscriptionRequest struct {
	Model string `form:"model"`

	FilePath string `form:"file_path,omitempty"` // not usually sent; omit if empty

	Prompt      string                     `form:"prompt,omitempty"`
	Temperature float32                    `form:"temperature,omitempty"`
	Language    string                     `form:"language,omitempty"`
	Format      openai.AudioResponseFormat `form:"format,omitempty"`

	TimestampGranularities []openai.TranscriptionTimestampGranularity `form:"timestamp_granularities,omitempty"`

	Stream bool `form:"stream,omitempty"`

	Extra map[string]json.RawMessage `form:"-"`
}

func (c *AudioTranscriptionRequest) UnmarshalJSON(data []byte) error {
	// 1) Unmarshal into the embedded OpenAI struct (using an alias to avoid recursion)
	type base AudioTranscriptionRequest
	var known base
	if err := json.Unmarshal(data, &known); err != nil {
		return err
	}
	c.Model = known.Model
	c.FilePath = known.FilePath
	c.Prompt = known.Prompt
	c.Temperature = known.Temperature
	c.Language = known.Language
	c.Format = known.Format
	c.TimestampGranularities = known.TimestampGranularities
	c.Stream = known.Stream

	// 2) Unmarshal into a generic map so we can see every key
	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}

	// 3) Delete the keys we already mapped into the typed struct
	lib.StripKnownKeys(all, reflect.TypeOf(known))

	// Whatever is left is vendor-specific
	c.Extra = all
	return nil
}

func (c AudioTranscriptionRequest) MarshalJSON() ([]byte, error) {
	// 1) Marshal the embedded OpenAI part on its own.
	type base AudioTranscriptionRequest // avoid recursion
	b, err := json.Marshal(base(c))
	if err != nil {
		return nil, err
	}

	// 2) Turn that JSON into a map we can extend.
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	// 3) Merge vendor-specific keys.  Extra wins if names collide;
	// comment out the assignment if you want the opposite.
	for k, v := range c.Extra {
		m[k] = v
	}

	// 4) Re-encode the combined map.
	return json.Marshal(m)
}

type AudioResponseExtra struct {
	openai.AudioResponse
	Extra map[string]json.RawMessage `json:"-"`
}

func (c *AudioResponseExtra) UnmarshalJSON(data []byte) error {
	// 1) Unmarshal into the embedded OpenAI struct (using an alias to avoid recursion)
	type base openai.AudioResponse
	var known base
	if err := json.Unmarshal(data, &known); err != nil {
		return err
	}
	c.AudioResponse = openai.AudioResponse(known)

	// 2) Unmarshal into a generic map so we can see every key
	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}

	// 3) Delete the keys we already mapped into the typed struct
	lib.StripKnownKeys(all, reflect.TypeOf(known))

	// Whatever is left is vendor-specific
	c.Extra = all
	return nil
}

func (c AudioResponseExtra) MarshalJSON() ([]byte, error) {
	// 1) Marshal the embedded OpenAI part on its own.
	type base openai.AudioResponse // avoid recursion
	b, err := json.Marshal(base(c.AudioResponse))
	if err != nil {
		return nil, err
	}

	// 2) Turn that JSON into a map we can extend.
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	// 3) Merge vendor-specific keys.  Extra wins if names collide;
	// comment out the assignment if you want the opposite.
	for k, v := range c.Extra {
		m[k] = v
	}

	// 4) Re-encode the combined map.
	return json.Marshal(m)
}
