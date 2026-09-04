import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { FC, PropsWithChildren } from 'react';
import { ThemeProvider as StyledThemeProvider } from 'styled-components';
import { describe, expect, it, vi } from 'vitest';
import theme from '../../ui/theme';
import { Models } from './Models';

vi.mock('../common/LayoutHeader', () => ({
  LayoutHeader: ({ children, title }) => (
    <header>
      {title}
      {children}
    </header>
  ),
}));
vi.mock('../common/View', () => ({
  View: ({ children }) => <main>{children}</main>,
}));
vi.mock('../common/QueryError', () => ({ default: () => null }));
vi.mock('../dashboard/BalanceBlock.styles', () => ({
  BtnAccent: ({ children, ...props }) => <button {...props}>{children}</button>,
}));
vi.mock('./ModelsTable', () => ({
  default: ({ models, isLoading }) => (
    <div data-testid="registry-state">
      {isLoading ? 'loading' : models.length}
    </div>
  ),
}));
vi.mock('./PinnedFilesTable', () => ({
  default: ({ pinnedFiles }) => (
    <div data-testid="pinned-state">{pinnedFiles.length}</div>
  ),
}));
vi.mock('./FileSelectionModal', () => ({ default: () => null }));

const ThemeProvider = StyledThemeProvider as unknown as FC<
  PropsWithChildren<{ theme: typeof theme }>
>;

describe('Models progressive loading', () => {
  it('does not enumerate IPFS pins until the Pinned Models tab is selected', async () => {
    const getPinnedFiles = vi.fn().mockResolvedValue([{ fileCIDHash: 'cid' }]);
    const getModelsPage = vi.fn().mockResolvedValue([]);
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false, staleTime: 30_000 },
      },
    });

    render(
      <QueryClientProvider client={queryClient}>
        <ThemeProvider theme={theme}>
          <Models
            setSelectedModel={vi.fn()}
            getIpfsVersion={vi.fn().mockResolvedValue({ version: '1.0' })}
            getModelsPage={getModelsPage}
            openSelectDownloadFolder={vi.fn()}
            addFileToIpfs={vi.fn()}
            getPinnedFiles={getPinnedFiles}
            pinFile={vi.fn()}
            unpinFile={vi.fn()}
            toasts={{ toast: vi.fn() }}
            client={{}}
            config={{}}
          />
        </ThemeProvider>
      </QueryClientProvider>,
    );

    expect(getPinnedFiles).not.toHaveBeenCalled();
    await waitFor(() => expect(getModelsPage).toHaveBeenCalledOnce());

    fireEvent.click(screen.getByRole('tab', { name: 'Pinned Models' }));
    await waitFor(() => expect(getPinnedFiles).toHaveBeenCalledOnce());
    expect((await screen.findByTestId('pinned-state')).textContent).toBe('1');

    fireEvent.click(screen.getByRole('tab', { name: 'Registry' }));
    fireEvent.click(screen.getByRole('tab', { name: 'Pinned Models' }));
    expect(getPinnedFiles).toHaveBeenCalledOnce();
  });
});
