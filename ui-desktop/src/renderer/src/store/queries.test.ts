import { QueryClient } from '@tanstack/react-query';
import { describe, expect, it, vi } from 'vitest';
import { computeStakedFunds, countOpenSessions, queryKeys } from './queries';

const WEI = 10 ** 18;
const secondsFromNow = (s: number) => Math.floor(Date.now() / 1000) + s;

describe('computeStakedFunds', () => {
  it('returns 0 when sessions have not loaded', () => {
    expect(computeStakedFunds(undefined)).toBe('0');
  });

  // Note the asymmetry: `undefined` means "not loaded yet" and yields the bare
  // '0', whereas a loaded-but-empty list yields the formatted '0.00'.
  it('returns a formatted zero for an empty list', () => {
    expect(computeStakedFunds([])).toBe('0.00');
  });

  it('sums stake across open sessions', () => {
    expect(
      computeStakedFunds([
        { Stake: 1 * WEI, EndsAt: secondsFromNow(3600) },
        { Stake: 2.5 * WEI, EndsAt: secondsFromNow(7200) },
      ]),
    ).toBe('3.50');
  });

  it('excludes explicitly closed sessions', () => {
    expect(
      computeStakedFunds([
        { Stake: 1 * WEI, EndsAt: secondsFromNow(3600) },
        { Stake: 5 * WEI, EndsAt: secondsFromNow(3600), ClosedAt: 12345 },
      ]),
    ).toBe('1.00');
  });

  it('excludes sessions whose end time has passed', () => {
    expect(
      computeStakedFunds([
        { Stake: 1 * WEI, EndsAt: secondsFromNow(3600) },
        { Stake: 9 * WEI, EndsAt: secondsFromNow(-3600) },
      ]),
    ).toBe('1.00');
  });

  it('does not throw on malformed entries', () => {
    expect(() => computeStakedFunds([null as any])).not.toThrow();
  });
});

// Gates the "you have N open sessions" confirmation before a wallet switch.
// Under-counting means switching silently kills a live chat; over-counting
// nags the user on every switch.
describe('countOpenSessions', () => {
  it('counts only sessions that are still open', () => {
    expect(
      countOpenSessions([
        { Stake: 1, EndsAt: secondsFromNow(3600) },
        { Stake: 1, EndsAt: secondsFromNow(3600), ClosedAt: 999 },
        { Stake: 1, EndsAt: secondsFromNow(-10) },
        { Stake: 1, EndsAt: secondsFromNow(60) },
      ]),
    ).toBe(2);
  });

  it.each([
    ['undefined', undefined],
    ['empty', []],
    ['not an array', {} as any],
  ])('returns 0 for %s', (_label, input) => {
    expect(countOpenSessions(input as any)).toBe(0);
  });

  it('skips malformed entries instead of throwing', () => {
    expect(() =>
      countOpenSessions([null as any, undefined as any]),
    ).not.toThrow();
    expect(countOpenSessions([null as any])).toBe(0);
  });
});

describe('queryKeys', () => {
  // Chat and Wallet must resolve to the same cache entry for a given wallet,
  // otherwise opening a session on one tab leaves the other showing stale data
  // — the cause of the "staked but the app forgot" report.
  it('is stable for the same address', () => {
    expect(queryKeys.sessions('0xabc')).toEqual(queryKeys.sessions('0xabc'));
    expect(queryKeys.balances('0xabc')).toEqual(queryKeys.balances('0xabc'));
  });

  it('separates different addresses', () => {
    expect(queryKeys.sessions('0xabc')).not.toEqual(
      queryKeys.sessions('0xdef'),
    );
    expect(queryKeys.agents('0xabc')).not.toEqual(queryKeys.agents('0xdef'));
  });

  it('tolerates a missing address without collapsing to undefined', () => {
    expect(queryKeys.sessions(undefined)).toEqual(['sessions', '']);
    expect(queryKeys.balances(undefined)).toEqual(['balances', '']);
  });

  it('scopes chat funding and selected-model bids to the active wallet', () => {
    expect(queryKeys.chatFunding('0xabc')).toEqual(['chatFunding', '0xabc']);
    expect(queryKeys.chatFunding('0xabc')).not.toEqual(
      queryKeys.chatFunding('0xdef'),
    );
    expect(queryKeys.modelBids('0xabc', 'model-1')).toEqual([
      'modelBids',
      '0xabc',
      'model-1',
    ]);
    expect(queryKeys.modelBids('0xabc', 'model-1')).not.toEqual(
      queryKeys.modelBids('0xabc', 'model-2'),
    );
  });

  it('reuses a fresh selected-model bid result instead of refetching it', async () => {
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const loadBids = vi.fn().mockResolvedValue([{ Id: 'bid-1' }]);
    const options = {
      queryKey: queryKeys.modelBids('0xabc', 'model-1'),
      queryFn: loadBids,
      staleTime: 60_000,
    };

    await queryClient.fetchQuery(options);
    await queryClient.fetchQuery(options);

    expect(loadBids).toHaveBeenCalledOnce();
  });
});
