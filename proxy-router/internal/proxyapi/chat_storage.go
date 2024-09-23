package proxyapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/sashabaranov/go-openai"
)

// ChatStorage handles storing conversations to files.
type ChatStorage struct {
	dirPath     string                 // Directory path to store the files
	fileMutexes map[string]*sync.Mutex // Map to store mutexes for each file
}

// NewChatStorage creates a new instance of ChatStorage.
func NewChatStorage(dirPath string) *ChatStorage {
	return &ChatStorage{
		dirPath:     dirPath,
		fileMutexes: make(map[string]*sync.Mutex),
	}
}

// StorePromptResponseToFile stores the prompt and response to a file.
func (cs *ChatStorage) StorePromptResponseToFile(identifier string, isSession bool, prompt interface{}, responses []interface{}, promptAt, responseAt time.Time) error {
	var dir string
	if isSession {
		dir = "sessions"
	} else {
		dir = "models"
	}

	path := filepath.Join(cs.dirPath, dir)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	filePath := filepath.Join(path, identifier+".json")
	cs.initFileMutex(filePath)

	// Lock the file mutex
	cs.fileMutexes[filePath].Lock()
	defer cs.fileMutexes[filePath].Unlock()

	var data []map[string]interface{}
	if _, err := os.Stat(filePath); err == nil {
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(fileContent, &data); err != nil {
			return err
		}
	}

	response := ""
	for _, r := range responses {
		switch v := r.(type) {
		case ChatCompletionResponse:
			response += fmt.Sprintf("%v", v.Choices[0].Delta.Content)
		case *openai.ChatCompletionStreamResponse:
			response += fmt.Sprintf("%v", v.Choices[0].Delta.Content)
		case aiengine.ProdiaGenerationResult:
			response += fmt.Sprintf("%v", v.ImageUrl)
		case *aiengine.ProdiaGenerationResult:
			response += fmt.Sprintf("%v", v.ImageUrl)
		default:
			return fmt.Errorf("unknown response type")
		}
	}

	newEntry := map[string]interface{}{
		"prompt":     prompt,
		"response":   response,
		"promptAt":   promptAt.UnixMilli(),
		"responseAt": responseAt.UnixMilli(),
	}

	data = append(data, newEntry)
	updatedContent, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, updatedContent, 0644); err != nil {
		return err
	}

	return nil
}

// initFileMutex initializes a mutex for the file if not already present.
func (cs *ChatStorage) initFileMutex(filePath string) {
	if _, exists := cs.fileMutexes[filePath]; !exists {
		cs.fileMutexes[filePath] = &sync.Mutex{}
	}
}
