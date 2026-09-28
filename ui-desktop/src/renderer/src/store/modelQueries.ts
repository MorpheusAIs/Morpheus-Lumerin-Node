import { queryKeys } from './queries';

export const MODEL_PAGE_SIZE = 100;

export const modelPagesQueryOptions = (
  getModelsPage: (params: { offset: number; limit: number }) => Promise<any[]>,
) => ({
  queryKey: queryKeys.modelPages,
  queryFn: ({ pageParam }: { pageParam: number }) =>
    getModelsPage({ offset: pageParam, limit: MODEL_PAGE_SIZE }),
  initialPageParam: 0,
  getNextPageParam: (lastPage: any[], pages: any[][]) =>
    lastPage.length === MODEL_PAGE_SIZE
      ? pages.length * MODEL_PAGE_SIZE
      : undefined,
});
