package main

import (
	"archive/zip"
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

// Gets the list of files to extract based on the OS
func getFilesToExtract() []string {
	if runtime.GOOS == "windows" {
		return []string{"llama-server.exe", "llama.dll", "ggml.dll", "ggml-base.dll", "ggml-cpu.dll", "ggml-rpc.dll"}
	}
	return []string{"build/bin/llama-server"}
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

// Extracts specific files from a zip archive
func extractFilesFromZip(zipPath string, filesToExtract []string, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	extractedFiles := make(map[string]bool)
	for _, f := range filesToExtract {
		extractedFiles[f] = false
	}

	for _, f := range r.File {
		if _, shouldExtract := extractedFiles[f.Name]; shouldExtract {
			destPath := filepath.Join(destDir, filepath.Base(f.Name))
			log.Printf("Extracting %s to %s", f.Name, destPath)

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

			if _, err = io.Copy(outFile, rc); err != nil {
				return err
			}

			if err := os.Chmod(destPath, 0755); err != nil {
				return fmt.Errorf("failed to set execute permission for %s: %v", destPath, err)
			}

			extractedFiles[f.Name] = true
		}
	}

	for file, extracted := range extractedFiles {
		if !extracted {
			return fmt.Errorf("file %s not found in zip %s", file, zipPath)
		}
	}

	return nil
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

// Replaces placeholder in run commands with the actual model name
func replaceModelPlaceholder(runCommands []string, modelName string) []string {
	updatedCommands := make([]string, len(runCommands))
	for i, cmd := range runCommands {
		updatedCommands[i] = strings.ReplaceAll(cmd, "{model_name}", modelName)
	}
	return updatedCommands
}

func main() {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error finding executable path: %v", err)
	}
	base := filepath.Dir(exePath)

	configPath := filepath.Join(base, "mor-launch.json")
	config, err := readConfig(configPath)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	binName := getBinName()
	filesToExtract := getFilesToExtract()
	modelFile := filepath.Join(base, config.ModelName)

	modelDownloadURL := fmt.Sprintf("%s/%s/%s/resolve/main/%s", config.ModelURL, config.ModelOwner, config.ModelRepo, config.ModelName)

	isLocalMode := len(os.Args) > 1 && os.Args[1] == "local"

	runLlamaServer := false

	if isLocalMode {
		filesNeeded := make([]string, len(filesToExtract))
		for i, f := range filesToExtract {
			filesNeeded[i] = filepath.Join(base, filepath.Base(f))
		}

		if !filesExist(filesNeeded...) {
			log.Println("Required binaries not found. Downloading...")
			llamaZip := filepath.Join(base, fmt.Sprintf("%s-%s.zip", config.LlamaFileBase, binName))
			llamaDownloadURL := fmt.Sprintf("%s/%s/%s-%s.zip", config.LlamaURL, config.LlamaRelease, config.LlamaFileBase, binName)

			if err := downloadFile(llamaZip, llamaDownloadURL); err != nil {
				log.Fatalf("Failed to download Llama binary: %v", err)
			}

			if err := extractFilesFromZip(llamaZip, filesToExtract, base); err != nil {
				log.Fatalf("Failed to extract binaries: %v", err)
			}
		}

		if !filesExist(modelFile) {
			log.Println("Model file not found. Downloading...")
			if err := downloadFile(modelFile, modelDownloadURL); err != nil {
				log.Fatalf("Failed to download model file: %v", err)
			}
		}

		runLlamaServer = true
	} else if filesExist(filepath.Join(base, "llama-server"), modelFile) || filesExist(filepath.Join(base, "llama-server.exe"), modelFile) {
		runLlamaServer = true
	}

	config.Run = replaceModelPlaceholder(config.Run, config.ModelName)

	var wg sync.WaitGroup
	for _, cmdStr := range config.Run {
		if !runLlamaServer && strings.Contains(cmdStr, "llama-server") {
			log.Println("Skipping llama-server command based on startup mode.")
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
