package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/google/shlex"
)

// Config struct for loading values from mor-launch.json
type Config struct {
	LlamaURL      string   `json:"llama_url"`
	LlamaRelease  string   `json:"llama_release"`
	LlamaFileBase string   `json:"llama_filebase"`
	ModelURL      string   `json:"model_url"`
	ModelOwner    string   `json:"model_owner"`
	ModelRepo     string   `json:"model_repo"`
	ModelName     string   `json:"model_name"`
	Run           []string `json:"run"`
}

// Detects OS and architecture to determine the correct binary name
func getBinName() string {
	osMap := map[string]string{
		"darwin":  "macos",
		"linux":   "ubuntu",
		"windows": "win-avx2",
	}
	archMap := map[string]string{
		"arm64": "arm64",
		"amd64": "x64",
	}

	osName, osExists := osMap[runtime.GOOS]
	archName, archExists := archMap[runtime.GOARCH]

	if !osExists || !archExists {
		log.Fatalf("Unsupported OS/Architecture: %s-%s", runtime.GOOS, runtime.GOARCH)
	}

	return fmt.Sprintf("%s-%s", osName, archName)
}

// Prompts the user for confirmation
func askForConfirmation(prompt string) bool {
	fmt.Print(prompt + " [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Failed to read input: %v", err)
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// Checks if the required files already exist
func filesExist(files ...string) bool {
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// Downloads a file from a given URL
func downloadFile(filepath string, url string) error {
	if _, err := os.Stat(filepath); err == nil {
		log.Printf("File already exists: %s", filepath)
		return nil
	}

	log.Printf("Downloading: %s", url)
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// Extracts a specific file from a zip archive
func extractFileFromZip(zipPath, fileToExtract, destPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == fileToExtract {
			log.Printf("Extracting %s from %s", fileToExtract, zipPath)
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			outFile, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, rc)
			return err
		}
	}
	return fmt.Errorf("file %s not found in zip %s", fileToExtract, zipPath)
}

// Reads configuration from mor-launch.json
func readConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &Config{}
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}

func main() {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error finding executable path: %v", err)
	}
	base := filepath.Dir(exePath)

	// Load configuration from mor-launch.json
	configPath := filepath.Join(base, "mor-launch.json")
	config, err := readConfig(configPath)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Determine correct binary name
	binName := getBinName()
	llamaZip := filepath.Join(base, fmt.Sprintf("%s-%s.zip", config.LlamaFileBase, binName))
	llamaBinary := filepath.Join(base, "llama-server")
	modelFile := filepath.Join(base, config.ModelName)

	// Check if files exist to bypass the prompt
	autoRunLlamaServer := filesExist(llamaBinary, modelFile)

	// Ask for confirmation only if files are not already present
	runLlamaServer := autoRunLlamaServer || askForConfirmation("Do you want to download and run the local model?")

	if runLlamaServer && !autoRunLlamaServer {
		// Construct URLs
		llamaDownloadURL := fmt.Sprintf("%s/%s/%s-%s.zip", config.LlamaURL, config.LlamaRelease, config.LlamaFileBase, binName)
		modelDownloadURL := fmt.Sprintf("%s/%s/%s/resolve/main/%s", config.ModelURL, config.ModelOwner, config.ModelRepo, config.ModelName)

		// Download necessary files
		if err := downloadFile(llamaZip, llamaDownloadURL); err != nil {
			log.Fatalf("Failed to download Llama binary: %v", err)
		}
		if err := downloadFile(modelFile, modelDownloadURL); err != nil {
			log.Fatalf("Failed to download model file: %v", err)
		}

		// Extract llama-server binary
		if err := extractFileFromZip(llamaZip, "build/bin/llama-server", llamaBinary); err != nil {
			log.Fatalf("Failed to extract llama-server: %v", err)
		}

		// Set execute permission on llama-server
		if err := os.Chmod(llamaBinary, 0755); err != nil {
			log.Fatalf("Failed to set execute permission on llama-server: %v", err)
		}
	}

	// Execute commands from config
	var wg sync.WaitGroup
	for _, cmdStr := range config.Run {
		if !runLlamaServer && strings.Contains(cmdStr, "llama-server") {
			log.Println("Skipping llama-server command based on user input.")
			continue
		}

		wg.Add(1)
		go func(cmdStr string) {
			defer wg.Done()
			args, err := shlex.Split(cmdStr)
			if err != nil {
				log.Printf("Error parsing command: %s", cmdStr)
				return
			}
			if len(args) == 0 {
				log.Printf("Empty command, skipping")
				return
			}

			cmd := exec.Command(args[0], args[1:]...)
			cmd.Dir = base
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				log.Printf("Error creating stdout pipe: %v", err)
				return
			}

			if err := cmd.Start(); err != nil {
				log.Printf("Error starting command: %s, %v", cmdStr, err)
				return
			}

			if _, err := io.Copy(os.Stdout, stdout); err != nil {
				log.Printf("Error reading stdout: %v", err)
				return
			}

			if err := cmd.Wait(); err != nil {
				log.Printf("Error waiting for command: %s, %v", cmdStr, err)
			}
		}(cmdStr)
	}
	wg.Wait()
}
