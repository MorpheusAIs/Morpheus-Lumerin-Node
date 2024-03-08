import { createSelector } from 'reselect';

// Returns the "connectivity" state branch
export const getConnectivity = state => state.connectivity;

// Returns if the wallet is online or not (check reducer to see conditions)
export const getIsOnline = createSelector(
  getConnectivity,
  connectivity => connectivity.isOnline
);
