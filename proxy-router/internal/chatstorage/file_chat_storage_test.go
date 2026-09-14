package chatstorage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	openai "github.com/sashabaranov/go-openai"
)

func TestLoadChatRepairsExistingPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose Unix permission bits")
	}

	dirPath := filepath.Join(t.TempDir(), "chats")
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dirPath, 0o755); err != nil {
		t.Fatal(err)
	}

	history := gcs.ChatHistory{
		Title:   "existing chat",
		ModelId: "model-1",
		Messages: []gcs.ChatMessage{{
			Prompt:   map[string]interface{}{"messages": []interface{}{}},
			PromptAt: 1,
		}},
	}
	content, err := json.Marshal(history)
	if err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(dirPath, "existing.json")
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filePath, 0o644); err != nil {
		t.Fatal(err)
	}

	storage := NewChatStorage(dirPath)
	loaded, err := storage.LoadChatFromFile("existing")
	if err != nil {
		t.Fatalf("LoadChatFromFile() error = %v", err)
	}
	if loaded.Title != history.Title {
		t.Fatalf("title = %q, want %q", loaded.Title, history.Title)
	}
	assertPermission(t, dirPath, chatDirectoryMode)
	assertPermission(t, filePath, chatFileMode)
}

func TestChatWritesAreAtomicPrivateAndPreserveHistory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose Unix permission bits or stable file identity")
	}

	dirPath := filepath.Join(t.TempDir(), "chats")
	storage := NewChatStorage(dirPath)
	prompt := func(content string) *gcs.OpenAICompletionRequestExtra {
		return &gcs.OpenAICompletionRequestExtra{
			ChatCompletionRequest: openai.ChatCompletionRequest{
				Messages: []openai.ChatCompletionMessage{{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				}},
			},
		}
	}

	now := time.Unix(1, 0)
	if err := storage.StorePromptResponseToFile("chat-1", false, "model-1", prompt("first"), nil, now, now); err != nil {
		t.Fatalf("first StorePromptResponseToFile() error = %v", err)
	}
	filePath := filepath.Join(dirPath, "chat-1.json")
	beforeAppend, err := os.Stat(filePath)
	if err != nil {
		t.Fatal(err)
	}

	// Simulate a chat created by an older release, then verify the next write
	// repairs both existing paths before replacing the JSON atomically.
	if err := os.Chmod(dirPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filePath, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := storage.StorePromptResponseToFile("chat-1", false, "model-1", prompt("second"), nil, now, now); err != nil {
		t.Fatalf("second StorePromptResponseToFile() error = %v", err)
	}
	afterAppend, err := os.Stat(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(beforeAppend, afterAppend) {
		t.Error("append rewrote the existing inode; expected atomic replacement")
	}
	assertPermission(t, dirPath, chatDirectoryMode)
	assertPermission(t, filePath, chatFileMode)

	beforeTitleUpdate := afterAppend
	if err := os.Chmod(filePath, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := storage.UpdateChatTitle("chat-1", "renamed"); err != nil {
		t.Fatalf("UpdateChatTitle() error = %v", err)
	}
	afterTitleUpdate, err := os.Stat(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(beforeTitleUpdate, afterTitleUpdate) {
		t.Error("title update rewrote the existing inode; expected atomic replacement")
	}
	assertPermission(t, filePath, chatFileMode)

	loaded, err := storage.LoadChatFromFile("chat-1")
	if err != nil {
		t.Fatalf("LoadChatFromFile() error = %v", err)
	}
	if loaded.Title != "renamed" {
		t.Errorf("title = %q, want renamed", loaded.Title)
	}
	if len(loaded.Messages) != 2 {
		t.Errorf("message count = %d, want 2", len(loaded.Messages))
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "chat-1.json" {
		t.Fatalf("chat directory entries = %v, expected only chat-1.json", entryNames(entries))
	}
}

func assertPermission(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Errorf("%s permissions = %04o, want %04o", path, got, want)
	}
}

func entryNames(entries []os.DirEntry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}
