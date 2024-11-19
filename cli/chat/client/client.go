package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sashabaranov/go-openai"
)

func NewApiGatewayClient(baseURL string, httpClient *http.Client) *ApiGatewayClient {

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &ApiGatewayClient{
		BaseURL:    baseURL,
		HttpClient: httpClient,
	}
}

type ApiGatewayClient struct {
	BaseURL    string
	HttpClient *http.Client
}

// Helper function to make GET requests
func (c *ApiGatewayClient) getRequest(ctx context.Context, endpoint string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+endpoint, nil)

	if err != nil {
		return err
	}

	resp, err := c.HttpClient.Do(req)

	if err != nil {
		fmt.Println("http client error: ", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("unexpected status code: %d", resp.StatusCode)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *ApiGatewayClient) requestChatCompletionStream(ctx context.Context, endpoint string, request *openai.ChatCompletionRequest, callback CompletionCallback, modelId string, sessionId string) (*openai.ChatCompletionStreamResponse, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if sessionId != "" {
		req.Header.Set("session_id", sessionId)
	} else if modelId != "" {
		req.Header.Set("model_id", modelId)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		// Handle the completion of the stream
		if line == "data: [DONE]" {
			completion := &openai.ChatCompletionStreamResponse{
				Choices: []openai.ChatCompletionStreamChoice{
					{
						Delta: openai.ChatCompletionStreamChoiceDelta{
							Content: "[DONE]",
						},
					},
				},
			}

			return completion, nil
		}

		if strings.HasPrefix(line, "data: ") {
			data := line[6:] // Skip the "data: " prefix
			var completion *openai.ChatCompletionStreamResponse
			if err := json.Unmarshal([]byte(data), &completion); err != nil {
				fmt.Printf("Error decoding response: %v\n", err)
			}

			if completion.ID != "" {
				callback(completion)
			} else {
				var completion map[string]interface{}
				if err := json.Unmarshal([]byte(data), &completion); err != nil {
					fmt.Printf("Error decoding response: %v\n", err)
				}
				callback(completion)
			}

		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading stream: %v", err)
	}

	return nil, err
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
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		if err != nil {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, errResp["error"])
	}

	if resp == nil {
		return nil
	}

	if result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
	}

	return err
}

func (c *ApiGatewayClient) GetProxyRouterConfig(ctx context.Context) (interface{}, error) {
	var result map[string]interface{}
	err := c.getRequest(ctx, "/config", &result)
	return result, err
}

func (c *ApiGatewayClient) GetProxyRouterFiles(ctx context.Context) (interface{}, error) {
	var result map[string]interface{}
	err := c.getRequest(ctx, "/files", &result)
	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}
	return result, nil
}

func (c *ApiGatewayClient) HealthCheck(ctx context.Context) (interface{}, error) {
	var result map[string]interface{}
	err := c.getRequest(ctx, "/healthcheck", &result)

	return result, err
}

func (c *ApiGatewayClient) InitiateSession(ctx context.Context) (interface{}, error) {
	var result map[string]interface{}
	err := c.postRequest(ctx, "/proxy/sessions/initiate", nil, &result)

	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}

	return result, nil
}

func (c *ApiGatewayClient) PromptStream(ctx context.Context, message string, messagesContext []openai.ChatCompletionMessage, modelId string, sessionId string, flush CompletionCallback) (interface{}, error) {
	messagesContext = append(messagesContext, openai.ChatCompletionMessage{
		Role:    "user",
		Content: message,
	})

	request := &openai.ChatCompletionRequest{
		Messages: messagesContext,
		Stream:   true,
	}

	return c.requestChatCompletionStream(ctx, "/v1/chat/completions", request, flush, modelId, sessionId)
}

func (c *ApiGatewayClient) GetLatestBlock(ctx context.Context) (result uint64, err error) {

	err = c.getRequest(ctx, "/blockchain/latestBlock", &result)
	return result, err
}

func (c *ApiGatewayClient) GetAllProviders(ctx context.Context) (result map[string]interface{}, err error) {

	err = c.getRequest(ctx, "/blockchain/providers", &result)

	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}

	return result, nil
}

func (c *ApiGatewayClient) CreateNewProvider(ctx context.Context, addStake uint64, endpoint string) (result interface{}, err error) {
	request := struct {
		AddStake uint64 `json:"stake"`
		Endpoint string `json:"endpoint"`
	}{addStake, endpoint}

	err = c.postRequest(ctx, "/blockchain/providers", &request, &result)

	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}

	return result, nil
}

func (c *ApiGatewayClient) CreateNewModel(ctx context.Context, name string, ipfsID string, stake uint64, fee uint64, tags []string) (result map[string]interface{}, err error) {
	request := struct {
		Stake  uint64
		IpfsID string
		Name   string
		Fee    uint64
		Tags   []string
	}{stake, ipfsID, name, fee, tags}

	err = c.postRequest(ctx, "/blockchain/models", &request, &result)

	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}

	return result, nil
}

