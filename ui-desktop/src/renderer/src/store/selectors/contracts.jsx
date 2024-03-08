import { createSelector } from 'reselect';
import sortBy from 'lodash/sortBy';

// Returns the array of transactions of the current chain/wallet/address.
// The items are mapped to contain properties useful for rendering.
export const getContracts = state => state.contracts;

export const getActiveContracts = createSelector(getContracts, contractsData =>
  sortBy(Object.values(contractsData.actives), 'timestamp')
);

export const getDraftContracts = createSelector(
  getContracts,
  contractsData => contractsData.drafts
);

export const getMergeAllContracts = createSelector(
  getActiveContracts,
  getDraftContracts,
  (activeContracts, draftContracts) =>
    [...activeContracts, ...draftContracts].reverse()
);

// Returns if the current wallet/address has transactions on the active chain
export const hasContracts = createSelector(
  getContracts,
  contractsData => Object.keys(contractsData.actives).length !== 0
);

export const getActiveContractsCount = createSelector(
  getActiveContracts,
  contractsData => contractsData.length
);

export const getDraftContractsCount = createSelector(
  getDraftContracts,
  contractsData => contractsData.length
);

// Returns wallet transactions sync status on the active chain
export const getContractsSyncStatus = createSelector(
  getContracts,
  contractsData => contractsData.syncStatus
);

export const getContractsLastUpdated = createSelector(
  getContracts,
  contractsData => contractsData.lastUpdated
);
