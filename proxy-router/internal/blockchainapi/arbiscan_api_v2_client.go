package blockchainapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

type ArbiscanApiV2Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
	log     lib.ILogger
}

const (
	BaseURL = "https://arbitrum.blockscout.com/api/v2"
)

func NewBlockscoutApiV2Client(baseURL string, log lib.ILogger) *ArbiscanApiV2Client {
	return NewBlockscoutApiV2ClientWithApiKey(baseURL, "", log)
}

func NewBlockscoutApiV2ClientWithApiKey(baseURL, apiKey string, log lib.ILogger) *ArbiscanApiV2Client {
	return &ArbiscanApiV2Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{},
		log:     log,
	}
}

func (a *ArbiscanApiV2Client) GetLastTransactions(ctx context.Context, address common.Address) ([]structs.MappedTransaction, error) {
	transactions, err := a.getLastTransactions(ctx, address)
	if err != nil {
		return nil, err
	}

	transfers, err := a.getLastTransfers(ctx, address)
	if err != nil {
		a.log.Error(err)
		return nil, err
	}

	return mergeTx(transactions, transfers), nil
}

func mergeTx(txs []structs.MappedTransaction, tr []structs.MappedTransaction) []structs.MappedTransaction {
	transfersMap := make(map[common.Hash][]structs.MappedTransaction)
	for _, t := range tr {
		if _, ok := transfersMap[t.Hash]; !ok {
			transfersMap[t.Hash] = make([]structs.MappedTransaction, 0)
		}
		transfersMap[t.Hash] = append(transfersMap[t.Hash], t)
	}

	for i, tx := range txs {
		if transfers, ok := transfersMap[tx.Hash]; ok {
			for _, t := range transfers {
				txs[i].Transfers = append(txs[i].Transfers, t.Transfers...)
			}
		}
	}

	return txs
}

func (a *ArbiscanApiV2Client) getLastTransactions(ctx context.Context, address common.Address) ([]structs.MappedTransaction, error) {
	url := fmt.Sprintf("%s/addresses/%s/transactions", a.baseURL, address.Hex())
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := a.client.Do(req)
	if err != nil {
		a.log.Error(err)

		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		a.log.Error(err)

		return nil, err
	}

	transactionsRes := &TransactionsRes{}

	err = json.Unmarshal(body, transactionsRes)
	if err != nil {
		a.log.Error(err)

		return nil, err
	}

	txs := make([]structs.MappedTransaction, len(transactionsRes.Items))

	for i, tx := range transactionsRes.Items {
		mapped := structs.MappedTransaction{
			Hash:      common.HexToHash(tx.Hash),
			From:      common.HexToAddress(tx.From.Hash),
			To:        common.HexToAddress(tx.To.Hash),
			Timestamp: tx.Timestamp,
		}

		if tx.To.IsContract {
			methodNameRaw, input, err := lib.DecodeInput(common.FromHex(tx.RawInput))
			if err != nil {
				a.log.Warnf("cannot decode transaction input for api response, method %s, input %s", tx.Method, tx.RawInput)
			}

			mapped.Contract = &structs.ContractInteraction{
				ContractAddress: common.HexToAddress(tx.To.Hash),
				ContractName:    tx.To.Name,
				MethodName:      methodNameRaw,   // decode it
				DecodedInput:    mapInput(input), // decode it
			}
		} else {
			mapped.Transfers = []structs.TokenTransfer{
				{
					From:          common.HexToAddress(tx.From.Hash),
					To:            common.HexToAddress(tx.To.Hash),
					Value:         tx.Value,
					TokenAddress:  nil,
					TokenSymbol:   "ETH",
					TokenName:     "Ethereum",
					TokenIcon:     "",
					TokenDecimals: 18,
				},
			}
		}

		txs[i] = mapped
	}

	return txs, nil
}

func (a *ArbiscanApiV2Client) getLastTransfers(ctx context.Context, address common.Address) ([]structs.MappedTransaction, error) {
	url := fmt.Sprintf("%s/addresses/%s/token-transfers", a.baseURL, address.Hex())
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	transactionsRes := &TokenTransferRes{}

	err = json.Unmarshal(body, transactionsRes)
	if err != nil {
		a.log.Error(err)
		return nil, err
	}

	txs := make([]structs.MappedTransaction, len(transactionsRes.Items))

	for i, tx := range transactionsRes.Items {
		mapped := structs.MappedTransaction{
			Hash:      common.HexToHash(tx.TxHash),
			From:      common.HexToAddress(tx.From.Hash),
			To:        common.HexToAddress(tx.To.Hash),
			Timestamp: tx.Timestamp,
		}

		if tx.To.IsContract {
			mapped.Contract = &structs.ContractInteraction{
				ContractAddress: common.HexToAddress(tx.To.Hash),
				ContractName:    tx.To.Name,
			}
		}
		tokenAddress := common.HexToAddress(tx.Token.Address)
		decimals, _ := strconv.Atoi(tx.Token.Decimals)
		mapped.Transfers = []structs.TokenTransfer{
			{
				From:          common.HexToAddress(tx.From.Hash),
				To:            common.HexToAddress(tx.To.Hash),
				Value:         tx.Total.Value,
				TokenAddress:  &tokenAddress,
				TokenSymbol:   tx.Token.Symbol,
				TokenName:     tx.Token.Name,
				TokenIcon:     tx.Token.IconURL,
				TokenDecimals: decimals,
			},
		}
		txs[i] = mapped
	}

	return txs, nil
}

