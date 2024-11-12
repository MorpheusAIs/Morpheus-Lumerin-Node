package chatstorage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/completion"
	"github.com/sashabaranov/go-openai"
)

// ChatStorage handles storing conversations to files.
type ChatStorage struct {
	dirPath     string                 // Directory path to store the files
	fileMutexes map[string]*sync.Mutex // Map to store mutexes for each file
}

type Response interface {
}

// NewChatStorage creates a new instance of ChatStorage.
func NewChatStorage(dirPath string) *ChatStorage {
	return &ChatStorage{
		dirPath:     dirPath,
		fileMutexes: make(map[string]*sync.Mutex),
	}
}

// StorePromptResponseToFile stores the prompt and response to a file.
func (cs *ChatStorage) StorePromptResponseToFile(identifier string, isLocal bool, modelId string, prompt *openai.ChatCompletionRequest, responses []*completion.ChunkImpl, promptAt time.Time, responseAt time.Time) error {
	if err := os.MkdirAll(cs.dirPath, os.ModePerm); err != nil {
		return err
	}

	filePath := filepath.Join(cs.dirPath, identifier+".json")
	cs.initFileMutex(filePath)

	// Lock the file mutex
	cs.fileMutexes[filePath].Lock()
	defer cs.fileMutexes[filePath].Unlock()

	var data gcs.ChatHistory
	if _, err := os.Stat(filePath); err == nil {
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(fileContent, &data); err != nil {
			return err
		}
	}

	messages := make([]gcs.ChatCompletionMessage, 0)
	for _, r := range prompt.Messages {
		messages = append(messages, gcs.ChatCompletionMessage{
			Content: r.Content,
			Role:    r.Role,
		})
	}

	p := gcs.OpenAiCompletitionRequest{
		Messages:         messages,
		Model:            prompt.Model,
		MaxTokens:        prompt.MaxTokens,
		Temperature:      prompt.Temperature,
		TopP:             prompt.TopP,
		FrequencyPenalty: prompt.FrequencyPenalty,
		PresencePenalty:  prompt.PresencePenalty,
		Stop:             prompt.Stop,
	}

	newEntry := gcs.ChatMessage{
		Prompt:         p,
		Response:       responses,
		PromptAt:       promptAt.Unix(),
		ResponseAt:     responseAt.Unix(),
		IsImageContent: responses[0].Type == completion.ChunkTypeImage,
	}

	if data.Messages == nil && len(data.Messages) == 0 {
		data.ModelId = modelId
		data.Title = prompt.Messages[0].Content
		data.IsLocal = isLocal
	}

	newMessages := append(data.Messages, newEntry)
	data.Messages = newMessages

	updatedContent, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, updatedContent, 0644); err != nil {
		return err
	}

	return nil
}

func (cs *ChatStorage) GetChats() []gcs.Chat {
	var chats []gcs.Chat
	files, err := os.ReadDir(cs.dirPath)
	if err != nil {
		return chats
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		chatID := file.Name()
		chatID = chatID[:len(chatID)-5]

		fileContent, err := cs.LoadChatFromFile(chatID)
		if err != nil {
			continue
		}
		chats = append(chats, gcs.Chat{
			ChatID:    chatID,
			Title:     fileContent.Title,
			CreatedAt: fileContent.Messages[0].PromptAt,
			ModelID:   fileContent.ModelId,
			IsLocal:   fileContent.IsLocal,
		})
	}

	return chats
}

func (cs *ChatStorage) DeleteChat(identifier string) error {
	filePath := filepath.Join(cs.dirPath, identifier+".json")
	cs.initFileMutex(filePath)

	cs.fileMutexes[filePath].Lock()
	defer cs.fileMutexes[filePath].Unlock()

	if err := os.Remove(filePath); err != nil {
		return err
	}
	return nil
}

func (cs *ChatStorage) UpdateChatTitle(identifier string, title string) error {
	chat, err := cs.LoadChatFromFile(identifier)
	if err != nil {
		return err
	}
	chat.Title = title

	filePath := filepath.Join(cs.dirPath, identifier+".json")
	cs.initFileMutex(filePath)

	cs.fileMutexes[filePath].Lock()
	defer cs.fileMutexes[filePath].Unlock()

	updatedContent, err := json.MarshalIndent(chat, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, updatedContent, 0644); err != nil {
		return err
	}

	return nil
}

func (cs *ChatStorage) LoadChatFromFile(identifier string) (*gcs.ChatHistory, error) {
	filePath := filepath.Join(cs.dirPath, identifier+".json")
	cs.initFileMutex(filePath)

	cs.fileMutexes[filePath].Lock()
	defer cs.fileMutexes[filePath].Unlock()

	var data gcs.ChatHistory
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return &data, err
	}

	if err := json.Unmarshal(fileContent, &data); err != nil {
		fmt.Println("Error unmarshalling file content:", err)
		return nil, err
	}

	return &data, nil
}

// initFileMutex initializes a mutex for the file if not already present.
func (cs *ChatStorage) initFileMutex(filePath string) {
	if _, exists := cs.fileMutexes[filePath]; !exists {
		cs.fileMutexes[filePath] = &sync.Mutex{}
	}
}

type NoOpChatStorage struct{}

func NewNoOpChatStorage() *NoOpChatStorage {
	return &NoOpChatStorage{}
}

func (cs *NoOpChatStorage) LoadChatFromFile(chatID string) (*gcs.ChatHistory, error) {
	return nil, nil
}

func (cs *NoOpChatStorage) StorePromptResponseToFile(chatID string, isLocal bool, modelID string, prompt openai.ChatCompletionRequest, responses []interface{}, promptAt time.Time, responseAt time.Time) error {
	return nil
}

func (cs *NoOpChatStorage) GetChats() []gcs.Chat {
	return []gcs.Chat{}
}

func (cs *NoOpChatStorage) DeleteChat(chatID string) error {
	return nil
}

func (cs *NoOpChatStorage) UpdateChatTitle(chatID string, title string) error {
	return nil
}
