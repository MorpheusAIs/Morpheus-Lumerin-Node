const MAX_MODEL_TAGS = 64;
const MAX_MODEL_TAG_LENGTH = 64;

/**
 * Converts imperfect registry metadata into the stable shape UI consumers use.
 * The chain schema says Tags is string[], but older or malformed responses can
 * contain null, a comma-separated string, or non-string array members. A bad
 * row must lose optional badges, not crash the entire Chat route.
 */
export const normalizeModelTags = (rawTags: unknown): string[] => {
  const candidates = Array.isArray(rawTags)
    ? rawTags
    : typeof rawTags === 'string'
      ? rawTags.split(',')
      : [];
  const tags: string[] = [];

  for (const candidate of candidates.slice(0, MAX_MODEL_TAGS)) {
    if (
      typeof candidate !== 'string' &&
      typeof candidate !== 'number' &&
      typeof candidate !== 'boolean'
    ) {
      continue;
    }
    const tag = String(candidate).trim().slice(0, MAX_MODEL_TAG_LENGTH);
    if (tag) tags.push(tag);
  }

  return tags;
};

export const normalizeModelName = (rawName: unknown): string => {
  if (
    typeof rawName !== 'string' &&
    typeof rawName !== 'number' &&
    typeof rawName !== 'boolean'
  ) {
    return 'Unnamed model';
  }
  const name = String(rawName);
  return name.trim() ? name : 'Unnamed model';
};

export const normalizeModelMetadata = (
  value: unknown,
): Record<string, any> | null => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null;
  const model = value as Record<string, any>;
  return {
    ...model,
    Name: normalizeModelName(model.Name),
    Tags: normalizeModelTags(model.Tags),
  };
};

export const normalizeModelList = (value: unknown): Record<string, any>[] => {
  if (!Array.isArray(value)) return [];
  return value
    .map(normalizeModelMetadata)
    .filter((model): model is Record<string, any> => model !== null);
};
