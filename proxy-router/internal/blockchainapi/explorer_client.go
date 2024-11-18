package blockchainapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/ethereum/go-ethereum/common"
)

type ExplorerClient struct {
	explorerApiUrl string
	morTokenAddr   common.Address
	retryDelay     time.Duration
	maxRetries     uint8
}

func NewExplorerClient(explorerApiUrl string, morTokenAddr common.Address, retryDelay time.Duration, maxRetries uint8) *ExplorerClient {
	return &ExplorerClient{
		explorerApiUrl: explorerApiUrl,
		morTokenAddr:   morTokenAddr,
		retryDelay:     retryDelay,
		maxRetries:     maxRetries,
	}
}

func (e *ExplorerClient) GetEthTransactions(ctx context.Context, address common.Address, page uint64, limit uint8) ([]structs.RawTransaction, error) {
	query := fmt.Sprintf("?module=account&action=txlist&address=%s&page=%d&offset=%d&sort=desc", address.Hex(), page, limit)
	url := e.explorerApiUrl + query

	transactions, err := e.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (e *ExplorerClient) GetTokenTransactions(ctx context.Context, address common.Address, page uint64, limit uint8) ([]structs.RawTransaction, error) {
	query := fmt.Sprintf("?module=account&action=tokentx&contractaddress=%s&address=%s&page=%d&offset=%d&sort=desc", e.morTokenAddr.Hex(), address.Hex(), page, limit)
	url := e.explorerApiUrl + query

	transactions, err := e.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (e *ExplorerClient) doRequest(ctx context.Context, url string) ([]structs.RawTransaction, error) {
	return e.doRequestRetry(ctx, url, e.maxRetries)
}

func (e *ExplorerClient) doRequestRetry(ctx context.Context, url string, retriesLeft uint8) ([]structs.RawTransaction, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch transactions, HTTP status code: %d", resp.StatusCode)
	}

	var response structs.RawEthTransactionResponse

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode transactions response: url %s, err %w", url, err)
	}

	if response.Status == "0" {
		if response.Message == "No transactions found" {
			return make([]structs.RawTransaction, 0), nil
		}
		if retriesLeft == 0 {
			return nil, fmt.Errorf("failed to fetch transactions, response message: %s, result %s", response.Message, string(response.Result))
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(e.retryDelay):
		}
		return e.doRequestRetry(ctx, url, retriesLeft-1)
	}
	if response.Status != "1" {
		return nil, fmt.Errorf("failed to fetch transactions, response status: %s", response.Result)
	}

	var transactions []structs.RawTransaction
	if err := json.Unmarshal(response.Result, &transactions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transactions: %w", err)
	}

	return transactions, nil
}
