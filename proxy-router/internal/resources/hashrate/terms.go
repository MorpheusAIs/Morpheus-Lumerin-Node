package hashrate

import (
	"fmt"
	"math/big"
	"net/url"
	"time"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
)

var (
	ErrInvalidDestURL    = fmt.Errorf("invalid url")
	ErrCannotDecryptDest = fmt.Errorf("cannot decrypt")
)

// Terms holds the terms of the contract where destination is decrypted
type Terms struct {
	BaseTerms
	ValidatorURL   *url.URL
	DestinationURL *url.URL
}

func (p *Terms) Dest() *url.URL {
	return lib.CopyURL(p.ValidatorURL)
}

func (p *Terms) PoolDest() *url.URL {
	return lib.CopyURL(p.DestinationURL)
}

func (t *Terms) Encrypt(privateKey string) (*Terms, error) {
	var destUrl *url.URL

	if t.ValidatorURL != nil {
		dest, err := lib.EncryptString(t.ValidatorURL.String(), privateKey)
		if err != nil {
			return nil, err
		}

		destUrl, err = url.Parse(dest)
		if err != nil {
			return nil, err
		}
	} else {
		destUrl = nil
	}

	return &Terms{
		BaseTerms:    *t.Copy(),
		ValidatorURL: destUrl,
	}, nil
}

// EncryptedTerms holds the terms of the contract where destination is encrypted
type EncryptedTerms struct {
	BaseTerms
	ValidatorUrlEncrypted string
	DestEncrypted         string
}

func NewTerms(contractID, seller, buyer string, startsAt time.Time, duration time.Duration, hashrate float64, price *big.Int, profitTarget int8, state BlockchainState, isDeleted bool, balance *big.Int, hasFutureTerms bool, version uint32, validatorUrlEncrypted string, destEncrypted string, validatorAddress string) *EncryptedTerms {
	return &EncryptedTerms{
		BaseTerms: BaseTerms{
			contractID:     contractID,
			seller:         seller,
			buyer:          buyer,
			validator:      validatorAddress,
			startsAt:       startsAt,
			duration:       duration,
			hashrate:       hashrate,
			price:          price,
			profitTarget:   profitTarget,
			state:          state,
			isDeleted:      isDeleted,
			balance:        balance,
			hasFutureTerms: hasFutureTerms,
			version:        version,
		},
		ValidatorUrlEncrypted: validatorUrlEncrypted,
		DestEncrypted:         destEncrypted,
	}
}

// Decrypt decrypts the validator url, if error returns the terms with dest set to nil and error
func (t *EncryptedTerms) Decrypt(privateKey string) (*Terms, error) {
	var (
		returnErr error
	)

	terms := &Terms{
		BaseTerms:    *t.Copy(),
		ValidatorURL: nil,
	}

	if t.ValidatorUrlEncrypted == "" {
		return terms, nil
	}

	dest, err := lib.DecryptString(t.ValidatorUrlEncrypted, privateKey)
	if err != nil {
		return terms, lib.WrapError(ErrCannotDecryptDest, fmt.Errorf("%s: %s", err, t.ValidatorUrlEncrypted))
	}

	destUrl, err := url.Parse(dest)
	if err != nil {
		return terms, lib.WrapError(ErrInvalidDestURL, fmt.Errorf("%s: %s", err, dest))
	}

	terms.ValidatorURL = destUrl
	return terms, returnErr
}

// Decrypt decrypts the destination pool url, if error returns the terms with dest set to nil and error
func (t *EncryptedTerms) DecryptPoolDest(privateKey string) (*Terms, error) {
	var (
		returnErr error
	)

	terms := &Terms{
		BaseTerms:    *t.Copy(),
		ValidatorURL: nil,
	}

	if t.DestEncrypted == "" {
		return terms, nil
	}

	dest, err := lib.DecryptString(t.DestEncrypted, privateKey)
	if err != nil {
		return terms, lib.WrapError(ErrCannotDecryptDest, fmt.Errorf("%s: %s", err, t.DestEncrypted))
	}

	destUrl, err := url.Parse(dest)
	if err != nil {
		return terms, lib.WrapError(ErrInvalidDestURL, fmt.Errorf("%s: %s", err, dest))
	}

	terms.DestinationURL = destUrl
	return terms, returnErr
}

// BaseTerms holds the terms of the contract with common methods for both encrypted and decrypted terms
type BaseTerms struct {
	contractID     string
	seller         string
	buyer          string
	validator      string
	startsAt       time.Time
	duration       time.Duration
	hashrate       float64
	price          *big.Int
	profitTarget   int8
	state          BlockchainState
	isDeleted      bool
	balance        *big.Int
	hasFutureTerms bool
	version        uint32
}

func (b *BaseTerms) ID() string {
	return b.contractID
}

func (b *BaseTerms) Seller() string {
	return b.seller
}

func (b *BaseTerms) Buyer() string {
	return b.buyer
}

func (b *BaseTerms) Validator() string {
	return b.validator
}

func (b *BaseTerms) StartTime() time.Time {
	return b.startsAt
}

func (p *BaseTerms) EndTime() time.Time {
	if p.startsAt.IsZero() {
		return time.Time{}
	}
	endTime := p.startsAt.Add(p.duration)
	return endTime
}

func (p *BaseTerms) Elapsed() time.Duration {
	if p.startsAt.IsZero() {
		return 0
	}
	return time.Since(p.startsAt)
}

func (b *BaseTerms) Duration() time.Duration {
	return b.duration
}

func (b *BaseTerms) HashrateGHS() float64 {
	return b.hashrate
}

func (b *BaseTerms) Price() *big.Int {
	return new(big.Int).Set(b.price) // copy
}

func (b *BaseTerms) ProfitTarget() int8 {
	return b.profitTarget
}

// PriceLMR returns price in LMR without decimals
func (b *BaseTerms) PriceLMR() float64 {
	price, _ := lib.NewRat(b.Price(), big.NewInt(1e8)).Float64()
	return price
}

func (p *BaseTerms) BlockchainState() BlockchainState {
	return p.state
}

func (b *BaseTerms) IsDeleted() bool {
	return b.isDeleted
}

func (b *BaseTerms) Balance() *big.Int {
	return new(big.Int).Set(b.balance) // copy
}

func (b *BaseTerms) HasFutureTerms() bool {
	return b.hasFutureTerms
}

func (b *BaseTerms) Version() uint32 {
	return b.version
}

func (b *BaseTerms) SetState(state BlockchainState) {
	b.state = state
}

func (b *BaseTerms) Copy() *BaseTerms {
	return &BaseTerms{
		contractID:     b.ID(),
		seller:         b.Seller(),
		buyer:          b.Buyer(),
		validator:      b.Validator(),
		startsAt:       b.StartTime(),
		duration:       b.Duration(),
		hashrate:       b.HashrateGHS(),
		state:          b.BlockchainState(),
		price:          b.Price(),
		isDeleted:      b.IsDeleted(),
		balance:        b.Balance(),
		hasFutureTerms: b.HasFutureTerms(),
		version:        b.Version(),
	}
}
