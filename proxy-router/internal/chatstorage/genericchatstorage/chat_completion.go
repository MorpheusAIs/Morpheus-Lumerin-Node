package genericchatstorage

import (
	"encoding/json"
	"reflect"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/sashabaranov/go-openai"
)

type OpenAICompletionRequestExtra struct {
	openai.ChatCompletionRequest
	Extra map[string]json.RawMessage `json:"-"`
}

func (c *OpenAICompletionRequestExtra) UnmarshalJSON(data []byte) error {
	type base openai.ChatCompletionRequest
	var known base
	if err := json.Unmarshal(data, &known); err != nil {
		return err
	}
	c.ChatCompletionRequest = openai.ChatCompletionRequest(known)

	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}
	lib.StripKnownKeys(all, reflect.TypeOf(known))
	c.Extra = all
	return nil
}

func (c OpenAICompletionRequestExtra) MarshalJSON() ([]byte, error) {
	type base openai.ChatCompletionRequest
	b, err := json.Marshal(base(c.ChatCompletionRequest))
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

// ChatCompletionResponse preserves the standard OpenAI response *and* any extra keys.
type ChatCompletionResponseExtra struct {
	openai.ChatCompletionResponse                            // typed, known part
	Extra                         map[string]json.RawMessage `json:"-"` // unknown bits
	originalJSON                  map[string]json.RawMessage // preserve original structure
}

func (c *ChatCompletionResponseExtra) UnmarshalJSON(data []byte) error {
	// 1) Unmarshal into the embedded OpenAI struct (using an alias to avoid recursion)
	type base openai.ChatCompletionResponse
	var known base
	if err := json.Unmarshal(data, &known); err != nil {
		return err
	}
	c.ChatCompletionResponse = openai.ChatCompletionResponse(known)

	// 2) Unmarshal into a generic map so we can see every key
	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}

	// Store the original JSON structure to preserve exact formatting
	c.originalJSON = make(map[string]json.RawMessage)
	for k, v := range all {
		c.originalJSON[k] = v
	}

	// 3) Delete the keys we already mapped into the typed struct
	lib.StripKnownKeys(all, reflect.TypeOf(known))

	// Whatever is left is vendor-specific
	c.Extra = all
	return nil
}

func (c ChatCompletionResponseExtra) MarshalJSON() ([]byte, error) {
	// Use the original JSON structure if available (preserves original fields and omits defaults)
	if c.originalJSON != nil {
		m := make(map[string]json.RawMessage)
		for k, v := range c.originalJSON {
			m[k] = v
		}
		
		// Merge vendor-specific keys from Extra if they were modified
		for k, v := range c.Extra {
			m[k] = v
		}
		
		return json.Marshal(m)
	}

	// Fallback to the old method if originalJSON is not available
	// (e.g., if the struct was created programmatically)
	type base openai.ChatCompletionResponse
	b, err := json.Marshal(base(c.ChatCompletionResponse))
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

type ChatCompletionStreamResponseExtra struct {
	openai.ChatCompletionStreamResponse                            // typed, known part
	Extra                               map[string]json.RawMessage `json:"-"` // unknown bits
	originalJSON                        map[string]json.RawMessage // preserve original structure
}

func (c *ChatCompletionStreamResponseExtra) UnmarshalJSON(data []byte) error {
	// 1) Unmarshal into the embedded OpenAI struct (using an alias to avoid recursion)
	type base openai.ChatCompletionStreamResponse
	var known base
	if err := json.Unmarshal(data, &known); err != nil {
		return err
	}
	c.ChatCompletionStreamResponse = openai.ChatCompletionStreamResponse(known)

	// 2) Unmarshal into a generic map so we can see every key
	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}

	// Store the original JSON structure to preserve exact formatting
	c.originalJSON = make(map[string]json.RawMessage)
	for k, v := range all {
		c.originalJSON[k] = v
	}

	// 3) Delete the keys we already mapped into the typed struct
	lib.StripKnownKeys(all, reflect.TypeOf(known))

	// Whatever is left is vendor-specific
	c.Extra = all
	return nil
}

func (c ChatCompletionStreamResponseExtra) MarshalJSON() ([]byte, error) {
	// Use the original JSON structure if available (preserves original fields and omits defaults)
	if c.originalJSON != nil {
		m := make(map[string]json.RawMessage)
		for k, v := range c.originalJSON {
			m[k] = v
		}
		
		// Merge vendor-specific keys from Extra if they were modified
		for k, v := range c.Extra {
			m[k] = v
		}
		
		return json.Marshal(m)
	}

	// Fallback to the old method if originalJSON is not available
	// (e.g., if the struct was created programmatically)
	type base openai.ChatCompletionStreamResponse
	b, err := json.Marshal(base(c.ChatCompletionStreamResponse))
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
