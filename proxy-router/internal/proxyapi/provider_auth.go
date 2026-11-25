package proxyapi

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"sync"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

// ProviderAuthResolver maps an on-chain provider address to the address whose
// secp256k1 key authenticates that provider's morrpc frames.
//
// For an EOA provider that is the address itself. For a *contract* provider
// (e.g. a custody contract, which holds no private key) morrpc frames are signed
// by the contract's owner() — the same key SessionSettlement._isValidProviderReceipt
// accepts via its contract-owner branch. Resolving the owner here lets a consumer
// authenticate a contract provider WITHOUT weakening the check: the owner is read
// from the consumer's own trusted RPC, keyed by the exact provider address the
// consumer chose, so a man-in-the-middle still cannot impersonate the provider.
type ProviderAuthResolver interface {
	AuthorizedSigner(ctx context.Context, providerAddr common.Address) (common.Address, error)
}

// ownerSelector is keccak256("owner()")[:4].
var ownerSelector = crypto.Keccak256([]byte("owner()"))[:4]

// providerOwnerStaticcallGas is SessionSettlement's
// PROVIDER_OWNER_STATICCALL_GAS: the gas _isValidProviderReceipt forwards to
// owner().
const providerOwnerStaticcallGas = 100_000

// ownerCallGas is the gas limit of the owner() eth_call. An eth_call pays the
// intrinsic gas of a transaction out of its limit first: the base cost and
// the 4 calldata bytes of the selector, none of them zero. Adding both leaves
// owner() exactly the gas the contract forwards to it.
const ownerCallGas = providerOwnerStaticcallGas + params.TxGas + 4*params.TxDataNonZeroGasEIP2028

// ethCaller is the minimal read-only on-chain surface the resolver needs;
// *ethclient.Client satisfies it.
type ethCaller interface {
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
	CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
}

// OnChainProviderAuthResolver resolves the authorized signer via eth_getCode +
// owner(), caching results. A provider's owner is effectively static for a
// session's lifetime, and the on-chain approval remains the ultimate money
// authority, so a cached value cannot be used to steal funds.
type OnChainProviderAuthResolver struct {
	client ethCaller
	mu     sync.RWMutex
	cache  map[common.Address]common.Address
}

func NewProviderAuthResolver(client ethCaller) *OnChainProviderAuthResolver {
	return &OnChainProviderAuthResolver{client: client, cache: make(map[common.Address]common.Address)}
}

func (r *OnChainProviderAuthResolver) AuthorizedSigner(ctx context.Context, providerAddr common.Address) (common.Address, error) {
	r.mu.RLock()
	cached, ok := r.cache[providerAddr]
	r.mu.RUnlock()
	if ok {
		return cached, nil
	}

	signer, err := r.resolve(ctx, providerAddr)
	if err != nil {
		return common.Address{}, err
	}

	r.mu.Lock()
	r.cache[providerAddr] = signer
	r.mu.Unlock()
	return signer, nil
}

func (r *OnChainProviderAuthResolver) resolve(ctx context.Context, providerAddr common.Address) (common.Address, error) {
	code, err := r.client.CodeAt(ctx, providerAddr, nil)
	if err != nil {
		return common.Address{}, fmt.Errorf("code at provider %s: %w", providerAddr.Hex(), err)
	}
	if len(code) == 0 {
		return providerAddr, nil // EOA provider signs as itself
	}
	out, err := r.client.CallContract(ctx, ethereum.CallMsg{To: &providerAddr, Data: ownerSelector, Gas: ownerCallGas}, nil)
	if err != nil {
		return common.Address{}, fmt.Errorf("call owner() on %s: %w", providerAddr.Hex(), err)
	}
	// Accept only what SessionSettlement._isValidProviderReceipt accepts: exactly
	// one 32-byte word whose upper 12 bytes are zero. Any other answer names no
	// owner the chain would trust.
	if len(out) != 32 {
		return common.Address{}, fmt.Errorf("owner() on %s returned %d bytes, want 32", providerAddr.Hex(), len(out))
	}
	if !bytes.Equal(out[:12], make([]byte, 12)) {
		return common.Address{}, fmt.Errorf("owner() on %s returned a word that is not an address", providerAddr.Hex())
	}
	owner := common.BytesToAddress(out)
	if owner == (common.Address{}) {
		return common.Address{}, fmt.Errorf("owner() on %s is the zero address", providerAddr.Hex())
	}
	return owner, nil
}
