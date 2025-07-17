package genericchatstorage

import (
	"encoding/json"
	"reflect"
	"strings"

	openai "github.com/sashabaranov/go-openai" // or whatever path you use
)

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
	stripKnownKeys(all, reflect.TypeOf(known))

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
	stripKnownKeys(all, reflect.TypeOf(known))

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
	stripKnownKeys(all, reflect.TypeOf(known))

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

// stripKnownKeys removes JSON field names found in the struct type from m.
func stripKnownKeys(m map[string]json.RawMessage, t reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		if comma := strings.Index(tag, ","); comma != -1 {
			tag = tag[:comma]
		}
		delete(m, tag)
	}
}

// unmarshallWithExtras unmarshals data into knownStructPtr and
// returns the vendor-specific keys that were left over.
func unmarshalWithExtras[T any](data []byte, knownStructPtr *T) (map[string]json.RawMessage, error) {
	// 1) Into the typed struct
	if err := json.Unmarshal(data, knownStructPtr); err != nil {
		return nil, err
	}

	// 2) Everything into a map
	all := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &all); err != nil {
		return nil, err
	}

	// 3) Strip the fields we already know
	stripKnownKeys(all, reflect.TypeOf(*knownStructPtr))

	return all, nil
}
