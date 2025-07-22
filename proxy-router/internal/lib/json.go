package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
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

func StripKnownKeys(m map[string]json.RawMessage, t reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		if comma := strings.Index(tag, ","); comma != -1 {
			tag = tag[:comma]
		}
		delete(m, tag)
	}
}

func ToJSONFragment(s string) json.RawMessage {
	// If it already parses as JSON (number, bool, object, array, null) keep it.
	if json.Valid([]byte(s)) {
		return json.RawMessage(s)
	}
	// Otherwise treat it as a string.
	return json.RawMessage(strconv.Quote(s))
}

// helper: build set of keys defined in struct tags
func FormTagSet(t reflect.Type) map[string]struct{} {
	set := make(map[string]struct{})
	for i := 0; i < t.NumField(); i++ {
		if tag, ok := t.Field(i).Tag.Lookup("form"); ok && tag != "-" {
			name, _, _ := strings.Cut(tag, ",")
			if name == "" {
				name = t.Field(i).Name
			}
			set[name] = struct{}{}
		}
	}
	return set
}
