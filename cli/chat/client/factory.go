package client

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type GetEnvFunc func(key string) string

func NewApiGatewayClientFromEnv(getEnv GetEnvFunc) (*ApiGatewayClient, error) {
	apiHost := getEnv("API_HOST")
	if apiHost == "" {
		apiHost = "http://localhost:8082"
	}

	client := NewApiGatewayClient(apiHost, http.DefaultClient)

	cookiePath, err := client.GetCookiePath(context.Background())
	if err != nil {
		return nil, fmt.Errorf("can't get cookie path: %v", err)
	}

	data, err := os.ReadFile(cookiePath)
	if err != nil {
		return nil, fmt.Errorf("can't read cookie file: %v", err)
	}

	line := strings.TrimSpace(string(data))
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid cookie format")
	}

	login, pass := parts[0], parts[1]
	client.SetAuth(login, pass)

	return client, nil
}
