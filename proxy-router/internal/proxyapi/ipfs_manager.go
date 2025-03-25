package proxyapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ipfs/boxo/files"
	ipfspath "github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
)

type IpfsManager struct {
	node *rpc.HttpApi
	log  lib.ILogger
}

// ProgressReader is an io.Reader wrapper that reports progress
type ProgressReader struct {
	Reader     io.Reader
	Total      int64
	Downloaded int64
	OnProgress ProgressCallback
	Ctx        context.Context // Add context for cancellation checks
}

// ProgressCallback is a function that reports download progress
type ProgressCallback func(downloaded, total int64) error

// NewIpfsManager connects to the local Kubo (IPFS) node using the RPC client.
func NewIpfsManager(log lib.ILogger) *IpfsManager {
	node, err := rpc.NewLocalApi()
	if err != nil {
		log.Error("Error creating IPFS client:", err)
		return &IpfsManager{node: nil, log: log}
	}
	return &IpfsManager{node: node, log: log}
}

func (i *IpfsManager) IsNodeReady() error {
	if i.node == nil {
		return fmt.Errorf("IPFS node is not ready")
	}
	return nil
}

// PinnedFileMetadata represents metadata about a file stored in IPFS
type PinnedFileMetadata struct {
	FileName    string   `json:"fileName"`
	FileSize    int64    `json:"fileSize"`
	FileCID     string   `json:"fileCID"`
	Tags        []string `json:"tags"`
	ID          string   `json:"id"`
	ModelName   string   `json:"modelName"`
	MetadataCID string   `json:"metadataCID,omitempty"`
}

// AddFileResult contains CIDs for both the file and its metadata
type AddFileResult struct {
	FileCID     string `json:"fileCID"`
	MetadataCID string `json:"metadataCID"`
}

// AddFile adds a file and its metadata to IPFS.
func (i *IpfsManager) AddFile(ctx context.Context, filePath string, tags []string, id string, modelName string) (*AddFileResult, error) {
	if err := i.IsNodeReady(); err != nil {
		return nil, err
	}

	// Open and get file info
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Add the main file to IPFS
	fileNode := files.NewReaderFile(f)
	addedFile, err := i.node.Unixfs().Add(ctx, fileNode)
	if err != nil {
		return nil, fmt.Errorf("failed to add file: %w", err)
	}
	fileCID := addedFile.RootCid().String()

	// Create metadata
	metadata := PinnedFileMetadata{
		FileName:  fileInfo.Name(),
		FileSize:  fileInfo.Size(),
		FileCID:   fileCID,
		Tags:      tags,
		ID:        id,
		ModelName: modelName,
	}

	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Add metadata to IPFS
	metadataNode := files.NewBytesFile(metadataJSON)
	addedMetadata, err := i.node.Unixfs().Add(ctx, metadataNode)
	if err != nil {
		return nil, fmt.Errorf("failed to add metadata: %w", err)
	}

	return &AddFileResult{
		FileCID:     fileCID,
		MetadataCID: addedMetadata.RootCid().String(),
	}, nil
}

// Pin pins a CID on the local IPFS node.
func (i *IpfsManager) Pin(ctx context.Context, cidStr string) error {
	if err := i.IsNodeReady(); err != nil {
		return err
	}

	c, err := cid.Decode(cidStr)
	if err != nil {
		return fmt.Errorf("invalid CID: %w", err)
	}

	p := ipfspath.FromCid(c)
	if err := i.node.Pin().Add(ctx, p); err != nil {
		return fmt.Errorf("failed to pin CID: %w", err)
	}
	return nil
}

// Unpin removes a pin for a given CID.
func (i *IpfsManager) Unpin(ctx context.Context, cidStr string) error {
	if err := i.IsNodeReady(); err != nil {
		return err
	}

	c, err := cid.Decode(cidStr)
	if err != nil {
		return fmt.Errorf("invalid CID: %w", err)
	}

	p := ipfspath.FromCid(c)
	if err := i.node.Pin().Rm(ctx, p); err != nil {
		return fmt.Errorf("failed to unpin CID: %w", err)
	}
	return nil
}

