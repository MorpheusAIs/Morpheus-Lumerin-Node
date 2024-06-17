package rpcproxy

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Lumerin-protocol/Morpheus-Lumerin-Node/proxy-router/internal/rpcproxy/structs"
)

type ExplorerClient struct {
	explorerApiUrl string
	morTokenAddr   string
}

func NewExplorerClient(explorerApiUrl string, morTokenAddr string) *ExplorerClient {
	return &ExplorerClient{
		explorerApiUrl: explorerApiUrl,
		morTokenAddr:   morTokenAddr,
	}
}

func (e *ExplorerClient) GetEthTransactions(address string, page string, limit string) ([]structs.RawTransaction, error) {
	query := fmt.Sprintf("?module=account&action=txlist&address=%s&page=%s&offset=%s", address, page, limit)
	url := e.explorerApiUrl + query

	transactions, err := e.doRequest(url)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (e *ExplorerClient) GetTokenTransactions(address string, page string, limit string) ([]structs.RawTransaction, error) {
	query := fmt.Sprintf("?module=account&action=tokentx&contractaddress=%s&address=%s&page=%s&offset=%s", e.morTokenAddr, address, page, limit)
	url := e.explorerApiUrl + query

	transactions, err := e.doRequest(url)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (e *ExplorerClient) doRequest(url string) ([]structs.RawTransaction, error) {
	req, err := http.NewRequest("GET", url, nil)
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
		return nil, fmt.Errorf("failed to fetch eth transactions, HTTP status code: %d", resp.StatusCode)
	}

	var response structs.RawEthTransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if response.Status == "0" {
		return make([]structs.RawTransaction, 0), nil
	}

	if response.Status != "1" {
		return nil, fmt.Errorf("failed to fetch eth transactions, response status: %s", response.Result)
	}

	return response.Result, nil
}
