package contract

import (
	"context"
	"errors"
	"time"

	"github.com/Lumerin-protocol/contracts-go/implementation"
	"github.com/ethereum/go-ethereum/common"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/repositories/contracts"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources"

	hashrateContract "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate"
)

type ControllerSeller struct {
	*ContractWatcherSellerV2

	syncStateCh chan struct{}
	store       *contracts.HashrateEthereum
	privKey     string
}

func NewControllerSeller(contract *ContractWatcherSellerV2, store *contracts.HashrateEthereum, privKey string) *ControllerSeller {
	return &ControllerSeller{
		syncStateCh:             make(chan struct{}, 1),
		ContractWatcherSellerV2: contract,
		store:                   store,
		privKey:                 privKey,
	}
}

func (c *ControllerSeller) Run(ctx context.Context) error {
	defer func() {
		_ = c.log.Close()
	}()

	sub, err := c.store.CreateImplementationSubscription(ctx, common.HexToAddress(c.ID()))
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	c.log.Infof("started watching contract as seller, address %s", c.ID())

	if c.ShouldBeRunning() {
		go func() {
			select {
			case <-ctx.Done():
				return
			case sub.Ch() <- &implementation.ImplementationContractPurchased{}:
			}
		}()
	}

	for {
		select {
		case <-c.syncStateCh:
			err := c.LoadTermsFromBlockchain(ctx)
			if err != nil {
				c.log.Errorf("error loading terms: %s", err)
				c.contractErr.Store(err)
			} else {
				c.contractErr.Store(nil)
			}

		case event := <-sub.Events():
			err := c.controller(ctx, event)
			if err != nil {
				c.log.Errorf("error handling event %T: %s", event, err)
				c.contractErr.Store(err)
			} else {
				c.contractErr.Store(nil)
			}
		case err := <-sub.Err():
			if c.IsRunning() {
				c.ContractWatcherSellerV2.StopFulfilling()
				c.log.Infof("waiting for contract watcher to stop")
				<-c.ContractWatcherSellerV2.Done()
				c.log.Infof("contract watcher stopped")
			}
			return err
		case <-ctx.Done():
			c.log.Infof("context done, stopping contract watcher")
			if c.IsRunning() {
				c.ContractWatcherSellerV2.StopFulfilling()
				c.log.Infof("waiting for contract watcher to stop")
				<-c.ContractWatcherSellerV2.Done()
				c.log.Infof("contract watcher stopped")
			}
			return ctx.Err()
		case <-c.ContractWatcherSellerV2.Done():
			err := c.ContractWatcherSellerV2.Err()
			if err != nil {
				// fulfillment error, buyer will close on underdelivery
				c.log.Warnf("seller contract ended with error: %s", err)
				c.ContractWatcherSellerV2.Reset()
				continue
			}

			// no error, seller closes the contract after expiration
			c.log.Infof("seller contract ended without error")

			waitBeforeClose := 10 * time.Second
			c.log.Infof("sleeping %s", waitBeforeClose)
			time.Sleep(waitBeforeClose)

			c.log.Infof("closing contract id %s, startsAt %s, duration %s, elapsed %s", c.ID(), c.StartTime(), c.Duration(), c.Elapsed())
			err = c.store.CloseContract(ctx, c.ID(), contracts.CloseoutTypeWithoutClaim, c.privKey)
			if errors.Is(err, contracts.ErrNotRunning) {
				c.log.Infof("contract is not running, nothing to close")
				c.ContractWatcherSellerV2.Reset()
				continue
			}
			if err != nil {
				c.log.Errorf("error closing contract: %s", err)
				continue
			}

			c.log.Warnf("seller contract closed")
			c.ContractWatcherSellerV2.Reset()
		}
	}
}

