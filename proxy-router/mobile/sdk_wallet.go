package mobile

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	wallet "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/wallet"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tyler-smith/go-bip39"
)

// CreateWallet generates a new BIP-39 mnemonic and derives the wallet.
// Returns the mnemonic (back it up!) and the Ethereum address.
func (s *SDK) CreateWallet() (mnemonic string, address string, err error) {
	entropy, err := bip39.NewEntropy(128) // 12 words
	if err != nil {
		return "", "", fmt.Errorf("generate entropy: %w", err)
	}
	mnemonic, err = bip39.NewMnemonic(entropy)
	if err != nil {
		return "", "", fmt.Errorf("generate mnemonic: %w", err)
	}
	err = s.wallet.SetMnemonic(mnemonic, DefaultDerivationPath)
	if err != nil {
		return "", "", fmt.Errorf("set mnemonic: %w", err)
	}
	addr, err := s.getAddress()
	if err != nil {
		return "", "", err
	}
	s.log.Infof("wallet created: %s", addr)
	return mnemonic, addr, nil
}

// ImportMnemonic imports a wallet from an existing BIP-39 mnemonic.
func (s *SDK) ImportMnemonic(mnemonic string) (address string, err error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return "", fmt.Errorf("invalid mnemonic")
	}
	err = s.wallet.SetMnemonic(mnemonic, DefaultDerivationPath)
	if err != nil {
		return "", fmt.Errorf("set mnemonic: %w", err)
	}
	return s.getAddress()
}

// ImportPrivateKey imports a wallet from a hex-encoded private key.
func (s *SDK) ImportPrivateKey(hexKey string) (address string, err error) {
	pk, err := lib.StringToHexString(hexKey)
	if err != nil {
		return "", fmt.Errorf("parse private key: %w", err)
	}
	err = s.wallet.SetPrivateKey(pk)
	if err != nil {
		return "", fmt.Errorf("set private key: %w", err)
	}
	return s.getAddress()
}

// VerifyMnemonicMatchesCurrent returns true if the mnemonic derives the same address as the loaded wallet (no mutation).
func (s *SDK) VerifyMnemonicMatchesCurrent(mnemonic string) (bool, error) {
	mnemonic = strings.TrimSpace(mnemonic)
	if !bip39.IsMnemonicValid(mnemonic) {
		return false, fmt.Errorf("invalid mnemonic")
	}
	current, err := s.getAddress()
	if err != nil {
		return false, err
	}
	w, err := wallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return false, err
	}
	path, err := accounts.ParseDerivationPath(DefaultDerivationPath)
	if err != nil {
		return false, err
	}
	pk, err := w.DerivePrivateKey(path)
	if err != nil {
		return false, err
	}
	addr, err := lib.PrivKeyToAddr(pk)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(addr.Hex(), current), nil
}

// VerifyPrivateKeyMatchesCurrent returns true if the hex private key matches the loaded wallet (no mutation).
func (s *SDK) VerifyPrivateKeyMatchesCurrent(hexKey string) (bool, error) {
	current, err := s.getAddress()
	if err != nil {
		return false, err
	}
	pk, err := lib.StringToHexString(hexKey)
	if err != nil {
		return false, err
	}
	addr, err := lib.PrivKeyBytesToAddr(pk)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(addr.Hex(), current), nil
}

// ExportPrivateKey returns the private key as a hex string.
func (s *SDK) ExportPrivateKey() (string, error) {
	pk, err := s.wallet.GetPrivateKey()
	if err != nil {
		return "", err
	}
	return pk.Hex(), nil
}

// GetAddress returns the wallet's Ethereum address.
func (s *SDK) GetAddress() (string, error) {
	return s.getAddress()
}

func (s *SDK) getAddress() (string, error) {
	pk, err := s.wallet.GetPrivateKey()
	if err != nil {
		return "", fmt.Errorf("get private key: %w", err)
	}
	addr, err := lib.PrivKeyBytesToAddr(pk)
	if err != nil {
		return "", fmt.Errorf("derive address: %w", err)
	}
	return addr.Hex(), nil
}

// GetBalance returns ETH and MOR balances as decimal strings (in wei).
func (s *SDK) GetBalance(ctx context.Context) (*Balance, error) {
	if err := s.checkClosed(); err != nil {
		return nil, err
	}
	eth, mor, err := s.blockchain.GetBalance(ctx)
	if err != nil {
		return nil, err
	}
	return &Balance{
		ETH: eth.String(),
		MOR: mor.String(),
	}, nil
}

// GetBalanceJSON returns the balance as a JSON string (for FFI).
func (s *SDK) GetBalanceJSON(ctx context.Context) (string, error) {
	b, err := s.GetBalance(ctx)
	if err != nil {
		return "", err
	}
	return toJSON(b)
}

// SendETH sends native ETH (amount in wei, decimal string) to an 0x address. Waits for mining.
func (s *SDK) SendETH(ctx context.Context, toHex string, amountWei string) (txHash string, err error) {
	if err := s.checkClosed(); err != nil {
		return "", err
	}
	if !common.IsHexAddress(toHex) {
		return "", fmt.Errorf("invalid recipient address")
	}
	to := common.HexToAddress(toHex)
	amt, ok := new(big.Int).SetString(amountWei, 10)
	if !ok || amt.Sign() <= 0 {
		return "", fmt.Errorf("invalid amount: use wei as a decimal string")
	}
	h, err := s.blockchain.SendETH(ctx, to, amt, "")
	if err != nil {
		return "", err
	}
	return h.Hex(), nil
}

// SendMOR sends MOR ERC-20 (amount in token wei, 18 decimals, decimal string) to an 0x address.
func (s *SDK) SendMOR(ctx context.Context, toHex string, amountWei string) (txHash string, err error) {
	if err := s.checkClosed(); err != nil {
		return "", err
	}
	if !common.IsHexAddress(toHex) {
		return "", fmt.Errorf("invalid recipient address")
	}
	to := common.HexToAddress(toHex)
	amt, ok := new(big.Int).SetString(amountWei, 10)
	if !ok || amt.Sign() <= 0 {
		return "", fmt.Errorf("invalid amount: use wei as a decimal string")
	}
	h, err := s.blockchain.SendMOR(ctx, to, amt, "")
	if err != nil {
		return "", err
	}
	return h.Hex(), nil
}
