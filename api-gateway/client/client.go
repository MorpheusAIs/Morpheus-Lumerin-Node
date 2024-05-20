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

func NewApiGatewayClient(baseURL string, httpClient *http.Client) *ApiGatewayClient {
	return &ApiGatewayClient{
		BaseURL: baseURL,
		Client:  httpClient,
	}
}

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
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
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

	if result == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *ApiGatewayClient) GetProxyRouterConfig(ctx context.Context) (interface{}, error) {
	var result interface{}
	err := c.getRequest(ctx, "/config", &result)
	return result, err
}

func (c *ApiGatewayClient) GetProxyRouterFiles(ctx context.Context) (interface{}, error) {
	var result interface{}
	err := c.getRequest(ctx, "/files", &result)
	if err != nil {
		return nil, fmt.Errorf("internal error: %v; http status: %v", err, http.StatusInternalServerError)
	}
	return result, nil
}

func (c *ApiGatewayClient) HealthCheck(ctx context.Context) (interface{}, error) {
	var result interface{}
	err := c.getRequest(ctx, "/healthcheck", &result)
	return result, err
}

func (c *ApiGatewayClient) InitiateSession(ctx context.Context) (interface{}, error) {
	var result interface{}
	err := c.postRequest(ctx, "/proxy/sessions/initiate", nil, &result)

	if err != nil {
		return nil, fmt.Errorf("internal error: %v; http status: %v", err, http.StatusInternalServerError)
	}

	return result, nil
}

func (c *ApiGatewayClient) SendPrompt(ctx context.Context, prompt string, messages []string) (interface{}, error) {
	var result interface{}
	err := c.postRequest(ctx, "/proxy/sessions/:id/prompt", nil, &result)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	return result, nil
}

func (c *ApiGatewayClient) Prompt(ctx context.Context, req interface{}) (interface{}, error) {
	var result interface{}
	err := c.postRequest(ctx, "/v1/chat/completions", req, &result)
	return result, err
}

func (c *ApiGatewayClient) PromptStream(ctx context.Context, req interface{}, flush interface{}) (interface{}, error) {

	return nil, errors.New("streaming not implemented")
}

func (c *ApiGatewayClient) GetLatestBlock(ctx context.Context) (result uint64, err error) {

	err = c.getRequest(ctx, "/blockchain/latestBlock", &result)
	return result, err
}

func (c *ApiGatewayClient) GetAllProviders(ctx context.Context) (result []string, err error) {

	err = c.getRequest(ctx, "/blockchain/providers", &result)

	if err != nil {
		return nil, fmt.Errorf("internal error: %v; http status: %v", err, http.StatusInternalServerError)
	}

	return result, nil
}

func (c *ApiGatewayClient) GetAllModels(ctx context.Context) (result []string, err error) {

	err = c.getRequest(ctx, "/blockchain/models", &result)

	if err != nil {
		return nil, fmt.Errorf("internal error: %v; http status: %v", err, http.StatusInternalServerError)
	}

	return result, nil
}

func (c *ApiGatewayClient) GetBidsByProvider(ctx context.Context, providerAddr string, offset *big.Int, limit uint8) (bids []string, err error) {

	endpoint := fmt.Sprintf("/blockchain/providers/%s/bids?offset=%s&limit=%d", providerAddr, offset.String(), limit)
	err = c.getRequest(ctx, endpoint, &bids)
	if err != nil {
		return nil, fmt.Errorf("internal error: %v; http status: %v", err, http.StatusInternalServerError)
	}
	return bids, err
}

func (c *ApiGatewayClient) GetBidsByModelAgent(ctx context.Context, modelAgentId [32]byte, offset *big.Int, limit uint8) (result []string, err error) {

	endpoint := fmt.Sprintf("/blockchain/models/%x/bids?offset=%s&limit=%d", modelAgentId, offset.String(), limit)
	err = c.getRequest(ctx, endpoint, &result)

	if err != nil {
		return nil, fmt.Errorf("internal error: %v; http status: %v", err, http.StatusInternalServerError)
	}

	return result, err
}

func (c *ApiGatewayClient) OpenSession(ctx *gin.Context) (err error) {

	err = c.postRequest(ctx, "/blockchain/sessions", nil, nil)

	if err != nil {
		return fmt.Errorf("internal error: %v; http status: %v", err, http.StatusInternalServerError)
	}

	return nil
}

func (c *ApiGatewayClient) CloseSession(ctx *gin.Context) error {
	err := c.postRequest(ctx, "/blockchain/sessions/:id/close", nil, nil)

	if err != nil {
		return fmt.Errorf("internal error: %v; http status: %v", err, http.StatusInternalServerError)
	}

	return nil
}
