package structs

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
)

type RawEthTransactionResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

type TxType string

const (
	Transfer TxType = "Transfer"
	Approve  TxType = "Approve"
)

type MappedTransaction struct {
	Hash      common.Hash          `json:"hash"`
	From      common.Address       `json:"from"`
	To        common.Address       `json:"to"`
	Contract  *ContractInteraction `json:"contract"`
	Transfers []TokenTransfer      `json:"transfers"`
	Timestamp string               `json:"timestamp"`
}

type ContractInteraction struct {
	ContractAddress common.Address `json:"contractAddress"`
	ContractName    string         `json:"contractName"`
	MethodName      string         `json:"methodName"`
	DecodedInput    []InputEntry   `json:"decodedInput"`
}

type InputEntry struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type TokenTransfer struct {
	From          common.Address  `json:"from"`
	To            common.Address  `json:"to"`
	Value         string          `json:"value"`
	TokenAddress  *common.Address `json:"tokenAddress"` // nil for eth transfers
	TokenSymbol   string          `json:"tokenSymbol"`
	TokenName     string          `json:"tokenName"`
	TokenIcon     string          `json:"tokenIcon"`
	TokenDecimals int             `json:"tokenDecimals"`
}

type RawTransaction struct {
	BlockHash         string `json:"blockHash"`
	BlockNumber       string `json:"blockNumber"`
	Confirmations     string `json:"confirmations"`
	ContractAddress   string `json:"contractAddress"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`

	From             string `json:"from"`
	FunctionName     string `json:"functionName"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	GasPriceBid      string `json:"gasPriceBid"`
	GasUsed          string `json:"gasUsed"`
	Hash             string `json:"hash"`
	Input            string `json:"input"`
	IsError          string `json:"isError"`
	MethodId         string `json:"methodId"`
	Nonce            string `json:"nonce"`
	TimeStamp        string `json:"timeStamp"`
	To               string `json:"to"`
	TransactionIndex string `json:"transactionIndex"`
	TxReceiptStatus  string `json:"txreceipt_status"`
	Value            string `json:"value"`
}
