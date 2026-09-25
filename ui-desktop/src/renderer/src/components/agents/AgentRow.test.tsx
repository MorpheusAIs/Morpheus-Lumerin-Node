import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useIsOverflow } from '@renderer/hooks/useIsOverflow';
import { AgentRowComp } from './AgentRow';

vi.mock('@renderer/hooks/useIsOverflow', () => ({ useIsOverflow: vi.fn() }));
const overflow = vi.mocked(useIsOverflow);
const agent = {
  username: 'Research helper',
  perms: ['chat', 'wallet.read'],
  allowances: {
    MOR: '12.345',
    'another-token-with-a-long-name': '98765432109876543210',
  },
} as any;
const cfg = {
  symbol: 'MOR',
  symbolEth: 'ETH',
  morTokenAddress: '0x1111111111111111111111111111111111111111',
};
const row = () => (
  <AgentRowComp
    agent={agent}
    cfg={cfg}
    actions={<button type="button">Transactions</button>}
  />
);

describe('AgentRow allowances', () => {
  beforeEach(() => overflow.mockReturnValue({ x: false, y: false }));

  it('keeps every allowance mounted when an overflowing preview is detected', () => {
    overflow.mockReturnValue({ x: true, y: true });
    render(row());
    expect(screen.getByText('12.345')).toBeVisible();
    expect(screen.getByText('98765432109876543210')).toBeVisible();
    const button = screen.getByRole('button', { name: 'Show all allowances' });
    expect(button).toHaveAttribute('aria-expanded', 'false');
    expect(
      document.getElementById(button.getAttribute('aria-controls')!),
    ).toContainElement(screen.getByText('12.345'));
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('expands inline with keyboard and remains collapsible after the overflow clears', async () => {
    const user = userEvent.setup();
    overflow.mockReturnValue({ x: false, y: true });
    const { rerender } = render(row());
    screen.getByRole('button', { name: 'Show all allowances' }).focus();
    await user.keyboard('{Enter}');
    overflow.mockReturnValue({ x: false, y: false });
    rerender(row());
    expect(screen.getByRole('button', { name: 'Show less' })).toHaveAttribute(
      'aria-expanded',
      'true',
    );
    await user.click(screen.getByRole('button', { name: 'Show less' }));
    expect(
      screen.queryByRole('button', { name: 'Show less' }),
    ).not.toBeInTheDocument();
    expect(screen.getByText('12.345')).toBeVisible();
  });

  it('adapts disclosure availability after resizing without dropping values or actions', () => {
    const { rerender } = render(row());
    expect(
      screen.queryByRole('button', { name: 'Show all allowances' }),
    ).not.toBeInTheDocument();
    overflow.mockReturnValue({ x: false, y: true });
    rerender(row());
    expect(
      screen.getByRole('button', { name: 'Show all allowances' }),
    ).toBeVisible();
    overflow.mockReturnValue({ x: false, y: false });
    rerender(row());
    expect(
      screen.queryByRole('button', { name: 'Show all allowances' }),
    ).not.toBeInTheDocument();
    expect(screen.getByText('12.345')).toBeVisible();
    expect(screen.getByRole('button', { name: 'Transactions' })).toBeEnabled();
  });

  it('handles an agent without allowances and distinguishes empty permissions', () => {
    render(
      <AgentRowComp
        agent={{ ...agent, allowances: undefined, perms: [] }}
        cfg={cfg}
        actions={null}
      />,
    );
    expect(screen.getAllByText('None')).toHaveLength(2);
    expect(
      screen.queryByRole('button', { name: 'Show all allowances' }),
    ).not.toBeInTheDocument();
  });
});
