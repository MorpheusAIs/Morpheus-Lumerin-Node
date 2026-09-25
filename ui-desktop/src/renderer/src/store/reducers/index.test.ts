import { describe, expect, it } from 'vitest';
import reducer from './index';

const ADDRESS = '0x1111111111111111111111111111111111111111';

const bootstrap = (wallet?: Record<string, unknown>) => ({
  type: 'initial-state-received',
  payload: {
    config: { chain: { chainId: 'base' } },
    ...(wallet ? { chain: { wallet } } : {}),
  },
});

describe('root reducer wallet hydration', () => {
  it('does not let a late empty bootstrap overwrite an open wallet', () => {
    const opened = reducer(undefined, {
      type: 'open-wallet',
      payload: { chain: 'base', address: ADDRESS, isActive: true },
    });

    const hydrated = reducer(
      opened,
      bootstrap({ address: '', isActive: false }),
    );

    expect(hydrated.chain.wallet.address).toBe(ADDRESS);
    expect(hydrated.chain.wallet.isActive).toBe(true);
    expect(hydrated.chain.wallet).toMatchObject({
      syncStatus: 'up-to-date',
      transactions: {},
      token: {
        lmrBalance: 0,
        transactions: {},
        symbol: 'LMR',
        symbolEth: 'ETH',
      },
    });
  });

  it('hydrates the nested persisted wallet before a runtime wallet is open', () => {
    const hydrated = reducer(
      undefined,
      bootstrap({
        address: ADDRESS,
        isActive: false,
        ethBalance: 42,
        token: { lmrBalance: 84, symbol: 'MOR', symbolEth: 'ETH' },
      }),
    );

    expect(hydrated.chain.wallet).toMatchObject({
      address: ADDRESS,
      isActive: false,
      ethBalance: 42,
      token: { lmrBalance: 84, symbol: 'MOR', symbolEth: 'ETH' },
    });
  });

  it('keeps cold defaults when no persisted wallet exists', () => {
    const hydrated = reducer(undefined, bootstrap());

    expect(hydrated.chain.wallet.address).toBe('');
    expect(hydrated.chain.wallet.isActive).toBe(false);
  });
});
