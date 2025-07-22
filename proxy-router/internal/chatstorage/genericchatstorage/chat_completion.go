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

	// 3) Delete the keys we already mapped into the typed struct
	lib.StripKnownKeys(all, reflect.TypeOf(known))

	// Whatever is left is vendor-specific
	c.Extra = all
	return nil
}

func (c ChatCompletionResponseExtra) MarshalJSON() ([]byte, error) {
	// 1) Marshal the embedded OpenAI part on its own.
	type base openai.ChatCompletionResponse // avoid recursion
	b, err := json.Marshal(base(c.ChatCompletionResponse))
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

type ChatCompletionStreamResponseExtra struct {
	openai.ChatCompletionStreamResponse                            // typed, known part
	Extra                               map[string]json.RawMessage `json:"-"` // unknown bits
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

	// 3) Delete the keys we already mapped into the typed struct
	lib.StripKnownKeys(all, reflect.TypeOf(known))

	// Whatever is left is vendor-specific
	c.Extra = all
	return nil
}

func (c ChatCompletionStreamResponseExtra) MarshalJSON() ([]byte, error) {
	// 1) Marshal the embedded OpenAI part on its own.
	type base openai.ChatCompletionStreamResponse // avoid recursion
	b, err := json.Marshal(base(c.ChatCompletionStreamResponse))
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
