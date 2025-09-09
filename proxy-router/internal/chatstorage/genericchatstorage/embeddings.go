package genericchatstorage

import (
	"encoding/json"
	"reflect"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/sashabaranov/go-openai"
)

// EmbeddingsRequest is a wrapper around the standard OpenAI Embedding request which keeps unknown fields in `Extra`.
// This mirrors the implementation style of AudioSpeechRequest and OpenAICompletionRequestExtra.

type EmbeddingsRequest struct {
	openai.EmbeddingRequest
	Extra map[string]json.RawMessage `json:"-"`
}

func (c *EmbeddingsRequest) UnmarshalJSON(data []byte) error {
	type base openai.EmbeddingRequest
	var known base
	if err := json.Unmarshal(data, &known); err != nil {
		return err
	}
	c.EmbeddingRequest = openai.EmbeddingRequest(known)

	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}
	lib.StripKnownKeys(all, reflect.TypeOf(known))
	c.Extra = all
	return nil
}

func (c EmbeddingsRequest) MarshalJSON() ([]byte, error) {
	type base openai.EmbeddingRequest
	b, err := json.Marshal(base(c.EmbeddingRequest))
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

type EmbeddingsResponse struct {
	openai.EmbeddingResponse
	Extra map[string]json.RawMessage `json:"-"`
}

func (c *EmbeddingsResponse) UnmarshalJSON(data []byte) error {
	type base openai.EmbeddingResponse
	var known base
	if err := json.Unmarshal(data, &known); err != nil {
		return err
	}
	c.EmbeddingResponse = openai.EmbeddingResponse(known)

	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}
	lib.StripKnownKeys(all, reflect.TypeOf(known))
	c.Extra = all
	return nil
}

func (c EmbeddingsResponse) MarshalJSON() ([]byte, error) {
	type base openai.EmbeddingResponse
	b, err := json.Marshal(base(c.EmbeddingResponse))
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
