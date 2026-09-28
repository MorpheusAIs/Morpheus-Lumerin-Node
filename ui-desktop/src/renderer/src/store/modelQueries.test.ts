import { describe, expect, it, vi } from 'vitest';
import { MODEL_PAGE_SIZE, modelPagesQueryOptions } from './modelQueries';

describe('model page query', () => {
  it('requests only the first page and advances by the page size', async () => {
    const getModelsPage = vi.fn().mockResolvedValue([{ Id: 'model-1' }]);
    const options = modelPagesQueryOptions(getModelsPage);

    await options.queryFn({ pageParam: 0 });

    expect(getModelsPage).toHaveBeenCalledWith({
      offset: 0,
      limit: MODEL_PAGE_SIZE,
    });
    expect(
      options.getNextPageParam(
        Array.from({ length: MODEL_PAGE_SIZE }, (_, Id) => ({ Id })),
        [Array.from({ length: MODEL_PAGE_SIZE })],
      ),
    ).toBe(MODEL_PAGE_SIZE);
    expect(
      options.getNextPageParam([{ Id: 1 }], [[{ Id: 1 }]]),
    ).toBeUndefined();
  });

  it('uses the raw page size for pagination even when a row is malformed', () => {
    const options = modelPagesQueryOptions(vi.fn());
    const rawPage = Array.from({ length: MODEL_PAGE_SIZE }, (_, Id) =>
      Id === 50 ? null : { Id },
    );

    expect(options.getNextPageParam(rawPage, [rawPage])).toBe(MODEL_PAGE_SIZE);
  });
});
