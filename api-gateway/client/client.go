package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ApiGatewayClient struct {
	BaseURL string
	Client  *http.Client
}

// Helper function to make GET requests
func (c *ApiGatewayClient) getRequest(ctx context.Context, endpoint string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// Helper function to make POST requests
func (c *ApiGatewayClient) postRequest(ctx context.Context, endpoint string, body interface{}, result interface{}) error {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+endpoint, reader)
	if err != nil {
		return err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *ApiGatewayClient) GetConfig(ctx context.Context) (interface{}, error) {
	var result interface{}
	err := c.getRequest(ctx, "/config", &result)
	return result, err
}

func (c *ApiGatewayClient) GetFiles(ctx context.Context) (int, interface{}) {
	var result interface{}
	err := c.getRequest(ctx, "/files", &result)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}
	return http.StatusOK, result
}

func (c *ApiGatewayClient) HealthCheck(ctx context.Context) (interface{}, error) {
	var result interface{}
	err := c.getRequest(ctx, "/healthcheck", &result)
	return result, err
}

func (c *ApiGatewayClient) InitiateSession(ctx context.Context) (int, interface{}) {
	var result interface{}
	err := c.postRequest(ctx, "/proxy/sessions/initiate", nil, &result)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}
	return http.StatusOK, result
}

func (c *ApiGatewayClient) SendPrompt(ctx context.Context) (bool, int, interface{}) {
	var result interface{}
	err := c.postRequest(ctx, "/proxy/sessions/:id/prompt", nil, &result)
	if err != nil {
		return false, http.StatusInternalServerError, err.Error()
	}
	return true, http.StatusOK, result
}

func (c *ApiGatewayClient) Prompt(ctx context.Context, req interface{}) (interface{}, error) {
	var result interface{}
	err := c.postRequest(ctx, "/v1/chat/completions", req, &result)
	return result, err
}

func (c *ApiGatewayClient) PromptStream(ctx context.Context, req interface{}, flush interface{}) (interface{}, error) {
	// Implementation for streaming responses can be complex and may require handling HTTP/2 or WebSockets
	return nil, errors.New("streaming not implemented")
}

func (c *ApiGatewayClient) GetLatestBlock(ctx context.Context) (uint64, error) {
	var result uint64
	err := c.getRequest(ctx, "/blockchain/latestBlock", &result)
	return result, err
}

func (c *ApiGatewayClient) GetAllProviders(ctx context.Context) (int, gin.H) {
	var result gin.H
	err := c.getRequest(ctx, "/blockchain/providers", &result)
	if err != nil {
		return http.StatusInternalServerError, gin.H{"error": err.Error()}
	}
	return http.StatusOK, result
}

func (c *ApiGatewayClient) GetAllModels(ctx context.Context) (int, gin.H) {
	var result gin.H
	err := c.getRequest(ctx, "/blockchain/models", &result)
	if err != nil {
		return http.StatusInternalServerError, gin.H{"error": err.Error()}
	}
	return http.StatusOK, result
}

func (c *ApiGatewayClient) GetBidsByProvider(ctx context.Context, providerAddr string, offset *big.Int, limit uint8) (int, gin.H) {
	var result gin.H
	endpoint := fmt.Sprintf("/blockchain/providers/%s/bids?offset=%s&limit=%d", providerAddr, offset.String(), limit)
	err := c.getRequest(ctx, endpoint, &result)
	if err != nil {
		return http.StatusInternalServerError, gin.H{"error": err.Error()}
	}
	return http.StatusOK, result
}

func (c *ApiGatewayClient) GetBidsByModelAgent(ctx context.Context, modelAgentId [32]byte, offset *big.Int, limit uint8) (int, gin.H) {
	var result gin.H
	endpoint := fmt.Sprintf("/blockchain/models/%x/bids?offset=%s&limit=%d", modelAgentId, offset.String(), limit)
	err := c.getRequest(ctx, endpoint, &result)
	if err != nil {
		return http.StatusInternalServerError, gin.H{"error": err.Error()}
	}
	return http.StatusOK, result
}

func (c *ApiGatewayClient) OpenSession(ctx *gin.Context) (int, gin.H) {
	var result gin.H
	err := c.postRequest(ctx, "/blockchain/sessions", nil, &result)
	if err != nil {
		return http.StatusInternalServerError, gin.H{"error": err.Error()}
	}
	return http.StatusOK, result
}

func (c *ApiGatewayClient) CloseSession(ctx *gin.Context) (int, gin.H) {
	var result gin.H
	err := c.postRequest(ctx, "/blockchain/sessions/:id/close", nil, &result)
	if err != nil {
		return http.StatusInternalServerError, gin.H{"error": err.Error()}
	}
	return http.StatusOK, result
}