func mapInput(input []lib.InputEntry) []structs.InputEntry {
	entries := make([]structs.InputEntry, len(input))
	for i, in := range input {
		entries[i] = structs.InputEntry{
			Key:   in.Key,
			Type:  in.Type,
			Value: in.Value,
		}
	}
	return entries
}

// Example response
// https://arbitrum.blockscout.com/api-docs#operations-default-get_address_txs

type TransactionsRes struct {
	Items []Transaction `json:"items"`
}

type Transaction struct {
	Confirmations   int         `json:"confirmations"`
	Result          string      `json:"result"`        // "Error: (Awaiting internal transactions for reason)"
	Hash            string      `json:"hash"`          // "0x5d90a9da2b8da402b11bc92c8011ec8a62a2d59da5c7ac4ae0f73ec51bb73368"
	RevertReason    string      `json:"revert_reason"` // "Error: (Awaiting internal transactions for reason)"
	Type            int         `json:"type"`
	CreatedContract interface{} `json:"created_contract"`
	Value           string      `json:"value"`
	From            Address     `json:"from"`
	To              Address     `json:"to"`
	Status          string      `json:"status"`
	Method          string      `json:"method"`
	Timestamp       string      `json:"timestamp"`
	ExchangeRate    string      `json:"exchange_rate"`
	BlockNumber     int         `json:"block_number"`
	RawInput        string      `json:"raw_input"`
	// GasUsed                        string        `json:"gas_used"`
	// GasLimit                       string        `json:"gas_limit"`
	// GasPrice                       string        `json:"gas_price"`
	// MaxPriorityFeePerGas           string        `json:"max_priority_fee_per_gas"`
	// MaxFeePerGas                   string        `json:"max_fee_per_gas"`
	// PriorityFee                    string        `json:"priority_fee"`
	// TransactionBurntFee            string        `json:"transaction_burnt_fee"`
	// TxBurntFee                     string        `json:"tx_burnt_fee"`
	// Fee                            Fee           `json:"fee"`
	// BaseFeePerGas                  string        `json:"base_fee_per_gas"`
	// Nonce                          int           `json:"nonce"`

	// ConfirmationDuration           []int         `json:"confirmation_duration"`
	// TokenTransfersOverflow         interface{}   `json:"token_transfers_overflow"`
	// Position                       int           `json:"position"`
	// TransactionTag                 string        `json:"transaction_tag"`
	// AuthorizationList              []interface{} `json:"authorization_list"`
	// Actions                        []interface{} `json:"actions"`
	// HasErrorInInternalTransactions bool          `json:"has_error_in_internal_transactions"`
	// DecodedInput                   DecodedInput  `json:"decoded_input"`
	// TransactionTypes               []string      `json:"transaction_types"`
	// TokenTransfers                 interface{}   `json:"token_transfers"`
}

type Address struct {
	EnsDomainName string `json:"ens_domain_name"`
	Hash          string `json:"hash"` // "0xEb533ee5687044E622C69c58B1B12329F56eD9ad"
	Name          string `json:"name"`
	// Implementations []interface{} `json:"implementations"`
	IsContract bool `json:"is_contract"`
	// IsScam          bool          `json:"is_scam"`
	// IsVerified      bool          `json:"is_verified"`
	// Metadata        interface{}   `json:"metadata"`
	// PrivateTags     []interface{} `json:"private_tags"`
	// ProxyType       interface{}   `json:"proxy_type"`
	// PublicTags      []interface{} `json:"public_tags"`
	// WatchlistNames  []interface{} `json:"watchlist_names"`
}

type TokenTransferRes struct {
	Items []TokenTransfer `json:"items"`
}

type TokenTransfer struct {
	From     Address   `json:"from"`
	LogIndex int       `json:"log_index"`
	Method   string    `json:"method"`
	To       Address   `json:"to"`
	Token    TokenInfo `json:"token"`
	Total    struct {
		Decimals string `json:"decimals"`
		Value    string `json:"value"`
	}
	TxHash    string `json:"transaction_hash"`
	Timestamp string `json:"timestamp"`
}

type TokenInfo struct {
	IconURL      string  `json:"icon_url"`
	Name         string  `json:"name"`
	Decimals     string  `json:"decimals"`
	Symbol       string  `json:"symbol"`
	Address      string  `json:"address"`
	ExchangeRate float64 `json:"exchange_rate"`
}
