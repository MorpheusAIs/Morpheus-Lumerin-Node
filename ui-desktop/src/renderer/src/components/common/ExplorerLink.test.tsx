import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { FC, PropsWithChildren } from 'react';
import { ThemeProvider as StyledThemeProvider } from 'styled-components';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import theme from '../../ui/theme';
import ExplorerLink, { explorerHost } from './ExplorerLink';

const ThemeProvider = StyledThemeProvider as unknown as FC<
  PropsWithChildren<{ theme: typeof theme }>
>;

const openLink = () => window.openLink as unknown as ReturnType<typeof vi.fn>;

const renderLink = (props: React.ComponentProps<typeof ExplorerLink>) =>
  render(
    <ThemeProvider theme={theme}>
      <ExplorerLink {...props} />
    </ThemeProvider>,
  );

describe('explorerHost', () => {
  it('names the destination so the user knows where they are going', () => {
    expect(explorerHost('https://basescan.org/tx/0x1')).toBe('basescan.org');
    expect(explorerHost('https://www.basescan.org/tx/0x1')).toBe(
      'basescan.org',
    );
  });

  it('stays neutral rather than throwing on config we did not write', () => {
    expect(explorerHost(undefined)).toBe('the block explorer');
    expect(explorerHost('')).toBe('the block explorer');
    expect(explorerHost('not a url')).toBe('the block explorer');
  });
});

describe('ExplorerLink', () => {
  beforeEach(() => openLink().mockClear());

  it('opens through the main process rather than a renderer window', async () => {
    renderLink({ url: 'https://basescan.org/tx/0x1' });
    await userEvent.click(
      screen.getByRole('button', { name: 'View transaction on basescan.org' }),
    );
    expect(openLink()).toHaveBeenCalledWith('https://basescan.org/tx/0x1');
  });

  it('names the kind of thing at the other end', () => {
    renderLink({ url: 'https://basescan.org/address/0x1', kind: 'account' });
    expect(
      screen.getByRole('button', { name: 'View account on basescan.org' }),
    ).toBeInTheDocument();
  });

  it('keeps the caller text as the accessible name, host in the tooltip', () => {
    renderLink({ url: 'https://basescan.org/tx/0xabc', children: '0xabc' });
    const link = screen.getByRole('button', { name: '0xabc' });
    expect(link).toHaveAttribute('title', 'View transaction on basescan.org');
  });

  it('goes inert instead of opening nothing when there is no URL', async () => {
    renderLink({ url: null });
    const link = screen.getByRole('button');
    expect(link).toBeDisabled();
    await userEvent.click(link);
    expect(openLink()).not.toHaveBeenCalled();
  });

  it('marks a whole clickable row as leaving the app', () => {
    renderLink({ url: 'https://basescan.org/tx/0x1', variant: 'row' });
    // The row's own content is layout, so the label has to carry the meaning.
    expect(
      screen.getByRole('button', { name: 'View transaction on basescan.org' }),
    ).toBeInTheDocument();
  });
});
