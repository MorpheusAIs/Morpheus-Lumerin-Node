package proxyapi

import (
	"context"
	"errors"
	"math/big"
	"testing"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

// mockCaller lets a test script CodeAt (EOA vs contract) and owner().
type mockCaller struct {
	code      []byte
	owner     common.Address
	rawReturn []byte // if set, returned verbatim from CallContract (overrides owner padding)
	callErr   error  // if set, CallContract fails with it (a revert or out of gas)
	codeCalls int
	callCalls int
}

func (m *mockCaller) CodeAt(_ context.Context, _ common.Address, _ *big.Int) ([]byte, error) {
	m.codeCalls++
	return m.code, nil
}

func (m *mockCaller) CallContract(_ context.Context, _ ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	m.callCalls++
	if m.callErr != nil {
		return nil, m.callErr
	}
	if m.rawReturn != nil {
		return m.rawReturn, nil
	}
	return common.LeftPadBytes(m.owner.Bytes(), 32), nil
}

func TestAuthorizedSigner_EOAReturnsSelf(t *testing.T) {
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")
	r := NewProviderAuthResolver(&mockCaller{code: nil}) // empty code = EOA
	got, err := r.AuthorizedSigner(context.Background(), provider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != provider {
		t.Fatalf("EOA: want %s, got %s", provider.Hex(), got.Hex())
	}
}

func TestAuthorizedSigner_ContractReturnsOwner(t *testing.T) {
	custody := common.HexToAddress("0x2222222222222222222222222222222222222222")
	owner := common.HexToAddress("0x3333333333333333333333333333333333333333")
	mc := &mockCaller{code: []byte{0x60, 0x80}, owner: owner} // non-empty code = contract
	r := NewProviderAuthResolver(mc)

	got, err := r.AuthorizedSigner(context.Background(), custody)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != owner {
		t.Fatalf("contract: want owner %s, got %s", owner.Hex(), got.Hex())
	}

	// cache: a second lookup must not re-hit the chain.
	if _, err := r.AuthorizedSigner(context.Background(), custody); err != nil {
		t.Fatal(err)
	}
	if mc.codeCalls != 1 || mc.callCalls != 1 {
		t.Fatalf("cache miss: codeCalls=%d callCalls=%d, want 1/1", mc.codeCalls, mc.callCalls)
	}
}

// owner() names the signer only when SessionSettlement._isValidProviderReceipt
// would accept it: exactly one 32-byte word holding a clean, non-zero address.
func TestAuthorizedSigner_OwnerReturnData(t *testing.T) {
	custody := common.HexToAddress("0x2222222222222222222222222222222222222222")
	owner := common.HexToAddress("0x4444444444444444444444444444444444444444")
	word := common.LeftPadBytes(owner.Bytes(), 32)
	dirty := common.LeftPadBytes(owner.Bytes(), 32)
	dirty[0] = 0x01

	for _, tc := range []struct {
		name    string
		ret     []byte
		callErr error
		want    common.Address // zero = rejected
	}{
		{"one clean word", word, nil, owner},
		{"longer than one word", append(common.CopyBytes(word), word...), nil, common.Address{}},
		{"shorter than one word", word[1:], nil, common.Address{}},
		{"dirty upper bytes", dirty, nil, common.Address{}},
		{"zero address", make([]byte, 32), nil, common.Address{}},
		{"call failed", nil, errors.New("out of gas"), common.Address{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := NewProviderAuthResolver(&mockCaller{code: []byte{0x60}, rawReturn: tc.ret, callErr: tc.callErr})
			got, err := r.AuthorizedSigner(context.Background(), custody)
			if got != tc.want || (err == nil) != (tc.want != common.Address{}) {
				t.Fatalf("got %s, err %v; want %s", got.Hex(), err, tc.want.Hex())
			}
		})
	}
}

// TestContractProviderSignatureFlow is the gate-the-gate: a morrpc frame signed
// by the operator (owner) key must validate against the RESOLVED signer and be
// rejected against the custody address and against an attacker key — exactly the
// distinction that was breaking session-open for contract providers.
func TestContractProviderSignatureFlow(t *testing.T) {
	operatorKey, _ := crypto.GenerateKey()
	operatorAddr := crypto.PubkeyToAddress(operatorKey.PublicKey)
	custody := common.HexToAddress("0x2222222222222222222222222222222222222222")
	attackerKey, _ := crypto.GenerateKey()

	// provider signs the raw keccak of the frame (morrpc transport signing).
	params := []byte(`{"id":"1","result":"pong"}`)
	hash := crypto.Keccak256Hash(params)
	sig, err := crypto.Sign(hash.Bytes(), operatorKey)
	if err != nil {
		t.Fatal(err)
	}
	attackerSig, _ := crypto.Sign(hash.Bytes(), attackerKey)

	resolver := NewProviderAuthResolver(&mockCaller{code: []byte{0x60, 0x80}, owner: operatorAddr})
	signer, err := resolver.AuthorizedSigner(context.Background(), custody)
	if err != nil {
		t.Fatal(err)
	}

	if !lib.VerifySignatureAddr(params, sig, signer) {
		t.Fatal("operator-signed frame must validate against the resolved signer")
	}
	if lib.VerifySignatureAddr(params, sig, custody) {
		t.Fatal("frame must NOT validate against the custody address (the old, broken check)")
	}
	if lib.VerifySignatureAddr(params, attackerSig, signer) {
		t.Fatal("attacker-signed frame must be rejected — MITM protection intact")
	}
}

func TestProxySender_AuthorizedSignerUsesResolver(t *testing.T) {
	custody := common.HexToAddress("0x2222222222222222222222222222222222222222")
	owner := common.HexToAddress("0x3333333333333333333333333333333333333333")
	p := &ProxyServiceSender{}

	got, err := p.authorizedSigner(context.Background(), custody)
	if err != nil {
		t.Fatal(err)
	}
	if got != custody {
		t.Fatalf("nil resolver: want provider %s, got %s", custody.Hex(), got.Hex())
	}

	p.SetProviderAuthResolver(NewProviderAuthResolver(&mockCaller{code: []byte{0x60, 0x80}, owner: owner}))
	got, err = p.authorizedSigner(context.Background(), custody)
	if err != nil {
		t.Fatal(err)
	}
	if got != owner {
		t.Fatalf("wired resolver: want owner %s, got %s", owner.Hex(), got.Hex())
	}
}
