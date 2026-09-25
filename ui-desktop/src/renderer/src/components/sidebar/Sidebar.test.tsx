import { describe, expect, it, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { FC, PropsWithChildren } from 'react';
import { MemoryRouter } from 'react-router-dom';
import { ThemeProvider as StyledThemeProvider } from 'styled-components';
import theme from '../../ui/theme';
import { Sidebar } from './Sidebar';
import { SIDEBAR_COLLAPSED_KEY } from '../../hooks/usePersistedFlag';

// The nav lists reach into the redux/client contexts for routing and help
// links. None of that bears on whether the rail collapses, so they are stubbed
// to keep this a test of the layout preference rather than of the whole tree.
vi.mock('./PrimaryNav', () => ({ default: () => <nav data-testid="primary" /> }));
vi.mock('./SecondaryNav', () => ({
  default: () => <nav data-testid="secondary" />,
}));

const ThemeProvider = StyledThemeProvider as unknown as FC<
  PropsWithChildren<{ theme: typeof theme }>
>;

const renderSidebar = () =>
  render(
    <MemoryRouter>
      <ThemeProvider theme={theme}>
        <Sidebar
          address="0x1234567890123456789012345678901234567890"
          copyToClipboard={vi.fn()}
          onRouteIntent={vi.fn()}
        />
      </ThemeProvider>
    </MemoryRouter>,
  );

const aside = () => screen.getByLabelText('Morpheus navigation');

describe('Sidebar collapse', () => {
  beforeEach(() => window.localStorage.clear());

  it('starts expanded when no preference has been saved', () => {
    renderSidebar();

    expect(aside().getAttribute('data-sidebar-collapsed')).toBe('false');
    expect(screen.getByRole('button', { name: 'Collapse sidebar' })).toBeTruthy();
  });

  it('collapses on click and remembers the choice', async () => {
    renderSidebar();

    await userEvent.click(
      screen.getByRole('button', { name: 'Collapse sidebar' }),
    );

    expect(aside().getAttribute('data-sidebar-collapsed')).toBe('true');
    expect(window.localStorage.getItem(SIDEBAR_COLLAPSED_KEY)).toBe('true');
    // The one control still on screen has to say how to undo this.
    expect(screen.getByRole('button', { name: 'Expand sidebar' })).toBeTruthy();
  });

  it('restores the collapsed state on a later launch', () => {
    window.localStorage.setItem(SIDEBAR_COLLAPSED_KEY, 'true');

    renderSidebar();

    expect(aside().getAttribute('data-sidebar-collapsed')).toBe('true');
  });

  it('expands again from the collapsed state', async () => {
    window.localStorage.setItem(SIDEBAR_COLLAPSED_KEY, 'true');
    renderSidebar();

    await userEvent.click(
      screen.getByRole('button', { name: 'Expand sidebar' }),
    );

    expect(aside().getAttribute('data-sidebar-collapsed')).toBe('false');
    expect(window.localStorage.getItem(SIDEBAR_COLLAPSED_KEY)).toBe('false');
  });

  // The narrow-window overlay is a separate, deliberately transient toggle;
  // collapsing must not quietly repurpose it.
  it('leaves the mobile overlay flag alone', async () => {
    renderSidebar();

    await userEvent.click(
      screen.getByRole('button', { name: 'Collapse sidebar' }),
    );

    expect(aside().getAttribute('data-sidebar-expanded')).toBe('false');
  });
});
