package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// NormalizeJson returns normalized json message, without spaces and newlines
func NormalizeJson(msg []byte) ([]byte, error) {
	var a interface{}
	err := json.Unmarshal(msg, &a)
	if err != nil {
		return nil, err
	}
	return json.Marshal(a)
}

func MustUnmarshallString(value json.RawMessage) string {
	var res string
	err := json.Unmarshal(value, &res)
	if err != nil {
		panic(err)
	}
	return res
}

func ReadJSONFile(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Read the file contents
	fileContents, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	// Validate if the content is valid JSON
	var js json.RawMessage
	if err := json.Unmarshal(fileContents, &js); err != nil {
		return "", fmt.Errorf("invalid JSON: %v", err)
	}

	return string(fileContents), nil
}
