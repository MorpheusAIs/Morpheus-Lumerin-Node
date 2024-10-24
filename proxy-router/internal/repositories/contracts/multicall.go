package contracts

import (
	"context"
	"math/big"
	"sync"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/modelregistry"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/multicall3"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type Multicall struct {
	client    interfaces.EthClient
	multicall *multicall3.Multicall3
	model     *modelregistry.ModelRegistryCaller
}

func (m *Multicall) Multicall() {
	call1 := multicall3.Multicall3Call3{}
	m.multicall.Aggregate3(&bind.TransactOpts{})
}

func (m *Multicall) GetModels(ctx context.Context, modelIDs [][32]byte) error {
	mcclient := &MulticallCaller{}
	model, err := modelregistry.NewModelRegistryCaller(common.Address{}, mcclient)
	if err != nil {
		return err
	}
	for _, modelID := range modelIDs {
		model.GetModel(&bind.CallOpts{Context: ctx}, modelID)
	}
	return nil
}

type Result[T any] struct {
	Value T
	Err   error
}

func Promise[T any](fn func() (T, error)) <-chan Result[T] {
	ch := make(chan Result[T], 1)
	go func() {
		res, err := fn()
		ch <- Result[T]{res, err}
	}()
	return ch
}

type Call struct {
	data   *ethereum.CallMsg
	result []byte
	err    error
}

type MulticallCaller struct {
	calls []Call
	wg    sync.WaitGroup
}

func (mc *MulticallCaller) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	mc.wg.Add(1)
	c := Call{
		data: &call,
	}
	mc.calls = append(mc.calls, c)
	mc.wg.Wait()
	return c.result, c.err
}

func (mc *MulticallCaller) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	return nil, nil
}

func (mc *MulticallCaller) Flush()

var _ bind.ContractCaller = &MulticallCaller{}
