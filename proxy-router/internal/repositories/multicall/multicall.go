package multicall

import (
	"context"
	"fmt"
	"math/big"

	mc3 "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/multicall3"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type Multicall3Custom struct {
	mutlicall3Addr common.Address
	multicall3Abi  *abi.ABI
	client         bind.ContractCaller
}

var MULTICALL3_ADDR = common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")

func NewMulticall3(client bind.ContractCaller) *Multicall3Custom {
	return NewMulticall3Custom(client, MULTICALL3_ADDR)
}

func NewMulticall3Custom(client bind.ContractCaller, multicall3Addr common.Address) *Multicall3Custom {
	multicall3Abi, err := mc3.Multicall3MetaData.GetAbi()
	if err != nil {
		panic("invalid multicall3 ABI: " + err.Error())
	}
	return &Multicall3Custom{
		mutlicall3Addr: multicall3Addr,
		multicall3Abi:  multicall3Abi,
		client:         client,
	}
}

func (m *Multicall3Custom) Aggregate(ctx context.Context, calls []mc3.Multicall3Call) (blockNumer *big.Int, returnData [][]byte, err error) {
	res, err := m.aggregate(ctx, calls, "aggregate")
	if err != nil {
		return nil, nil, err
	}

	blockNumer, ok := res[0].(*big.Int)
	if !ok {
		return nil, nil, fmt.Errorf("Failed to parse block number")
	}
	returnData, ok = res[1].([][]byte)
	if !ok {
		return nil, nil, fmt.Errorf("Failed to parse return data")
	}

	return blockNumer, returnData, nil
}

func (m *Multicall3Custom) Aggregate3(ctx context.Context, calls []mc3.Multicall3Call3) ([]mc3.Multicall3Result, error) {
	res, err := m.aggregate(ctx, calls, "aggregate3")
	if err != nil {
		return nil, err
	}

	parsed, ok := res[0].([]struct {
		Success    bool    "json:\"success\""
		ReturnData []uint8 "json:\"returnData\""
	})
	if !ok {
		return nil, fmt.Errorf("Failed to parse result")
	}

	var results []mc3.Multicall3Result
	for _, vv := range parsed {
		results = append(results, mc3.Multicall3Result{
			Success:    vv.Success,
			ReturnData: vv.ReturnData,
		})
	}

	return results, nil
}

func (m *Multicall3Custom) aggregate(ctx context.Context, calls interface{}, method string) ([]interface{}, error) {
	calldata, err := m.multicall3Abi.Pack(method, calls)
	if err != nil {
		return nil, fmt.Errorf("Failed to pack data: %w", err)
	}

	result, err := m.client.CallContract(ctx, ethereum.CallMsg{
		To:   &m.mutlicall3Addr,
		Data: calldata,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to call contract: %w", err)
	}

	parsedResult, err := m.multicall3Abi.Unpack(method, result)
	if err != nil {
		return nil, fmt.Errorf("Failed to unpack result: %w", err)
	}

	return parsedResult, nil
}

var _ MulticallBackend = &Multicall3Custom{}
