import { describe, expect, it } from 'vitest';
import {
  formatMor,
  messageCostMor,
  readChunkUsage,
  summarizeUsage,
  totalUsage,
} from './messageUsage';

describe('readChunkUsage', () => {
  it('prefers the router\u2019s own count over the provider\u2019s', () => {
    // usage_from_consumer is what the session is settled against. A provider
    // reporting something different does not change what was paid.
    const usage = readChunkUsage({
      usage: { completion_tokens: 999, prompt_tokens: 999 },
      usage_from_consumer: { completion_tokens: 20, prompt_tokens: 10 },
    });
    expect(usage).toEqual({
      completionTokens: 20,
      promptTokens: 10,
      totalTokens: 30,
    });
  });

  it('falls back to the receiving side count', () => {
    expect(
      readChunkUsage({
        usage_from_provider: { completion_tokens: 5, prompt_tokens: 2 },
      }),
    ).toEqual({ completionTokens: 5, promptTokens: 2, totalTokens: 7 });
  });

  it('accepts a plain provider usage block when nothing else is present', () => {
    expect(
      readChunkUsage({ usage: { completion_tokens: 4, prompt_tokens: 1 } }),
    ).toEqual({ completionTokens: 4, promptTokens: 1, totalTokens: 5 });
  });

  it('uses a stated total rather than recomputing it', () => {
    // Some providers bill reasoning tokens that appear in neither half.
    expect(
      readChunkUsage({
        usage_from_consumer: {
          completion_tokens: 20,
          prompt_tokens: 10,
          total_tokens: 45,
        },
      }),
    ).toEqual({
      completionTokens: 20,
      promptTokens: 10,
      totalTokens: 45,
    });
  });

  it('returns null for the ordinary chunks that carry no usage', () => {
    expect(readChunkUsage({ choices: [{ delta: { content: 'hi' } }] })).toBe(
      null,
    );
    expect(readChunkUsage(null)).toBe(null);
    expect(readChunkUsage(undefined)).toBe(null);
    expect(readChunkUsage('a string chunk')).toBe(null);
  });

  it('skips an all-zero block and keeps looking', () => {
    // A zeroed usage object on every chunk is a placeholder, not a turn that
    // genuinely cost nothing.
    expect(
      readChunkUsage({
        usage: { completion_tokens: 7, prompt_tokens: 3 },
        usage_from_consumer: { completion_tokens: 0, prompt_tokens: 0 },
      }),
    ).toEqual({ completionTokens: 7, promptTokens: 3, totalTokens: 10 });
  });

  it('returns null when every block is empty', () => {
    expect(
      readChunkUsage({
        usage_from_consumer: { completion_tokens: 0, prompt_tokens: 0 },
      }),
    ).toBe(null);
  });

  it('treats a nonsense count as zero rather than showing it', () => {
    expect(
      readChunkUsage({
        usage_from_consumer: { completion_tokens: 12, prompt_tokens: -5 },
      }),
    ).toEqual({ completionTokens: 12, promptTokens: 0, totalTokens: 12 });
  });
});

describe('messageCostMor', () => {
  const oneMorPerSecond = '1000000000000000000';

  it('bills whole seconds, rounded up', () => {
    // The session is charged in whole seconds, so a 1.2s answer costs two.
    expect(messageCostMor(1_200, oneMorPerSecond)).toBe(2);
    expect(messageCostMor(2_000, oneMorPerSecond)).toBe(2);
  });

  it('converts from wei', () => {
    expect(messageCostMor(1_000, '1000000000000000')).toBeCloseTo(0.001, 12);
  });

  it('gives no figure when the price is unknown', () => {
    expect(messageCostMor(5_000, undefined)).toBeUndefined();
    expect(messageCostMor(5_000, 'not a number')).toBeUndefined();
  });

  it('gives no figure for a free bid rather than claiming zero spend', () => {
    expect(messageCostMor(5_000, 0)).toBeUndefined();
  });

  it('gives no figure when the answer was never timed', () => {
    expect(messageCostMor(undefined, oneMorPerSecond)).toBeUndefined();
    expect(messageCostMor(-1, oneMorPerSecond)).toBeUndefined();
  });

  it('charges nothing for an answer that took no measurable time', () => {
    expect(messageCostMor(0, oneMorPerSecond)).toBe(0);
  });
});

describe('formatMor', () => {
  it('keeps significant figures rather than fixed decimals', () => {
    expect(formatMor(0.00012345)).toBe('0.0001234 MOR');
    expect(formatMor(1.5)).toBe('1.5 MOR');
  });

  it('says a charge is small instead of rounding it to nothing', () => {
    expect(formatMor(0.00000001)).toBe('<0.0001 MOR');
  });

  it('prints an exact zero as zero', () => {
    expect(formatMor(0)).toBe('0 MOR');
  });

  it('prints nothing for a value that is not a number', () => {
    expect(formatMor(Number.NaN)).toBe('');
  });
});

describe('summarizeUsage', () => {
  const usage = {
    completionTokens: 318,
    promptTokens: 1_204,
    totalTokens: 1_522,
  };

  it('reads as one line of plain facts', () => {
    expect(summarizeUsage(usage, 0.002)).toBe('1,204 in · 318 out · 0.002 MOR');
  });

  it('omits cost rather than claiming zero when the price is unknown', () => {
    expect(summarizeUsage(usage)).toBe('1,204 in · 318 out');
  });

  it('shows a cost on its own when no tokens were reported', () => {
    expect(summarizeUsage(undefined, 0.002)).toBe('0.002 MOR');
  });

  it('is empty when nothing at all was measured', () => {
    expect(summarizeUsage(undefined)).toBe('');
    expect(
      summarizeUsage({
        completionTokens: 0,
        promptTokens: 0,
        totalTokens: 0,
      }),
    ).toBe('');
  });
});

describe('totalUsage', () => {
  it('adds up the turns that reported something', () => {
    const result = totalUsage([
      {
        costMor: 0.001,
        usage: { completionTokens: 20, promptTokens: 10, totalTokens: 30 },
      },
      { costMor: 0.002 },
      {
        usage: { completionTokens: 5, promptTokens: 1, totalTokens: 6 },
      },
    ]);
    expect(result.usage).toEqual({
      completionTokens: 25,
      promptTokens: 11,
      totalTokens: 36,
    });
    expect(result.costMor).toBeCloseTo(0.003, 12);
  });

  it('is zero for an empty transcript', () => {
    expect(totalUsage([])).toEqual({
      costMor: 0,
      usage: { completionTokens: 0, promptTokens: 0, totalTokens: 0 },
    });
  });

  it('ignores a turn whose cost is not a number', () => {
    expect(totalUsage([{ costMor: Number.NaN }]).costMor).toBe(0);
  });
});
