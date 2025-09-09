package genericchatstorage

import (
	"encoding/json"
	"reflect"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/sashabaranov/go-openai"
)

type AudioSpeechRequest struct {
	openai.CreateSpeechRequest
	Extra map[string]json.RawMessage `json:"-"`
}

func (c *AudioSpeechRequest) UnmarshalJSON(data []byte) error {
	type base openai.CreateSpeechRequest
	var known base
	if err := json.Unmarshal(data, &known); err != nil {
		return err
	}
	c.CreateSpeechRequest = openai.CreateSpeechRequest(known)

	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}
	lib.StripKnownKeys(all, reflect.TypeOf(known))
	c.Extra = all
	return nil
}

func (c AudioSpeechRequest) MarshalJSON() ([]byte, error) {
	type base openai.CreateSpeechRequest
	b, err := json.Marshal(base(c.CreateSpeechRequest))
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
