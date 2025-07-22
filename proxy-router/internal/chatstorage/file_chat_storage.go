package chatstorage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
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
func (cs *ChatStorage) StorePromptResponseToFile(identifier string, isLocal bool, modelId string, prompt interface{}, responses []gcs.Chunk, promptAt time.Time, responseAt time.Time) error {
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

	resps := make([]string, len(responses))
	for i, r := range responses {
		resps[i] = r.String()
	}

	isImageContent := false
	isVideoRawContent := false
	isAudioContent := false
	if len(responses) > 0 {
		isImageContent = responses[0].Type() == gcs.ChunkTypeImage
		isVideoRawContent = responses[0].Type() == gcs.ChunkTypeVideo
		isAudioContent = responses[0].Type() == gcs.ChunkTypeAudioTranscriptionText ||
			responses[0].Type() == gcs.ChunkTypeAudioTranscriptionJson ||
			responses[0].Type() == gcs.ChunkTypeAudioTranscriptionDelta
	}

	var newEntry gcs.ChatMessage
	var title string

	switch p := prompt.(type) {
	case *gcs.OpenAICompletionRequestExtra:
		newEntry = gcs.ChatMessage{
			Prompt:            prompt,
			Response:          strings.Join(resps, ""),
			PromptAt:          promptAt.Unix(),
			ResponseAt:        responseAt.Unix(),
			IsImageContent:    isImageContent,
			IsVideoRawContent: isVideoRawContent,
			IsAudioContent:    isAudioContent,
		}
		title = p.Messages[0].Content
	case *gcs.AudioTranscriptionRequest:
		// Store audio transcription request directly
		newEntry = gcs.ChatMessage{
			Prompt:            p,
			Response:          strings.Join(resps, ""),
			PromptAt:          promptAt.Unix(),
			ResponseAt:        responseAt.Unix(),
			IsImageContent:    isImageContent,
			IsVideoRawContent: isVideoRawContent,
			IsAudioContent:    isAudioContent,
		}
		// Use a default title for audio transcription or the prompt if available
		if p.Prompt != "" {
			title = "Audio Transcription: " + p.Prompt
		} else {
			title = "Audio Transcription"
		}
	case *gcs.AudioSpeechRequest:
		// Store audio speech request directly
		newEntry = gcs.ChatMessage{
			Prompt:            p,
			Response:          strings.Join(resps, ""),
			PromptAt:          promptAt.Unix(),
			ResponseAt:        responseAt.Unix(),
			IsImageContent:    isImageContent,
			IsVideoRawContent: isVideoRawContent,
			IsAudioContent:    isAudioContent,
		}
	default:
		return fmt.Errorf("unsupported prompt type: %T", prompt)
	}

	if chatHistory.Messages == nil && len(chatHistory.Messages) == 0 {
		chatHistory.ModelId = modelId
		chatHistory.Title = title
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