func (c *ControllerSeller) controller(ctx context.Context, event interface{}) error {
	switch e := event.(type) {
	case *implementation.ImplementationContractPurchased:
		return c.handleContractPurchased(ctx, e)
	case *implementation.ImplementationContractClosed:
		return c.handleContractClosed(ctx, e)
	case *implementation.ImplementationCipherTextUpdated:
		return c.handleCipherTextUpdated(ctx, e)
	case *implementation.ImplementationPurchaseInfoUpdated:
		return c.handlePurchaseInfoUpdated(ctx, e)
	}
	return nil
}

func (c *ControllerSeller) handleContractPurchased(ctx context.Context, event *implementation.ImplementationContractPurchased) error {
	c.log.Debugf("got purchased event for contract")
	if c.State() == resources.ContractStateRunning {
		c.log.Infof("contract is running, ignore")
		return nil
	}

	err := c.LoadTermsFromBlockchain(ctx)
	if err != nil {
		return err
	}

	if !c.ShouldBeRunning() {
		c.log.Infof("contract should not be running, ignore")
		return nil
	}

	c.ContractWatcherSellerV2.Reset()
	err = c.StartFulfilling()
	if err != nil {
		c.log.Errorf("error handleContractPurchased: %s", err)
	}

	return nil
}

func (c *ControllerSeller) handleContractClosed(ctx context.Context, event *implementation.ImplementationContractClosed) error {
	c.log.Warnf("got closed event for contract")

	if c.IsRunning() {
		c.log.Infof("contract is running, stopping")
		c.StopFulfilling()
		<-c.Done()
	}

	err := c.LoadTermsFromBlockchain(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c *ControllerSeller) handleCipherTextUpdated(ctx context.Context, event *implementation.ImplementationCipherTextUpdated) error {
	c.log.Debugf("got cipherTextUpdated event for contract")

	currentDest := c.Dest()

	terms, err := c.GetTermsFromBlockchain(ctx)
	if err != nil {
		// if we cannot decrypt dest, we still update terms with nil dest
		// and stop fulfilling
		if c.IsRunning() {
			c.ContractWatcherSellerV2.StopFulfilling()
			<-c.ContractWatcherSellerV2.Done()
		}
		c.SetTerms(terms)
		return err
	}

	//TODO: drop protocol before comparison
	newDest := terms.Dest().String()

	if currentDest == newDest {
		return nil
	}
	if c.IsRunning() {
		c.ContractWatcherSellerV2.StopFulfilling()
		<-c.ContractWatcherSellerV2.Done()
	}
	c.SetTerms(terms)
	err = c.ContractWatcherSellerV2.StartFulfilling()
	if err != nil {
		c.log.Errorf("error handleCipherTextUpdated: %s", err)
	}
	return nil
}

func (c *ControllerSeller) handlePurchaseInfoUpdated(ctx context.Context, event *implementation.ImplementationPurchaseInfoUpdated) error {
	c.log.Debugf("got purchaseInfoUpdated event for contract")

	err := c.LoadTermsFromBlockchain(ctx)
	if err != nil {
		return err
	}

	return nil
}

// LoadTermsFromBlockchain loads terms from blockchain and decrypts them, if decryption fails, still updates terms with nil dest
func (c *ControllerSeller) LoadTermsFromBlockchain(ctx context.Context) error {
	encryptedTerms, err := c.store.GetContract(ctx, c.ID())

	if err != nil {
		return err
	}

	terms, err := encryptedTerms.Decrypt(c.privKey)
	c.SetTerms(terms)

	return err
}

func (c *ControllerSeller) GetTermsFromBlockchain(ctx context.Context) (*hashrateContract.Terms, error) {
	encryptedTerms, err := c.store.GetContract(ctx, c.ID())

	if err != nil {
		return nil, err
	}

	terms, err := encryptedTerms.Decrypt(c.privKey)

	if err != nil {
		c.log.Errorf("error decrypting terms: %s", err)
		return terms, err // return retrieved terms even if error occured during decryption
	}

	return terms, nil
}

func (c *ControllerSeller) SyncState(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.syncStateCh <- struct{}{}:
	}
	return nil
}
