import { describe, expect, it } from 'vitest';
import {
  normalizeModelList,
  normalizeModelMetadata,
  normalizeModelName,
  normalizeModelTags,
} from './modelMetadata';

describe('normalizeModelTags', () => {
  it.each([null, undefined, { bad: 'shape' }, 42, true])(
    'returns an empty list for unsupported metadata %#',
    (rawTags) => {
      expect(normalizeModelTags(rawTags)).toEqual([]);
    },
  );

  it('supports comma-separated legacy strings', () => {
    expect(normalizeModelTags(' llm, tee, vision ')).toEqual([
      'llm',
      'tee',
      'vision',
    ]);
  });

  it('drops malformed entries while preserving safe primitive tags', () => {
    expect(
      normalizeModelTags([
        ' llm ',
        null,
        undefined,
        { nested: true },
        ['tee'],
        7,
        false,
        '',
      ]),
    ).toEqual(['llm', '7', 'false']);
  });
});

describe('normalizeModelMetadata', () => {
  it('stabilizes names and tags without discarding other model fields', () => {
    expect(
      normalizeModelMetadata({ Id: 'model-1', Name: 42, Tags: null }),
    ).toEqual({ Id: 'model-1', Name: '42', Tags: [] });
  });

  it.each([null, undefined, 'model', 42, [], true])(
    'rejects invalid model rows %#',
    (model) => {
      expect(normalizeModelMetadata(model)).toBeNull();
    },
  );
});

describe('normalizeModelName', () => {
  it.each([null, undefined, {}, [], '', '   '])(
    'uses a readable fallback for invalid metadata %#',
    (rawName) => {
      expect(normalizeModelName(rawName)).toBe('Unnamed model');
    },
  );

  it('preserves valid string and numeric names', () => {
    expect(normalizeModelName('Model one')).toBe('Model one');
    expect(normalizeModelName(42)).toBe('42');
  });
});

describe('normalizeModelList', () => {
  it('fails closed for a non-array response', () => {
    expect(normalizeModelList({ models: [] })).toEqual([]);
  });

  it('normalizes valid rows and removes malformed catalog entries', () => {
    expect(
      normalizeModelList([
        { Id: 'one', Name: 'One', Tags: null },
        null,
        'bad-row',
        ['also-bad'],
        { Id: 'two', Name: 'Two', Tags: 'llm, tee' },
      ]),
    ).toEqual([
      { Id: 'one', Name: 'One', Tags: [] },
      { Id: 'two', Name: 'Two', Tags: ['llm', 'tee'] },
    ]);
  });
});
