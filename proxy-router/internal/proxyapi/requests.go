package proxyapi

import (
	"encoding/json"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

type PingReq struct {
	ProviderAddr common.Address `json:"providerAddr" validate:"required,eth_addr"`
	ProviderURL  string         `json:"providerUrl"  validate:"required,hostname_port"`
}

type PingRes struct {
	PingMs int64 `json:"ping,omitempty"`
}

type InitiateSessionReq struct {
	User        common.Address `json:"user"        validate:"required,eth_addr"`
	Provider    common.Address `json:"provider"    validate:"required,eth_addr"`
	Spend       lib.BigInt     `json:"spend"       validate:"required,number" swaggertype:"string"`
	ProviderUrl string         `json:"providerUrl" validate:"required,hostname_port"`
	BidID       common.Hash    `json:"bidId"       validate:"required,hex32"`
}

type PromptReq struct {
	Signature string          `json:"signature" validate:"required,hexadecimal"`
	Message   json.RawMessage `json:"message"   validate:"required"`
	Timestamp string          `json:"timestamp" validate:"required,timestamp"`
}

type PromptHead struct {
	SessionID lib.Hash `header:"session_id" validate:"hex32"`
	ModelID   lib.Hash `header:"model_id"   validate:"hex32"`
	ChatID    lib.Hash `header:"chat_id"    validate:"hex32"`
}

type AgentPromptHead struct {
	SessionID lib.Hash `header:"session_id" validate:"hex32"`
	AgentId   lib.Hash `header:"agent_id"   validate:"hex32"`
}

type InferenceRes struct {
	Signature lib.HexString   `json:"signature,omitempty" validate:"required,hexadecimal"`
	Message   json.RawMessage `json:"message" validate:"required"`
	Timestamp uint64          `json:"timestamp" validate:"required,timestamp"`
}

type UpdateChatTitleReq struct {
	Title string `json:"title" validate:"required"`
}

type ResultResponse struct {
	Result bool `json:"result"`
}

type ChatCompletionRequestSwaggerExample struct {
	Stream   bool `json:"stream"`
	Messages []struct {
		Role    string `json:"role" example:"user"`
		Content string `json:"content" example:"tell me a joke"`
	} `json:"messages"`
}

type AudioSpeechRequestExample struct {
	Input string `json:"input" example:"This is a text to speech generation prompt."`
	Voice string `json:"voice" example:"af_bella"`
	ResponseFormat string `json:"response_format" example:"mp3"`
	Speed float64 `json:"speed" example:"0.5"`
}	

type EmbeddingsRequestExample struct {
	Input string `json:"input" example:"This is a text to generate embeddings for."`
	Dimensions int `json:"dimensions" example:"1024"`
	EncodingFormat string `json:"encoding_format" example:"float"`
	User string `json:"user"`
}

type CIDReq struct {
	CID lib.Hash `json:"cidHash" validate:"required,hex32" swaggertype:"string"`
}

type AddFileReq struct {
	FilePath  string        `json:"filePath" binding:"required" validate:"required"`
	Tags      []string      `json:"tags"`
	ID        lib.HexString `json:"id" swaggertype:"string"`
	ModelName string        `json:"modelName"`
}

type AddIpfsFileRes struct {
	FileCID         string        `json:"fileCID" validate:"required"`
	MetadataCID     string        `json:"metadataCID" validate:"required"`
	FileCIDHash     lib.HexString `json:"fileCIDHash" validate:"required" swaggertype:"string"`
	MetadataCIDHash lib.HexString `json:"metadataCIDHash" validate:"required" swaggertype:"string"`
}

type IpfsVersionRes struct {
	Version string `json:"version" validate:"required"`
}

type PinnedFileRes struct {
	FileName        string        `json:"fileName"`
	FileSize        int64         `json:"fileSize"`
	FileCID         string        `json:"fileCID" validate:"required"`
	FileCIDHash     lib.HexString `json:"fileCIDHash" validate:"required" swaggertype:"string"`
	Tags            []string      `json:"tags"`
	ID              string        `json:"id"`
	ModelName       string        `json:"modelName"`
	MetadataCID     string        `json:"metadataCID" validate:"required"`
	MetadataCIDHash lib.HexString `json:"metadataCIDHash" validate:"required" swaggertype:"string"`
}

type DownloadFileReq struct {
	DestinationPath string `json:"destinationPath" validate:"required"`
}

type DownloadProgressEvent struct {
	Status      string  `json:"status"`          // "downloading", "completed", "error"
	Downloaded  int64   `json:"downloaded"`      // Bytes downloaded so far
	Total       int64   `json:"total"`           // Total bytes to download
	Percentage  float64 `json:"percentage"`      // Percentage complete (0-100)
	Error       string  `json:"error,omitempty"` // Error message, if status is "error"
	TimeUpdated int64   `json:"timeUpdated"`     // Timestamp of the update
}

// DockerBuildReq defines the request for building a Docker image
type DockerBuildReq struct {
	ContextPath string            `json:"contextPath" binding:"required" validate:"required"`
	Dockerfile  string            `json:"dockerfile" validate:"required"`
	ImageName   string            `json:"imageName" binding:"required" validate:"required"`
	ImageTag    string            `json:"imageTag"`
	BuildArgs   map[string]string `json:"buildArgs"`
}

// DockerBuildRes defines the response for building a Docker image
type DockerBuildRes struct {
	ImageTag string `json:"imageTag" validate:"required"`
}

// DockerStartContainerReq defines the request for starting a Docker container
type DockerStartContainerReq struct {
	ImageName     string            `json:"imageName" binding:"required" validate:"required"`
	ContainerName string            `json:"containerName"`
	Env           []string          `json:"env"`
	Ports         map[string]string `json:"ports"`
	Volumes       map[string]string `json:"volumes"`
	NetworkMode   string            `json:"networkMode"`
}

// DockerStartContainerRes defines the response for starting a Docker container
type DockerStartContainerRes struct {
	ContainerID string `json:"containerId" validate:"required"`
}

// DockerContainerActionReq defines the request for container actions (stop, remove)
type DockerContainerActionReq struct {
	ContainerID string `json:"containerId" binding:"required" validate:"required"`
	Timeout     int    `json:"timeout,omitempty"`
	Force       bool   `json:"force,omitempty"`
}

// DockerContainerInfoRes defines the response for container info
type DockerContainerInfoRes struct {
	ContainerInfo
}

// DockerListContainersReq defines the request for listing containers
type DockerListContainersReq struct {
	All          bool              `json:"all"`
	FilterLabels map[string]string `json:"filterLabels"`
}

// DockerListContainersRes defines the response for listing containers
type DockerListContainersRes struct {
	Containers []ContainerInfo `json:"containers"`
}

// DockerLogsReq defines the request for container logs
type DockerLogsReq struct {
	ContainerID string `json:"containerId" binding:"required" validate:"required"`
	Tail        int    `json:"tail,omitempty"`
	Follow      bool   `json:"follow,omitempty"`
}

// DockerStreamBuildEvent defines a stream event for Docker image building
type DockerStreamBuildEvent struct {
	Status       string  `json:"status"` // Status message
	Stream       string  `json:"stream,omitempty"`
	Progress     string  `json:"progress,omitempty"`
	ID           string  `json:"id,omitempty"`
	Current      int64   `json:"current,omitempty"`
	Total        int64   `json:"total,omitempty"`
	Percentage   float64 `json:"percentage,omitempty"`
	Error        string  `json:"error,omitempty"`
	TimeUpdated  int64   `json:"timeUpdated"`
	ErrorDetails string  `json:"errorDetails,omitempty"`
}

// DockerVersionRes defines the response for Docker version
type DockerVersionRes struct {
	Version string `json:"version" validate:"required"`
}

// DockerPruneRes defines the response for Docker pruning operations
type DockerPruneRes struct {
	SpaceReclaimed int64 `json:"spaceReclaimed"`
}

type CallAgentToolReq struct {
	ToolName string                 `json:"toolName" validate:"required"`
	Input    map[string]interface{} `json:"input" validate:"required"`
}
