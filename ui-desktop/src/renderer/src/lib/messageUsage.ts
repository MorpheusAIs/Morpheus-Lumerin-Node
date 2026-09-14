/**
 * Token counts and cost for a single answer.
 *
 * The router attaches its own token counts to the final streamed chunk rather
 * than trusting the provider's, under the key `usage_from_consumer` on the
 * sending side and `usage_from_provider` on the receiving side. A provider's
 * own `usage` block may also be present and is read last, because it is the
 * only one of the three that is not what the session is settled against.
 *
 * Cost is not a token price. A Morpheus session is billed per second against
 * the bid it was opened at, so the charge for one answer is the time that
 * answer took, and that is what is shown. Tokens and cost are two separate
 * facts about the same turn rather than two views of one.
 */

export type MessageUsage = {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
};

/** Usage keys in the order the router considers them authoritative. */
const USAGE_KEYS = [
  'usage_from_consumer',
  'usage_from_provider',
  'usage',
] as const;

const toCount = (value: unknown): number => {
  const n = Number(value);
  // Negative or fractional token counts mean the field was not what we expected.
  // Showing nothing beats showing "-1 in".
  return Number.isFinite(n) && n >= 0 ? Math.round(n) : 0;
};

/**
 * The usage block on a streamed chunk, or null when the chunk does not carry
 * one. Most chunks do not: only the final one has counts, so this returns null
 * for the whole body of a stream and the caller keeps the last non-null answer.
 */
export function readChunkUsage(part: unknown): MessageUsage | null {
  if (!part || typeof part !== 'object') return null;
  const record = part as Record<string, unknown>;

  for (const key of USAGE_KEYS) {
    const block = record[key];
    if (!block || typeof block !== 'object') continue;
    const usage = block as Record<string, unknown>;
    const promptTokens = toCount(usage.prompt_tokens);
    const completionTokens = toCount(usage.completion_tokens);
    const totalTokens =
      toCount(usage.total_tokens) || promptTokens + completionTokens;
    // An all-zero block is a placeholder the provider sent on every chunk, not
    // a real measurement of a turn that used no tokens.
    if (totalTokens === 0) continue;
    return { completionTokens, promptTokens, totalTokens };
  }
  return null;
}

/**
 * MOR spent on the seconds this answer occupied, or undefined when the price is
 * unknown or the request is free. Seconds are rounded up because a session is
 * billed in whole seconds, so reporting 0.4s of a 1s charge would understate it.
 */
export function messageCostMor(
  elapsedMs: number | undefined,
  pricePerSecondWei: string | number | undefined,
): number | undefined {
  if (elapsedMs === undefined || pricePerSecondWei === undefined) {
    return undefined;
  }
  const price = Number(pricePerSecondWei);
  if (!Number.isFinite(price) || price <= 0) return undefined;
  if (!Number.isFinite(elapsedMs) || elapsedMs < 0) return undefined;
  const seconds = Math.ceil(elapsedMs / 1000);
  return (seconds * price) / 1e18;
}

/**
 * A MOR amount short enough to sit under a message. Very small charges are the
 * normal case, so this keeps four significant figures rather than a fixed
 * number of decimals, which would round almost every real answer to 0.
 */
export function formatMor(amount: number): string {
  if (!Number.isFinite(amount)) return '';
  if (amount === 0) return '0 MOR';
  if (amount < 0.0001) return '<0.0001 MOR';
  return `${Number(amount.toPrecision(4))} MOR`;
}

export function formatTokenCount(count: number): string {
  return count.toLocaleString('en-US');
}

/**
 * The one-line summary shown under an answer, or an empty string when there is
 * nothing measured to report. Cost is appended only when it is known, so a free
 * or locally-run model shows tokens alone rather than a misleading "0 MOR".
 */
export function summarizeUsage(
  usage: MessageUsage | undefined,
  costMor?: number,
): string {
  const parts: string[] = [];
  if (usage && usage.totalTokens > 0) {
    parts.push(`${formatTokenCount(usage.promptTokens)} in`);
    parts.push(`${formatTokenCount(usage.completionTokens)} out`);
  }
  if (costMor !== undefined && Number.isFinite(costMor)) {
    parts.push(formatMor(costMor));
  }
  return parts.join(' · ');
}

export type UsageTotals = { costMor: number; usage: MessageUsage };

/** Running totals for the session footer. Skips turns that reported nothing. */
export function totalUsage(
  messages: readonly { usage?: MessageUsage; costMor?: number }[],
): UsageTotals {
  return messages.reduce<UsageTotals>(
    (acc, message) => {
      if (message.usage) {
        acc.usage.promptTokens += message.usage.promptTokens;
        acc.usage.completionTokens += message.usage.completionTokens;
        acc.usage.totalTokens += message.usage.totalTokens;
      }
      if (message.costMor !== undefined && Number.isFinite(message.costMor)) {
        acc.costMor += message.costMor;
      }
      return acc;
    },
    {
      costMor: 0,
      usage: { completionTokens: 0, promptTokens: 0, totalTokens: 0 },
    },
  );
}
