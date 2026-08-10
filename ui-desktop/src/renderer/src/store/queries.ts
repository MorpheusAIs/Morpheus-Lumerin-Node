// Centralized react-query keys so that data shared across tabs (models,
// providers, sessions, balances, transactions) dedupes against a single cache
// entry instead of every tab refetching on mount.
//
// Keys are intentionally stable and parameterized only by values that change
// the result (e.g. the wallet address). The query *functions* live next to the
// components/HOCs that own the proxy-router fetch logic and are passed inline to
// `useQuery`, since they close over the IPC client + chain config.
export const queryKeys = {
  // { models, providers, meta, userBalances } from withChatState.getModelsData
  modelsData: ['modelsData'] as const,
  // models merged with their bids (withChatState.getBidsByModelId fan-out)
  modelsWithBids: ['modelsWithBids'] as const,
  // raw on-chain user sessions (paginated) — shared by Chat + Wallet
  sessions: (address?: string) => ['sessions', address ?? ''] as const,
  // saved chat-history titles (local proxy-router index)
  chatTitles: ['chatTitles'] as const,
  // provider connectivity ping results
  providersAvailability: ['providersAvailability'] as const,
  // wallet ETH/MOR balances (+ MOR rate)
  balances: (address?: string) => ['balances', address ?? ''] as const,
  // Blockscout transaction history
  transactions: (address?: string) => ['transactions', address ?? ''] as const,
  // raw on-chain model list (Models tab registry)
  allModels: ['allModels'] as const,
  // local IPFS node version / connectivity
  ipfsVersion: ['ipfsVersion'] as const,
  // IPFS-pinned model files
  pinnedFiles: ['pinnedFiles'] as const,
  // provider sessions + claimable balances (Providers tab)
  providerData: (providerId?: string) =>
    ['providerData', providerId ?? ''] as const,
  // registry of known wallets + which one the proxy-router currently holds.
  // Intentionally NOT parameterised by address: this is the list that tells you
  // what the address could be.
  wallets: ['wallets'] as const,
};

const isClosedSession = (item: any) =>
  item.ClosedAt || new Date().getTime() > item.EndsAt * 1000;

/** Number of sessions still open. Used to warn before switching wallets. */
export const countOpenSessions = (sessions: any[] | undefined): number => {
  if (!Array.isArray(sessions)) {
    return 0;
  }
  try {
    return sessions.filter((s) => s && !isClosedSession(s)).length;
  } catch (e) {
    return 0;
  }
};

// Sum of stake locked in currently-open sessions, in whole MOR (2 decimals).
// Mirrors the previous withDashboardState.getStakedFunds computation but works
// off the shared sessions cache instead of issuing its own paginated fetch.
export const computeStakedFunds = (sessions: any[] | undefined): string => {
  if (!sessions) {
    return '0';
  }
  try {
    const openSessions = sessions.filter((s) => !isClosedSession(s));
    const sum = openSessions.reduce((curr, next) => curr + next.Stake, 0);
    return (sum / 10 ** 18).toFixed(2);
  } catch (e) {
    console.log('Error', e);
    return '0';
  }
};
