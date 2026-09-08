import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { FC, PropsWithChildren } from 'react';
import { ThemeProvider as StyledThemeProvider } from 'styled-components';
import { describe, expect, it, vi } from 'vitest';
import theme from '../../ui/theme';
import { Agents } from './Agents';

vi.mock('../common/Modal', () => ({
  default: ({ isOpen, children, title, onRequestClose }: any) =>
    isOpen ? (
      <div role="dialog" aria-label={title}>
        <button onClick={onRequestClose} type="button">
          Close dialog
        </button>
        {children}
      </div>
    ) : null,
}));
vi.mock('../common/LayoutHeader', () => ({
  LayoutHeader: ({ title }: any) => <h1>{title}</h1>,
}));
const ThemeProvider = StyledThemeProvider as unknown as FC<
  PropsWithChildren<{ theme: typeof theme }>
>;

const props = (txModal: any) =>
  ({
    pendingAgents: [],
    activeAgents: [],
    allowanceRequests: [],
    agentsLoading: false,
    agentsError: null,
    retryAgents: vi.fn(),
    txModal,
    setTxModal: vi.fn(),
    handleApproveAccess: vi.fn(),
    handleApproveAllowance: vi.fn(),
    handleDeleteAgent: vi.fn(),
    symbol: 'MOR',
    symbolEth: 'ETH',
    morTokenAddress: '',
    txUrlResolver: (tx: string) => `https://example.com/transaction/${tx}`,
  }) as any;

describe('Agents transaction dialog', () => {
  it('shows loading feedback while requesting history and keeps close available', async () => {
    const input = props({ state: 'loading', agentName: 'helper' });
    render(
      <ThemeProvider theme={theme}>
        <Agents {...input} />
      </ThemeProvider>,
    );
    expect(screen.getByRole('status')).toHaveTextContent(
      'Loading transactions',
    );
    await userEvent.click(screen.getByRole('button', { name: 'Close dialog' }));
    expect(input.setTxModal).toHaveBeenCalledWith({ state: 'pending' });
  });

  it('offers a retry for the same agent after a failed request', async () => {
    const input = props({
      state: 'error',
      agentName: 'helper',
      error: 'Check your local node connection and try again.',
    });
    render(
      <ThemeProvider theme={theme}>
        <Agents {...input} />
      </ThemeProvider>,
    );
    expect(screen.getByRole('alert')).toHaveTextContent(
      'Check your local node',
    );
    await userEvent.click(
      screen.getByRole('button', { name: 'Retry transactions' }),
    );
    expect(input.setTxModal).toHaveBeenCalledWith({
      state: 'loading',
      agentName: 'helper',
    });
  });

  it('distinguishes a successful empty history from a request in progress', () => {
    render(
      <ThemeProvider theme={theme}>
        <Agents
          {...props({ state: 'success', agentName: 'helper', data: [] })}
        />
      </ThemeProvider>,
    );
    expect(screen.getByRole('status')).toHaveTextContent(
      'No transactions recorded',
    );
    expect(screen.queryByText('Loading transactions…')).not.toBeInTheDocument();
  });

  it('shows transaction links safely in a successful history', () => {
    render(
      <ThemeProvider theme={theme}>
        <Agents
          {...props({ state: 'success', agentName: 'helper', data: ['0xabc'] })}
        />
      </ThemeProvider>,
    );
    const link = screen.getByRole('link', { name: '0xabc' });
    expect(link).toHaveAttribute(
      'href',
      'https://example.com/transaction/0xabc',
    );
    expect(link).toHaveAttribute('rel', 'noopener noreferrer');
    expect(link).toHaveStyle({ overflowWrap: 'anywhere' });
  });
});
