package genericchatstorage

import (
	"encoding/json"
	"testing"

	"github.com/sashabaranov/go-openai"
)

// A stored chat is JSON on disk. When it is read back, ChatMessage.Prompt (an
// `interface{}`) unmarshals into map[string]interface{} — NOT the concrete
// OpenAiCompletionRequest. AppendChatHistory must still recover the turns.
//
// Before the fix this test fails: the type assertion
//     chat.Prompt.(OpenAiCompletionRequest)
// can never succeed after a round-trip, so every stored turn is silently dropped
// and the model is handed a conversation with no history — i.e. no memory.
func TestAppendChatHistory_SurvivesJSONRoundTrip(t *testing.T) {
	stored := ChatHistory{
		Messages: []ChatMessage{
			{
				Prompt: OpenAiCompletionRequest{
					Messages: []ChatCompletionMessage{
						{Role: "user", Content: "My favourite colour is teal."},
					},
				},
				Response: "Noted — teal.",
			},
		},
	}

	// Exactly what LoadChatFromFile does: write JSON, read it back into the struct.
	raw, err := json.Marshal(stored)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var loaded ChatHistory
	if err := json.Unmarshal(raw, &loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, ok := loaded.Messages[0].Prompt.(OpenAiCompletionRequest); ok {
		t.Fatal("precondition failed: Prompt should be a generic map after a JSON round-trip")
	}

	req := &OpenAICompletionRequestExtra{
		ChatCompletionRequest: openai.ChatCompletionRequest{
			Messages: []openai.ChatCompletionMessage{
				{Role: "user", Content: "What is my favourite colour?"},
			},
		},
	}

	got := loaded.AppendChatHistory(req)

	// prior user turn + prior assistant reply + the new prompt
	if len(got.Messages) != 3 {
		t.Fatalf("history dropped: got %d message(s), want 3 — the model would have no memory", len(got.Messages))
	}
	if got.Messages[0].Content != "My favourite colour is teal." {
		t.Errorf("first message = %q, want the stored user turn", got.Messages[0].Content)
	}
	if got.Messages[1].Role != "assistant" || got.Messages[1].Content != "Noted — teal." {
		t.Errorf("second message = %+v, want the stored assistant reply", got.Messages[1])
	}
	if got.Messages[2].Content != "What is my favourite colour?" {
		t.Errorf("last message = %q, want the new prompt", got.Messages[2].Content)
	}
}
