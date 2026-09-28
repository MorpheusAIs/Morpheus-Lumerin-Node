import { act, fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, useLocation } from 'react-router';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import QuickStartGuide, {
  openQuickStartGuide,
  QUICK_START_STORAGE_KEY,
} from './QuickStartGuide';

function LocationIndicator() {
  const location = useLocation();
  return <output data-testid="current-route">{location.pathname}</output>;
}

function renderGuide(initialEntry = '/wallet') {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <nav aria-label="Test navigation">
        {['wallet', 'chat', 'workspace', 'settings'].map((page) => (
          <a href={`#/${page}`} data-guide={page} key={page}>
            {page}
          </a>
        ))}
        <button data-guide-launcher onClick={openQuickStartGuide} type="button">
          Quick start
        </button>
      </nav>
      <button type="button">An app action</button>
      <LocationIndicator />
      <QuickStartGuide />
    </MemoryRouter>,
  );
}

describe('QuickStartGuide', () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => vi.restoreAllMocks());

  it('offers a non-modal invitation without forcing a tour or changing the route', () => {
    renderGuide('/workspace');

    expect(
      screen.getByRole('complementary', { name: 'Getting started' }),
    ).toBeVisible();
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(
      screen.queryByText('Your wallet, your address'),
    ).not.toBeInTheDocument();
    expect(screen.getByTestId('current-route')).toHaveTextContent('/workspace');
    expect(document.activeElement).toBe(document.body);
    expect(window.localStorage.getItem(QUICK_START_STORAGE_KEY)).toBeNull();
  });

  it('remembers dismissal across mounts but allows replay from the sidebar', async () => {
    const user = userEvent.setup();
    const view = renderGuide();
    await user.click(
      screen.getByRole('button', { name: 'Dismiss quick tour invitation' }),
    );
    expect(window.localStorage.getItem(QUICK_START_STORAGE_KEY)).toBe(
      'dismissed',
    );

    view.unmount();
    renderGuide();
    expect(screen.queryByRole('complementary')).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Quick start' }));
    expect(
      screen.getByRole('heading', { name: 'Your wallet, your address' }),
    ).toHaveFocus();
    expect(screen.getByLabelText('Step 1 of 4')).toBeInTheDocument();
  });

  it('only navigates when the user explicitly opens a page, and never claims to open a session', async () => {
    const user = userEvent.setup();
    renderGuide();
    await user.click(screen.getByRole('button', { name: 'Take a quick tour' }));
    expect(screen.getByText('You’re in Wallet')).toBeVisible();
    await user.click(screen.getByRole('button', { name: 'Next' }));

    expect(
      screen.getByRole('heading', { name: 'Choose your model & session' }),
    ).toHaveFocus();
    expect(screen.getByTestId('current-route')).toHaveTextContent('/wallet');
    expect(screen.getByText(/No subscriptions/)).toHaveTextContent(
      'unused stake returns',
    );
    await user.click(screen.getByRole('button', { name: 'Open Chat' }));
    expect(screen.getByTestId('current-route')).toHaveTextContent('/chat');
    expect(screen.getByText('You’re in Chat')).toBeVisible();
    expect(
      screen.getByRole('heading', { name: 'Choose your model & session' }),
    ).toHaveFocus();
  });

  it('explains project continuity and ends with a persistent completed preference', async () => {
    const user = userEvent.setup();
    renderGuide();
    await user.click(screen.getByRole('button', { name: 'Take a quick tour' }));
    await user.click(screen.getByRole('button', { name: 'Next' }));
    await user.click(screen.getByRole('button', { name: 'Next' }));

    expect(
      screen.getByText(/Your project history stays when a session ends/),
    ).toBeVisible();
    await user.click(screen.getByRole('button', { name: 'Open Workspace' }));
    expect(screen.getByTestId('current-route')).toHaveTextContent('/workspace');
    await user.click(screen.getByRole('button', { name: 'Next' }));
    expect(
      screen.getByRole('heading', { name: 'Know where to check' }),
    ).toBeVisible();
    expect(
      screen.getByText(/A connection problem is not a balance of zero/),
    ).toBeVisible();
    await user.click(screen.getByRole('button', { name: 'Done' }));

    expect(screen.queryByRole('complementary')).not.toBeInTheDocument();
    expect(window.localStorage.getItem(QUICK_START_STORAGE_KEY)).toBe(
      'completed',
    );
    expect(screen.getByRole('button', { name: 'Quick start' })).toHaveFocus();
  });

  it('highlights only the current navigation target and cleans up on skip', async () => {
    const user = userEvent.setup();
    renderGuide();
    await user.click(screen.getByRole('button', { name: 'Quick start' }));
    const wallet = screen.getByRole('link', { name: 'wallet' });
    const chat = screen.getByRole('link', { name: 'chat' });
    expect(wallet).toHaveAttribute('data-guide-active', 'true');
    await user.click(screen.getByRole('button', { name: 'Next' }));
    expect(wallet).not.toHaveAttribute('data-guide-active');
    expect(chat).toHaveAttribute('data-guide-active', 'true');
    await user.click(screen.getByRole('button', { name: 'Back' }));
    expect(wallet).toHaveAttribute('data-guide-active', 'true');
    expect(chat).not.toHaveAttribute('data-guide-active');
    await user.click(screen.getByRole('button', { name: 'Skip tour' }));
    expect(wallet).not.toHaveAttribute('data-guide-active');
  });

  it('minimizes without losing the step, then resumes with keyboard focus', async () => {
    const user = userEvent.setup();
    renderGuide();
    await user.click(screen.getByRole('button', { name: 'Take a quick tour' }));
    await user.click(screen.getByRole('button', { name: 'Next' }));
    await user.click(
      screen.getByRole('button', { name: 'Minimize quick start guide' }),
    );

    const resume = screen.getByRole('button', { name: /Resume tour/ });
    expect(resume).toHaveFocus();
    expect(resume).toHaveTextContent('2 / 4');
    expect(
      screen.queryByRole('heading', { name: 'Choose your model & session' }),
    ).not.toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'chat' })).not.toHaveAttribute(
      'data-guide-active',
    );
    await user.click(resume);
    expect(
      screen.getByRole('heading', { name: 'Choose your model & session' }),
    ).toHaveFocus();
  });

  it('closes on Escape inside the guide and restores focus to its launcher', async () => {
    const user = userEvent.setup();
    renderGuide();
    const launcher = screen.getByRole('button', { name: 'Quick start' });
    await user.click(launcher);
    await user.keyboard('{Escape}');

    expect(screen.queryByRole('complementary')).not.toBeInTheDocument();
    expect(launcher).toHaveFocus();
  });

  it('does not intercept Escape or trap focus while the user works elsewhere', async () => {
    const user = userEvent.setup();
    renderGuide();
    await user.click(screen.getByRole('button', { name: 'Quick start' }));
    await user.click(screen.getByRole('button', { name: 'An app action' }));
    await user.keyboard('{Escape}');

    expect(
      screen.getByRole('complementary', { name: 'Quick start guide' }),
    ).toBeVisible();
    expect(screen.getByRole('button', { name: 'An app action' })).toHaveFocus();
  });

  it('survives unavailable preference storage and missing navigation targets', async () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('Storage unavailable');
    });
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('Storage unavailable');
    });
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <QuickStartGuide />
      </MemoryRouter>,
    );
    await user.click(screen.getByRole('button', { name: 'Take a quick tour' }));
    expect(
      screen.getByRole('heading', { name: 'Your wallet, your address' }),
    ).toBeVisible();
    await user.click(screen.getByRole('button', { name: 'Skip tour' }));
    expect(screen.queryByRole('complementary')).not.toBeInTheDocument();
  });

  it('resets to the first step on replay and cleans up its event listener on unmount', async () => {
    const user = userEvent.setup();
    const view = renderGuide();
    await user.click(screen.getByRole('button', { name: 'Quick start' }));
    await user.click(screen.getByRole('button', { name: 'Next' }));
    act(() => openQuickStartGuide());
    const guide = screen.getByRole('complementary', {
      name: 'Quick start guide',
    });
    expect(
      within(guide).getByRole('heading', { name: 'Your wallet, your address' }),
    ).toHaveFocus();
    view.unmount();
    expect(() =>
      fireEvent(window, new Event('morpheus:open-quick-start')),
    ).not.toThrow();
  });
});
