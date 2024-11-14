package chatstorage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/sashabaranov/go-openai"
)

// ChatStorage handles storing conversations to files.
type ChatStorage struct {
	dirPath            string                 // Directory path to store the files
	fileMutexes        map[string]*sync.Mutex // Map to store mutexes for each file
	forwardChatContext bool
}

// NewChatStorage creates a new instance of ChatStorage.
func NewChatStorage(dirPath string) *ChatStorage {
	return &ChatStorage{
		dirPath:     dirPath,
		fileMutexes: make(map[string]*sync.Mutex),
	}
}

// StorePromptResponseToFile stores the prompt and response to a file.
func (cs *ChatStorage) StorePromptResponseToFile(identifier string, isLocal bool, modelId string, prompt *openai.ChatCompletionRequest, responses []gcs.Chunk, promptAt time.Time, responseAt time.Time) error {
	if err := os.MkdirAll(cs.dirPath, os.ModePerm); err != nil {
		return err
	}

	filePath := filepath.Join(cs.dirPath, identifier+".json")
	cs.initFileMutex(filePath)

	// Lock the file mutex
	cs.fileMutexes[filePath].Lock()
	defer cs.fileMutexes[filePath].Unlock()

	var chatHistory gcs.ChatHistory
	if _, err := os.Stat(filePath); err == nil {
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(fileContent, &chatHistory); err != nil {
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

	p := gcs.OpenAiCompletionRequest{
		Messages:         messages,
		Model:            prompt.Model,
		MaxTokens:        prompt.MaxTokens,
		Temperature:      prompt.Temperature,
		TopP:             prompt.TopP,
		FrequencyPenalty: prompt.FrequencyPenalty,
		PresencePenalty:  prompt.PresencePenalty,
		Stop:             prompt.Stop,
	}

	resps := make([]string, len(responses))
	for i, r := range responses {
		resps[i] = r.String()
	}

	isImageContent := false
	if len(responses) > 0 {
		isImageContent = responses[0].Type() == gcs.ChunkTypeImage
	}

	newEntry := gcs.ChatMessage{
		Prompt:         p,
		Response:       strings.Join(resps, ""),
		PromptAt:       promptAt.Unix(),
		ResponseAt:     responseAt.Unix(),
		IsImageContent: isImageContent,
	}

	if chatHistory.Messages == nil && len(chatHistory.Messages) == 0 {
		chatHistory.ModelId = modelId
		chatHistory.Title = prompt.Messages[0].Content
		chatHistory.IsLocal = isLocal
	}

	newMessages := append(chatHistory.Messages, newEntry)
	chatHistory.Messages = newMessages

	updatedContent, err := json.MarshalIndent(chatHistory, "", "  ")
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
