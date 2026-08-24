package genericchatstorage

import (
	"encoding/json"
	"testing"
)

// "content" arrives as a string for plain prompts and as an array of typed
// parts for multi-content (vision) prompts. The custom UnmarshalJSON must
// accept both — and must not lose the sibling fields while doing so.
func TestChatCompletionMessage_UnmarshalContentShapes(t *testing.T) {
	cases := []struct {
		name    string
		json    string
		want    ChatCompletionMessage
		wantErr bool
	}{
		{
			name: "string content",
			json: `{"role":"user","content":"hello","name":"n1","tool_call_id":"t1"}`,
			want: ChatCompletionMessage{Role: "user", Content: "hello", Name: "n1", ToolCallID: "t1"},
		},
		{
			name: "multi-content array keeps text parts and sibling fields",
			json: `{"role":"tool","name":"getweather","tool_call_id":"call_123","content":[{"type":"text","text":"hello"},{"type":"image_url","image_url":{"url":"data:image/png;base64,iVBORw0KGgo="}},{"type":"text","text":"world"}]}`,
			want: ChatCompletionMessage{Role: "tool", Content: "hello\nworld", Name: "getweather", ToolCallID: "call_123"},
		},
		{
			name: "image-only array flattens to empty content",
			json: `{"role":"user","content":[{"type":"image_url","image_url":{"url":"data:image/png;base64,iVBORw0KGgo="}}]}`,
			want: ChatCompletionMessage{Role: "user", Content: ""},
		},
		{
			name: "null content",
			json: `{"role":"assistant","content":null}`,
			want: ChatCompletionMessage{Role: "assistant", Content: ""},
		},
		{
			name: "absent content",
			json: `{"role":"assistant"}`,
			want: ChatCompletionMessage{Role: "assistant", Content: ""},
		},
		{
			name:    "content of an impossible type still errors",
			json:    `{"role":"user","content":123}`,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got ChatCompletionMessage
			err := json.Unmarshal([]byte(tc.json), &got)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