func (c *ApiGatewayClient) CreateNewProviderBid(ctx context.Context, model string, pricePerSecond uint64) (result map[string]interface{}, err error) {
	request := struct {
		Model          string `json:"modelId"`
		PricePerSecond uint64 `json:"pricePerSecond"`
	}{model, pricePerSecond}

	err = c.postRequest(ctx, "/blockchain/bids", &request, &result)

	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}

	return result, nil
}

func (c *ApiGatewayClient) GetAllModels(ctx context.Context) (result map[string]interface{}, err error) {

	err = c.getRequest(ctx, "/blockchain/models", &result)

	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}

	return result, nil
}

func (c *ApiGatewayClient) GetBidsByProvider(ctx context.Context, providerAddr string, offset *big.Int, limit uint8) (bids interface{}, err error) {
	endpoint := fmt.Sprintf("/blockchain/providers/%s/bids?offset=%s&limit=%d", providerAddr, offset.String(), limit)
	err = c.getRequest(ctx, endpoint, &bids)
	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}
	return bids, err
}

func (c *ApiGatewayClient) GetBidsByModelAgent(ctx context.Context, modelAgentId string, offset string, limit string) (result map[string]interface{}, err error) {

	endpoint := fmt.Sprintf("/blockchain/models/%s/bids?offset=%s&limit=%s", modelAgentId, offset, limit)
	err = c.getRequest(ctx, endpoint, &result)

	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}

	return result, err
}

func (c *ApiGatewayClient) ListUserSessions(ctx context.Context, user string) (result []SessionListItem, err error) {
	response := map[string][]SessionListItem{}

	err = c.getRequest(ctx, fmt.Sprintf("/blockchain/sessions/user?user=%s", user), &response)
	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}

	return response["sessions"], nil
}

func (c *ApiGatewayClient) ListProviderSessions(ctx context.Context, provider string) (result []SessionListItem, err error) {
	response := map[string][]SessionListItem{}

	err = c.getRequest(ctx, fmt.Sprintf("/blockchain/sessions/provider?provider=%s", provider), &response)
	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}

	return response["sessions"], nil
}

func (c *ApiGatewayClient) OpenStakeSession(ctx context.Context, req *SessionStakeRequest) (session *Session, err error) {

	err = c.postRequest(ctx, "/blockchain/sessions", req, session)

	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}

	return session, nil
}

func (c *ApiGatewayClient) OpenSession(ctx context.Context, req *SessionRequest) (session *Session, err error) {

	session = &Session{}
	err = c.postRequest(ctx, fmt.Sprintf("/blockchain/models/%s/session", req.ModelId), req, session)

	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}

	return session, nil
}

func (c *ApiGatewayClient) GetLocalModels(ctx context.Context) (models *[]interface{}, err error) {
	err = c.getRequest(ctx, "/v1/models", &models)

	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}

	return models, nil
}

func (c *ApiGatewayClient) CloseSession(ctx context.Context, sessionId string) error {

	err := c.postRequest(ctx, fmt.Sprintf("/blockchain/sessions/%s/close", sessionId), nil, nil)

	if err != nil {
		return fmt.Errorf("internal error: %v", err)
	}

	return nil
}

func (c *ApiGatewayClient) GetAllowance(ctx context.Context, spender string) (map[string]interface{}, error) {
	var result map[string]interface{}
	endpoint := fmt.Sprintf("/blockchain/allowance?spender=%s", spender)
	err := c.getRequest(ctx, endpoint, &result)
	return result, err
}

func (c *ApiGatewayClient) ApproveAllowance(ctx context.Context, spender string, amount uint64) (map[string]interface{}, error) {
	var result map[string]interface{}
	endpoint := fmt.Sprintf("/blockchain/approve?spender=%s&amount=%d", spender, amount)
	err := c.postRequest(ctx, endpoint, nil, &result)
	return result, err
}

func (c *ApiGatewayClient) CreateWallet(ctx context.Context, privateKey string) error {
	return c.postRequest(ctx, "/wallet", &WalletRequest{PrivateKey: privateKey}, nil)
}

type WalletResponse struct {
	Address string `json:"address"`
}

func (c *ApiGatewayClient) GetWallet(ctx context.Context) (*WalletResponse, error) {
	response := &WalletResponse{}

	err := c.getRequest(ctx, "/wallet", response)

	if err != nil {
		return nil, fmt.Errorf("internal error: %v", err)
	}

	return response, nil
}

func (c *ApiGatewayClient) GetBalance(ctx context.Context) (eth string, mor string, err error) {
	response := map[string]string{}

	err = c.getRequest(ctx, "/blockchain/balance", &response)
	if err != nil {
		return "", "", fmt.Errorf("internal error: %v", err)
	}

	return response["eth"], response["mor"], nil
}

func (c *ApiGatewayClient) GetDiamondAddress(ctx context.Context) (common.Address, error) {
	response := struct {
		Config struct {
			Marketplace struct {
				DiamondContractAddress string
			}
		}
	}{}

	err := c.getRequest(ctx, "/config", &response)
	if err != nil {
		return common.Address{}, fmt.Errorf("internal error: %v", err)
	}

	return common.HexToAddress(response.Config.Marketplace.DiamondContractAddress), nil
}
