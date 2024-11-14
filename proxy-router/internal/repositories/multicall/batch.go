package multicall

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/multicall3"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Batch executes multiple calls to the same method on the same contract in a single multicall and converts the results to the specified type
// TODO: configure max batch size, enabling multicall level and json-rpc level batching
func Batch[T any](ctx context.Context, mc MulticallBackend, contractABI *abi.ABI, addr common.Address, method string, callsArgs [][]interface{}) ([]T, error) {
	var calls []multicall3.Multicall3Call

	for _, args := range callsArgs {
		calldata, err := contractABI.Pack(method, args...)
		if err != nil {
			return nil, err
		}
		calls = append(calls, multicall3.Multicall3Call{
			Target:   addr,
			CallData: calldata,
		})
	}

	_, res, err := mc.Aggregate(ctx, calls)
	if err != nil {
		return nil, err
	}

	sessions := make([]T, len(res))
	for i, result := range res {
		var data interface{}
		err := contractABI.UnpackIntoInterface(&data, method, result)
		if err != nil {
			return nil, err
		}

		sessions[i] = *abi.ConvertType(data, new(T)).(*T)
	}

	return sessions, nil
}
