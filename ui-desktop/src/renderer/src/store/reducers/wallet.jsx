//@ts-check
import { handleActions } from 'redux-actions';
import get from 'lodash/get';

export const initialState = {
  syncStatus: 'up-to-date',
  allowSendTransaction: true,
  isActive: false,
  address: '',
  ethBalance: 0,
  transactions: {},
  page: 1,
  pageSize: 15,
  hasNextPage: true,
  token: {
    contract: '',
    lmrBalance: 0,
    transactions: {},
    symbol: 'LMR',
    symbolEth: 'ETH'
  },
  fee: ''
};

/**
 * Should filter transactions without receipt if we received ones
 */
const mergeTransactions = (stateTxs, payloadTxs) => {
  const txWithReceipts = payloadTxs.filter(tx => tx.receipt);
  const newStateTxs = { ...stateTxs };

  for (const tx of txWithReceipts) {
    const key = `${tx.transaction.hash}_${tx.receipt.tokenSymbol || 'ETH'}`;
    const oldStateTx = stateTxs[key];

    const isDifferentLogIndex =
      oldStateTx?.transaction?.logIndex &&
      tx?.transaction?.logIndex &&
      oldStateTx?.transaction?.logIndex !== tx?.transaction?.logIndex; // means that this is a second transaction within the same hash

    if (oldStateTx && !isDifferentLogIndex) {
      continue;
    }
    newStateTxs[key] = tx;
    // contract purchase emits 2 transactions with the same hash
    // as of now we merge corresponding amount values. Temporary fix, until refactoring trasactions totally

    // we sum transaction value if it is transfers within the same transaction, but with different logIndex
    // TODO: display both transactions in the UI either separately or as a single one with two outputs
    if (oldStateTx && isDifferentLogIndex) {
      if (
        newStateTxs[key].transaction.value &&
        oldStateTx.transaction.logIndex !== tx.transaction.logIndex
      ) {
        newStateTxs[key].transaction.value = String(
          Number(oldStateTx.transaction.value) + Number(tx.transaction.value)
        );
      }

      if (newStateTxs[key].transaction.input.amount) {
        newStateTxs[key].transaction.input.amount = String(
          Number(oldStateTx.transaction.input.amount) +
            Number(tx.transaction.input.amount)
        );
      }

      if (newStateTxs[key].receipt.value) {
        newStateTxs[key].receipt.value = String(
          Number(oldStateTx.receipt.value) + Number(tx.receipt.value)
        );
      }
    }
  }
  return newStateTxs;
};

const reducer = handleActions(
  {
    'initial-state-received': (state, { payload }) => ({
      ...state,
      ...get(payload, 'wallet', initialState),
      token: get(payload, 'wallet.token', initialState.token)
    }),

    'create-wallet': (state, { payload }) => ({
      ...state,
      address: payload.address
    }),

    'open-wallet': (state, { payload }) => ({
      ...state,
      address: payload.address,
      isActive: payload.isActive
    }),

    'eth-balance-changed': (state, { payload }) => ({
      ...state,
      ethBalance: payload.ethBalance
    }),

    'token-balance-changed': (state, { payload }) => ({
      ...state,
      token: {
        ...state.token,
        lmrBalance: payload.lmrBalance
      }
    }),

    'token-contract-received': (state, { payload }) => ({
      ...state,
      token: {
        ...state.token,
        contract: payload.contract
      }
    }),

    'token-transactions-changed': (state, { payload }) => ({
      ...state,
      token: {
        ...state.token,
        transactions: mergeTransactions(
          state.token.transactions,
          payload.transactions
        )
      }
    }),

    'transactions-next-page': (state, { payload }) => ({
      ...state,
      hasNextPage: payload.hasNextPage,
      page: payload.page
    }),

    'token-state-changed': (state, { payload }) => ({
      ...state,
      token: get(payload, 'wallet.token', state.token)
    }),

    'transactions-scan-started': state => ({
      ...state,
      syncStatus: 'syncing'
    }),

    'transactions-scan-finished': (state, { payload }) => ({
      ...state,
      syncStatus: payload.success ? 'up-to-date' : 'failed'
    }),

    'allow-send-transaction': (state, { payload }) => ({
      ...state,
      allowSendTransaction: payload.allowSendTransaction
    }),
    'set-marketplace-fee': (state, { payload }) => ({
      ...state,
      marketplaceFee: payload
    })
  },
  initialState
);

export default reducer;
