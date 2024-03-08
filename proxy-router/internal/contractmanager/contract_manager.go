package contractmanager

import (
	"context"
	"fmt"
	"sync"

	"github.com/Lumerin-protocol/contracts-go/clonefactory"
	"github.com/ethereum/go-ethereum/common"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/interfaces"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/repositories/contracts"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate"
)

type ContractManager struct {
	cfAddr    common.Address
	ownerAddr common.Address

	contracts   *lib.Collection[resources.Contract]
	contractsWG sync.WaitGroup

	createContract CreateContractFn
	store          *contracts.HashrateEthereum
	log            interfaces.ILogger
}

type CreateContractFn func(terms *hashrate.EncryptedTerms) (resources.Contract, error)

func NewContractManager(clonefactoryAddr, ownerAddr common.Address, createContractFn CreateContractFn, store *contracts.HashrateEthereum, log interfaces.ILogger) *ContractManager {
	return &ContractManager{
		cfAddr:         clonefactoryAddr,
		ownerAddr:      ownerAddr,
		contracts:      lib.NewCollection[resources.Contract](),
		createContract: createContractFn,
		store:          store,
		contractsWG:    sync.WaitGroup{},
		log:            log,
	}
}

func (cm *ContractManager) Run(ctx context.Context) error {
	defer func() {
		cm.log.Info("waiting for all contracts to stop")
		cm.contractsWG.Wait()
		cm.log.Info("all contracts stopped")
	}()

	contractIDs, err := cm.store.GetContractsIDs(ctx)
	if err != nil {
		return err
	}

	for _, id := range contractIDs {
		terms, err := cm.store.GetContract(ctx, id)
		if err != nil {
			return err
		}
		if cm.isOurContract(terms) {
			cm.AddContract(ctx, terms)
		}
	}

	sub, err := cm.store.CreateCloneFactorySubscription(ctx, cm.cfAddr)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()
	cm.log.Infof("subscribed to clonefactory events at address %s", cm.cfAddr.Hex())

	for {
		select {
		case <-ctx.Done():
			//TODO: wait until all child contracts are stopped
			return nil
		case event := <-sub.Events():
			err := cm.cloneFactoryController(ctx, event)
			if err != nil {
				return err
			}
		case err := <-sub.Err():
			return err
		}
	}
}

func (cm *ContractManager) cloneFactoryController(ctx context.Context, event interface{}) error {
	switch e := event.(type) {
	case *clonefactory.ClonefactoryContractCreated:
		return cm.handleContractCreated(ctx, e)
	case *clonefactory.ClonefactoryClonefactoryContractPurchased:
		return cm.handleContractPurchased(ctx, e)
	case *clonefactory.ClonefactoryContractDeleteUpdated:
		return cm.handleContractDeleteUpdated(ctx, e)
	}
	return nil
}

func (cm *ContractManager) handleContractCreated(ctx context.Context, event *clonefactory.ClonefactoryContractCreated) error {
	terms, err := cm.store.GetContract(ctx, event.Address.Hex())
	if err != nil {
		return err
	}
	if cm.isOurContract(terms) {
		cm.AddContract(ctx, terms)
	}
	return nil
}

func (cm *ContractManager) handleContractPurchased(ctx context.Context, event *clonefactory.ClonefactoryClonefactoryContractPurchased) error {
	cm.log.Debugf("clonefactory contract purchased event, address %s", event.Address.Hex())
	terms, err := cm.store.GetContract(ctx, event.Address.Hex())
	if err != nil {
		return err
	}
	if terms.Buyer() == cm.ownerAddr.String() || terms.Validator() == cm.ownerAddr.String() {
		cm.AddContract(ctx, terms)
	}
	return nil
}

func (cm *ContractManager) handleContractDeleteUpdated(ctx context.Context, event *clonefactory.ClonefactoryContractDeleteUpdated) error {
	ctr, ok := cm.contracts.Load(event.Address.Hex())
	if !ok {
		return nil
	}
	err := ctr.SyncState(ctx)
	if err != nil {
		cm.log.Errorf("contract sync state error %s", err)
	}
	return nil
}

func (cm *ContractManager) AddContract(ctx context.Context, data *hashrate.EncryptedTerms) {
	_, ok := cm.contracts.Load(data.ID())
	if ok {
		cm.log.Errorw("contract already exists in store", "CtrAddr", lib.AddrShort(data.ID()))
		return
	}

	cntr, err := cm.createContract(data)
	if err != nil {
		cm.log.Errorw("contract factory error", "err", err, "CtrAddr", lib.AddrShort(data.ID()))
		return
	}

	cm.contracts.Store(cntr)

	cm.contractsWG.Add(1)
	go func() {
		defer cm.contractsWG.Done()

		err := cntr.Run(ctx)
		cm.log.Warnw(fmt.Sprintf("exited from contract %s", err), "CtrAddr", lib.AddrShort(data.ID()))

		cm.contracts.Delete(cntr.ID())
	}()
}

func (cm *ContractManager) GetContracts() *lib.Collection[resources.Contract] {
	return cm.contracts
}

func (cm *ContractManager) GetContract(id string) (resources.Contract, bool) {
	return cm.contracts.Load(id)
}

func (cm *ContractManager) isOurContract(terms TermsCommon) bool {
	return terms.Seller() == cm.ownerAddr.String() || terms.Buyer() == cm.ownerAddr.String() || terms.Validator() == cm.ownerAddr.String()
}
