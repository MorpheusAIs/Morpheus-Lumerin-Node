// The router's model price index, as the picker consumes it.
//
// The picker deliberately carries no bids: loading them per model meant hundreds
// of requests and a list that stayed disabled until the last one landed, which
// is also why it could never be ordered by cost. The router now sweeps every
// model's live bids once behind its own cache and answers in a single call, and
// this module is the renderer's half of that: parsing the answer, comparing two
// models by price, and turning wei into something a row can print.

/** One model's live price range, exactly as the router sends it. */
export interface ModelPriceEntry {
  readonly model_id: string;
  /**
   * Decimal wei strings, per second of compute. Both are empty when the model
   * has no live bid. Empty is not zero: a model nobody serves has no price, and
   * coercing it to zero would sort it to the top of a cheapest-first list.
   */
  readonly min_price_per_second_wei: string;
  readonly max_price_per_second_wei: string;
  /** How many live bids the range was taken from. */
  readonly bid_count: number;
}

export interface ModelPricesResponse {
  readonly prices?: unknown;
  /** Models whose bids the router could not read on this sweep. */
  readonly failed_model_ids?: unknown;
}

export interface ModelPriceIndex {
  /** Price by model id. A missing id means "not priced", not "free". */
  readonly byModelId: ReadonlyMap<string, ModelPriceEntry>;
  /**
   * Ids the sweep failed on. Held apart from plain absence so a row can say
   * "could not be read" rather than implying the model has no providers.
   */
  readonly failedModelIds: ReadonlySet<string>;
}

export const EMPTY_MODEL_PRICE_INDEX: ModelPriceIndex = {
  byModelId: new Map(),
  failedModelIds: new Set(),
};

const isWeiString = (value: unknown): value is string =>
  typeof value === 'string' && /^\d+$/.test(value.trim());

/**
 * Builds the lookup the picker reads.
 *
 * Every field is re-checked rather than trusted. This crosses an IPC boundary
 * from a separate process, and a row that prints `undefined MOR/s` or a sort
 * that throws on a malformed entry are both worse than dropping the entry and
 * showing the model as unpriced.
 */
export const buildModelPriceIndex = (
  response: ModelPricesResponse | null | undefined,
): ModelPriceIndex => {
  const byModelId = new Map<string, ModelPriceEntry>();
  const prices = Array.isArray(response?.prices) ? response.prices : [];
  for (const raw of prices) {
    const entry = raw as Partial<ModelPriceEntry> | null;
    const modelId =
      typeof entry?.model_id === 'string' ? entry.model_id.trim() : '';
    if (!modelId) continue;
    const min = isWeiString(entry?.min_price_per_second_wei)
      ? (entry!.min_price_per_second_wei as string).trim()
      : '';
    const max = isWeiString(entry?.max_price_per_second_wei)
      ? (entry!.max_price_per_second_wei as string).trim()
      : '';
    const bidCount = Number(entry?.bid_count);
    byModelId.set(modelId, {
      model_id: modelId,
      min_price_per_second_wei: min,
      // A range needs both ends. Keeping a max without a min would let a row
      // print "– 0.19" with nothing on the left.
      max_price_per_second_wei: min ? max || min : '',
      bid_count: Number.isSafeInteger(bidCount) && bidCount > 0 ? bidCount : 0,
    });
  }

  const failed = Array.isArray(response?.failed_model_ids)
    ? response.failed_model_ids
    : [];
  const failedModelIds = new Set<string>();
  for (const id of failed) {
    if (typeof id === 'string' && id.trim()) failedModelIds.add(id.trim());
  }

  return { byModelId, failedModelIds };
};

/**
 * The cheapest live bid in wei, or undefined when the model has no price.
 *
 * BigInt rather than Number: these are 18-decimal wei values and the comparison
 * decides the order of the list, so it must not go through a float.
 */
export const minPriceWei = (
  entry: ModelPriceEntry | undefined,
): bigint | undefined => {
  if (!entry?.min_price_per_second_wei) return undefined;
  try {
    return BigInt(entry.min_price_per_second_wei);
  } catch {
    return undefined;
  }
};

/**
 * A wei-per-second string as MOR per second, for display only.
 *
 * A double is fine here and nowhere else: per-second prices are a fraction of a
 * MOR, the figure is rounded for the row anyway, and nothing is authorised off
 * it. Anything that decides a transaction amount stays in BigInt.
 */
export const weiToMorPerSecond = (value: string): number | undefined => {
  if (!isWeiString(value)) return undefined;
  const parsed = Number(value) / 1e18;
  return Number.isFinite(parsed) ? parsed : undefined;
};

export type PriceSort = 'recommended' | 'cheapest' | 'expensive';

export const PRICE_SORTS: { id: PriceSort; label: string; hint: string }[] = [
  {
    id: 'recommended',
    label: 'Recommended',
    hint: 'Online models first, then local, then by name.',
  },
  {
    id: 'cheapest',
    label: 'Cheapest first',
    hint: 'Lowest price per second of compute first. Models with no live provider are listed last.',
  },
  {
    id: 'expensive',
    label: 'Most expensive first',
    hint: 'Highest price per second of compute first. Models with no live provider are listed last.',
  },
];

/**
 * Orders two models by their cheapest live bid.
 *
 * Models with no price sort last in *both* directions. They are not free and
 * they are not expensive: there is nothing to pay because there is nobody to
 * pay, and floating them to either end of a price-ordered list would be a lie
 * in one direction and noise in the other.
 *
 * Returns 0 when neither has a price, leaving the caller's own tie-break to
 * decide, so the order stays stable rather than shuffling on every refetch.
 */
export const comparePriceAscending = (
  left: bigint | undefined,
  right: bigint | undefined,
): number => {
  if (left === undefined && right === undefined) return 0;
  if (left === undefined) return 1;
  if (right === undefined) return -1;
  if (left === right) return 0;
  return left < right ? -1 : 1;
};
