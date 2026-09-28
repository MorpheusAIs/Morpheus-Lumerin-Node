import { describe, expect, it } from 'vitest';
import {
  buildModelPriceIndex,
  comparePriceAscending,
  minPriceWei,
  weiToMorPerSecond,
} from './modelPrices';

describe('buildModelPriceIndex', () => {
  it('survives every shape the router is not supposed to send', () => {
    const index = buildModelPriceIndex({
      prices: [
        null,
        'not an object',
        { model_id: '   ' },
        { model_id: 'no-price' },
        {
          model_id: 'junk-price',
          min_price_per_second_wei: '1.5e18',
          max_price_per_second_wei: 'abc',
          bid_count: 2,
        },
        {
          model_id: 'negative-count',
          min_price_per_second_wei: '100',
          max_price_per_second_wei: '200',
          bid_count: -4,
        },
      ],
      failed_model_ids: ['failed-1', '', null, 42, '  failed-2  '],
    } as any);

    // An entry with no usable id is not an entry at all.
    expect(index.byModelId.has('')).toBe(false);
    expect(index.byModelId.has('   ')).toBe(false);

    // A model the sweep answered for but found no bids on keeps its row, with
    // empty prices. That is what lets a row say "no providers" rather than
    // "not checked yet".
    expect(index.byModelId.get('no-price')).toEqual({
      model_id: 'no-price',
      min_price_per_second_wei: '',
      max_price_per_second_wei: '',
      bid_count: 0,
    });

    // Non-integer wei is not wei.
    expect(index.byModelId.get('junk-price')?.min_price_per_second_wei).toBe('');
    expect(index.byModelId.get('junk-price')?.max_price_per_second_wei).toBe('');

    // A max without a valid min would render as "– 0.19".
    expect(index.byModelId.get('negative-count')?.bid_count).toBe(0);

    expect([...index.failedModelIds].sort()).toEqual(['failed-1', 'failed-2']);
  });

  it('treats a missing or malformed response as no prices at all', () => {
    for (const response of [
      undefined,
      null,
      {},
      { prices: 'nope', failed_model_ids: 7 },
    ]) {
      const index = buildModelPriceIndex(response as any);
      expect(index.byModelId.size).toBe(0);
      expect(index.failedModelIds.size).toBe(0);
    }
  });

  it('fills a missing max from the min so a single-bid model still renders', () => {
    const index = buildModelPriceIndex({
      prices: [
        {
          model_id: 'one-bid',
          min_price_per_second_wei: '1000',
          max_price_per_second_wei: '',
          bid_count: 1,
        },
      ],
    } as any);

    expect(index.byModelId.get('one-bid')).toEqual({
      model_id: 'one-bid',
      min_price_per_second_wei: '1000',
      max_price_per_second_wei: '1000',
      bid_count: 1,
    });
  });
});

describe('minPriceWei', () => {
  it('reads full-precision wei rather than a float', () => {
    // Beyond Number.MAX_SAFE_INTEGER on purpose: a double would round this and
    // two adjacent prices would compare equal.
    const wei = minPriceWei({
      model_id: 'big',
      min_price_per_second_wei: '9007199254740993',
      max_price_per_second_wei: '9007199254740993',
      bid_count: 1,
    });
    expect(wei).toBe(9007199254740993n);
  });

  it('is undefined when the model has no price', () => {
    expect(minPriceWei(undefined)).toBeUndefined();
    expect(
      minPriceWei({
        model_id: 'none',
        min_price_per_second_wei: '',
        max_price_per_second_wei: '',
        bid_count: 0,
      }),
    ).toBeUndefined();
  });
});

describe('comparePriceAscending', () => {
  it('orders priced models by price', () => {
    expect(comparePriceAscending(1n, 2n)).toBe(-1);
    expect(comparePriceAscending(2n, 1n)).toBe(1);
    expect(comparePriceAscending(2n, 2n)).toBe(0);
  });

  it('sinks unpriced models in both directions', () => {
    // This is the property the sort depends on: the caller multiplies the
    // priced comparison by -1 for "most expensive first", but must NOT
    // multiply this one, or unpriced models would lead that list.
    expect(comparePriceAscending(undefined, 5n)).toBe(1);
    expect(comparePriceAscending(5n, undefined)).toBe(-1);
  });

  it('leaves two unpriced models to the caller tie-break', () => {
    expect(comparePriceAscending(undefined, undefined)).toBe(0);
  });
});

describe('weiToMorPerSecond', () => {
  it('converts wei per second to MOR per second', () => {
    expect(weiToMorPerSecond('1000000000000000000')).toBe(1);
    expect(weiToMorPerSecond('100000000000000')).toBeCloseTo(0.0001, 12);
  });

  it('rejects anything that is not a decimal wei string', () => {
    expect(weiToMorPerSecond('')).toBeUndefined();
    expect(weiToMorPerSecond('1.5')).toBeUndefined();
    expect(weiToMorPerSecond('-1')).toBeUndefined();
    expect(weiToMorPerSecond('abc')).toBeUndefined();
    expect(weiToMorPerSecond(undefined as any)).toBeUndefined();
  });
});