// GetFileWithProgress downloads a file using its metadata CID with progress reporting
func (i *IpfsManager) GetFileWithProgress(ctx context.Context, metadataCIDStr string, destinationPath string, progressCallback ProgressCallback) error {
	if err := i.IsNodeReady(); err != nil {
		return err
	}

	c, err := cid.Decode(metadataCIDStr)
	if err != nil {
		return fmt.Errorf("invalid metadata CID: %w", err)
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 1) "Probe" phase for metadata: short context to see if metadata is available
	probeCtx, probeCancel := context.WithTimeout(ctx, 10*time.Second)
	defer probeCancel()

	probeMetadataNode, err := i.node.Unixfs().Get(probeCtx, ipfspath.FromCid(c))
	if err != nil {
		return fmt.Errorf("failed to find metadata within 10s: %w", err)
	}

	probeMetadataFile, ok := probeMetadataNode.(files.File)
	if !ok {
		probeMetadataNode.Close()
		return fmt.Errorf("metadata object at CID %s is not a regular file", metadataCIDStr)
	}

	// Read a small chunk to ensure metadata is available
	buf := make([]byte, 1)
	_, err = probeMetadataFile.Read(buf)
	if err != nil && err != io.EOF {
		probeMetadataFile.Close()
		return fmt.Errorf("failed to read metadata in probe phase: %w", err)
	}

	probeMetadataFile.Close()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 2) Get the full metadata with a reasonable timeout
	metadataCtx, metadataCancel := context.WithTimeout(ctx, 2*time.Minute)
	defer metadataCancel()

	// Get metadata file
	metadataNode, err := i.node.Unixfs().Get(metadataCtx, ipfspath.FromCid(c))
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}
	defer metadataNode.Close()

	metadataFile, ok := metadataNode.(files.File)
	if !ok {
		return fmt.Errorf("metadata object at CID %s is not a regular file", metadataCIDStr)
	}

	// Read and parse metadata
	metadataBytes, err := io.ReadAll(metadataFile)
	if err != nil {
		return fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata PinnedFileMetadata
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Get the file size for progress reporting
	fileSize := metadata.FileSize

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Now get the actual file using FileCID from metadata
	// 3) "Probe" phase for actual file: short context to see if we can get ANY data
	fileProbeCtx, fileProbeCancel := context.WithTimeout(ctx, 10*time.Second)
	defer fileProbeCancel()

	fileC, err := cid.Decode(metadata.FileCID)
	if err != nil {
		return fmt.Errorf("invalid file CID in metadata: %w", err)
	}

	probeNode, err := i.node.Unixfs().Get(fileProbeCtx, ipfspath.FromCid(fileC))
	if err != nil {
		return fmt.Errorf("failed to find file within 10s: %w", err)
	}

	probeFile, ok := probeNode.(files.File)
	if !ok {
		probeNode.Close()
		return fmt.Errorf("object at CID %s is not a regular file", metadata.FileCID)
	}

	// Read a small chunk to ensure data is available
	fileBuf := make([]byte, 1)
	_, err = probeFile.Read(fileBuf)
	if err != nil && err != io.EOF {
		probeFile.Close()
		return fmt.Errorf("failed to read data in probe phase: %w", err)
	}

	probeFile.Close()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// 4) Download phase with longer timeout but still respect parent context
	downloadCtx, downloadCancel := context.WithTimeout(ctx, 12*time.Hour)
	defer downloadCancel()

	node, err := i.node.Unixfs().Get(downloadCtx, ipfspath.FromCid(fileC))
	if err != nil {
		return fmt.Errorf("download phase failed: %w", err)
	}
	defer node.Close()

	fileNode, ok := node.(files.File)
	if !ok {
		return fmt.Errorf("object at CID %s is not a regular file", metadata.FileCID)
	}

	destFile, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if progressCallback != nil {
		// Create a proxy reader that tracks progress
		progressReader := &ProgressReader{
			Reader:     fileNode,
			Total:      fileSize,
			Downloaded: 0,
			OnProgress: progressCallback,
			Ctx:        ctx, // Pass context to reader for cancellation checks
		}

		_, err := io.CopyBuffer(destFile, progressReader, make([]byte, 32*1024))
		if err != nil {
			// If the error is due to context cancellation, return that as the main error
			if ctx.Err() != nil {
				// Clean up the partial file
				destFile.Close()
				os.Remove(destinationPath)
				return ctx.Err()
			}
			return fmt.Errorf("failed to copy content: %w", err)
		}
	} else {
		// Even without progress callback, we should periodically check context
		buf := make([]byte, 32*1024) // 32KB buffer
		for {
			// Check context before each read
			if ctx.Err() != nil {
				destFile.Close()
				os.Remove(destinationPath)
				return ctx.Err()
			}
			
			n, err := fileNode.Read(buf)
			if n > 0 {
				_, writeErr := destFile.Write(buf[:n])
				if writeErr != nil {
					return fmt.Errorf("failed to write to destination: %w", writeErr)
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read content: %w", err)
			}
		}
	}

	return nil
}

// Read reads data from the underlying reader and reports progress
func (r *ProgressReader) Read(p []byte) (int, error) {
	if r.Ctx != nil && r.Ctx.Err() != nil {
		return 0, r.Ctx.Err()
	}
	
	n, err := r.Reader.Read(p)
	if n > 0 {
		r.Downloaded += int64(n)
		if r.OnProgress != nil {
			if progressErr := r.OnProgress(r.Downloaded, r.Total); progressErr != nil {
				return n, progressErr
			}
		}
	}
	
	if r.Ctx != nil && r.Ctx.Err() != nil {
		return n, r.Ctx.Err()
	}
	
	return n, err
}

// GetFile downloads a file using GetFileWithProgress with a nil progress callback
func (i *IpfsManager) GetFile(ctx context.Context, metadataCIDStr string, destinationPath string) error {
	return i.GetFileWithProgress(ctx, metadataCIDStr, destinationPath, nil)
}

func (i *IpfsManager) GetVersion(ctx context.Context) (string, error) {
	if err := i.IsNodeReady(); err != nil {
		return "", err
	}

	var resp struct {
		Version string `json:"Version"`
		Commit  string `json:"Commit,omitempty"`
		Repo    string `json:"Repo,omitempty"`
		System  string `json:"System,omitempty"`
		Golang  string `json:"Golang,omitempty"`
	}

	err := i.node.Request("version").Exec(ctx, &resp)
	if err != nil {
		return "", fmt.Errorf("request for version failed: %w", err)
	}

	return resp.Version, nil
}

// GetPinnedFiles returns metadata for all pinned files.
// It only returns metadata files, ignoring the actual content files.
func (i *IpfsManager) GetPinnedFiles(ctx context.Context) ([]PinnedFileMetadata, error) {
	if err := i.IsNodeReady(); err != nil {
		return nil, err
	}

	pinChan := make(chan iface.Pin)
	errChan := make(chan error, 1)

	go func() {
		err := i.node.Pin().Ls(ctx, pinChan)
		errChan <- err
		close(errChan)
	}()

	var metadataList []PinnedFileMetadata
	for pin := range pinChan {
		if pin.Type() == "indirect" {
			continue
		}

		fileCtx, _ := context.WithTimeout(ctx, 2*time.Second)

		// Try to get and parse the pinned file as metadata
		node, err := i.node.Unixfs().Get(fileCtx, pin.Path())

		if err != nil {
			i.log.Debug("Failed to get pinned file:", err)
			continue
		}

		file, ok := node.(files.File)
		if !ok {
			node.Close()
			continue
		}

		// Read only the first few KB to check if it's a metadata file
		// Metadata files should be very small
		limitReader := io.LimitReader(file, 8192) // 8KB limit
		metadataBytes, err := io.ReadAll(limitReader)
		file.Close()

		if err != nil {
			i.log.Debug("Failed to read pinned file:", err)
			continue
		}

		// If we hit the limit, it's probably not a metadata file
		if len(metadataBytes) >= 8192 {
			i.log.Debug("Skipping large file, likely not metadata:", pin.Path().String())
			continue
		}

		var metadata PinnedFileMetadata
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			// Not a metadata file, skip it
			continue
		}

		// Validate that this is actually a metadata file by checking required fields
		if metadata.FileCID == "" || metadata.FileName == "" {
			continue
		}

		// Add the metadata CID itself
		metadata.MetadataCID = pin.Path().RootCid().String()
		metadata.FileCID = metadata.FileCID

		metadataList = append(metadataList, metadata)
	}

	if err := <-errChan; err != nil {
		return nil, fmt.Errorf("failed to list pinned files: %w", err)
	}

	return metadataList, nil
}
