import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react';
import { Provider as ReduxProvider } from 'react-redux';
import { createStore } from 'redux';
import { ThemeProvider } from 'styled-components';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { Provider as ClientProvider } from '../../store/hocs/clientContext';
import theme from '../../ui/theme';
import Onboarding from './Onboarding';

vi.mock('../common/PasswordStrengthMeter', () => ({ default: () => null }));

// Public BIP-39 test vector, never a real user's recovery phrase.
const TEST_PHRASE =
  'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about';
const TEST_PASSWORD = 'onboarding-test-password-only';
const TEST_ADDRESS = '0x1111111111111111111111111111111111111111';

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<T>((yes, no) => {
    resolve = yes;
    reject = no;
  });
  return { promise, resolve, reject };
}

function renderOnboarding(
  options: {
    createMnemonic?: () => Promise<string>;
    onCompleted?: (payload: unknown) => Promise<unknown>;
  } = {},
) {
  const client = {
    createMnemonic: vi.fn(options.createMnemonic ?? (async () => TEST_PHRASE)),
    isValidMnemonic: vi.fn((phrase: string) => phrase === TEST_PHRASE),
    suggestAddresses: vi.fn(async () => [TEST_ADDRESS]),
    onTermsLinkClick: vi.fn(),
  };
  const onCompleted = vi.fn(options.onCompleted ?? (async () => undefined));
  const store = createStore(() => ({ config: {} }));
  // styled-components v4's provider typing predates React 18.
  const Theme = ThemeProvider as any;
  render(
    <ReduxProvider store={store}>
      <ClientProvider value={client}>
        <Theme theme={theme}>
          <Onboarding onOnboardingCompleted={onCompleted} />
        </Theme>
      </ClientProvider>
    </ReduxProvider>,
  );
  return { client, onCompleted };
}

function selectWallet(mode: 'create' | 'import') {
  fireEvent.click(
    screen.getByRole('button', {
      name:
        mode === 'create' ? 'Create a new wallet' : 'Import an existing wallet',
    }),
  );
}

function acceptTerms() {
  fireEvent.click(
    screen.getByRole('checkbox', {
      name: 'I have read and accept these terms',
    }),
  );
  fireEvent.click(
    screen.getByRole('checkbox', {
      name: 'I have read and accept the software license',
    }),
  );
  fireEvent.click(screen.getByRole('button', { name: 'Accept and continue' }));
}

function enterPassword() {
  fireEvent.change(screen.getByTestId('pass-field'), {
    target: { value: TEST_PASSWORD },
  });
  fireEvent.change(screen.getByTestId('pass-again-field'), {
    target: { value: TEST_PASSWORD },
  });
  fireEvent.submit(screen.getByTestId('pass-form'));
}

async function reachPhraseVerification() {
  selectWallet('create');
  acceptTerms();
  enterPassword();
  expect(await screen.findByTestId('mnemonic-label')).toHaveTextContent(
    TEST_PHRASE,
  );
  fireEvent.click(screen.getByTestId('copied-mnemonic-btn'));
  fireEvent.change(screen.getByTestId('mnemonic-field'), {
    target: { value: TEST_PHRASE },
  });
}

describe('first-run onboarding integration', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('starts with wallet options instead of asking for an existing password', () => {
    const { client, onCompleted } = renderOnboarding();

    expect(
      screen.getByRole('heading', { name: 'Welcome to Morpheus' }),
    ).toBeVisible();
    expect(
      screen.getByRole('button', { name: 'Create a new wallet' }),
    ).toBeEnabled();
    expect(
      screen.getByRole('button', { name: 'Import an existing wallet' }),
    ).toBeEnabled();
    expect(screen.queryByTestId('pass-field')).not.toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'Login' }),
    ).not.toBeInTheDocument();
    expect(client.createMnemonic).not.toHaveBeenCalled();
    expect(onCompleted).not.toHaveBeenCalled();
  });

  it.each(['create', 'import'] as const)(
    'requires both visible terms consents before a new password for %s',
    (mode) => {
      const blockedFetch = vi.fn(() => {
        throw new Error('Offline');
      });
      vi.stubGlobal('fetch', blockedFetch);
      const { client, onCompleted } = renderOnboarding();
      selectWallet(mode);

      expect(
        screen.getByRole('heading', { name: 'Morpheus Terms of Use' }),
      ).toBeVisible();
      expect(
        screen.getByText(/THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY/),
      ).toBeVisible();
      const accept = screen.getByRole('button', {
        name: 'Accept and continue',
      });
      expect(accept).toBeDisabled();
      expect(screen.queryByTestId('pass-field')).not.toBeInTheDocument();
      fireEvent.click(
        screen.getByRole('checkbox', {
          name: 'I have read and accept these terms',
        }),
      );
      expect(accept).toBeDisabled();
      fireEvent.click(
        screen.getByRole('checkbox', {
          name: 'I have read and accept the software license',
        }),
      );
      expect(accept).toBeEnabled();
      fireEvent.click(accept);

      expect(
        screen.getByRole('heading', { name: 'Create your app password' }),
      ).toBeVisible();
      expect(screen.getByLabelText('New password')).toHaveAttribute(
        'autocomplete',
        'new-password',
      );
      expect(client.createMnemonic).not.toHaveBeenCalled();
      expect(onCompleted).not.toHaveBeenCalled();
      expect(blockedFetch).not.toHaveBeenCalled();

      fireEvent.submit(screen.getByTestId('pass-form'));
      expect(screen.getByText('Password is required')).toBeVisible();
      expect(client.createMnemonic).not.toHaveBeenCalled();
    },
  );

  it('expands the bundled software license without opening a browser or fetching content', () => {
    const blockedFetch = vi.fn(() => {
      throw new Error('Offline');
    });
    vi.stubGlobal('fetch', blockedFetch);
    const openWindow = vi.spyOn(window, 'open').mockImplementation(() => null);
    const { client } = renderOnboarding();
    selectWallet('create');

    const readLicense = screen.getByRole('button', {
      name: 'Read the software license',
    });
    expect(readLicense).toHaveAttribute('aria-expanded', 'false');
    expect(
      screen.queryByRole('region', { name: 'Morpheus software license' }),
    ).not.toBeInTheDocument();
    fireEvent.click(readLicense);

    const license = screen.getByRole('region', {
      name: 'Morpheus software license',
    });
    expect(license).toBeVisible();
    expect(license).toHaveTextContent('MIT License');
    expect(license).toHaveTextContent('Copyright (c) 2025 Morpheus');
    expect(license).toHaveTextContent(
      'THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY',
    );
    const hideLicense = screen.getByRole('button', {
      name: 'Hide the software license',
    });
    expect(hideLicense).toHaveAttribute('aria-expanded', 'true');
    expect(hideLicense).toHaveAttribute('aria-controls', license.id);
    fireEvent.click(hideLicense);

    expect(
      screen.queryByRole('region', { name: 'Morpheus software license' }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: 'Read the software license' }),
    ).toHaveAttribute('aria-expanded', 'false');
    expect(client.onTermsLinkClick).not.toHaveBeenCalled();
    expect(openWindow).not.toHaveBeenCalled();
    expect(window.openLink).not.toHaveBeenCalled();
    expect(blockedFetch).not.toHaveBeenCalled();
  });

  it('keeps the creation phrase and verification page after a backend failure, prevents duplicates and allows retry', async () => {
    const pending = deferred<void>();
    const completed = vi
      .fn()
      .mockImplementationOnce(() => pending.promise)
      .mockResolvedValue(undefined);
    const { client, onCompleted } = renderOnboarding({
      onCompleted: completed,
    });
    await reachPhraseVerification();
    expect(client.createMnemonic).toHaveBeenCalledTimes(1);

    fireEvent.submit(screen.getByTestId('mnemonic-form'));
    fireEvent.submit(screen.getByTestId('mnemonic-form'));
    expect(onCompleted).toHaveBeenCalledTimes(1);
    expect(screen.getByTestId('mnemonic-field')).toBeDisabled();
    expect(screen.getByRole('status')).toHaveTextContent(
      'Setting up your wallet',
    );
    await act(async () =>
      pending.reject(new Error('Local node is restarting. Try again.')),
    );

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Local node is restarting. Try again.',
    );
    expect(
      screen.getByRole('heading', { name: 'Recovery Passphrase' }),
    ).toBeVisible();
    expect(screen.getByTestId('mnemonic-field')).toHaveValue(TEST_PHRASE);
    expect(screen.getByRole('button', { name: 'Create wallet' })).toBeEnabled();
    fireEvent.submit(screen.getByTestId('mnemonic-form'));
    await waitFor(() => expect(onCompleted).toHaveBeenCalledTimes(2));
    expect(onCompleted).toHaveBeenLastCalledWith({
      password: TEST_PASSWORD,
      mnemonic: TEST_PHRASE,
      privateKey: '',
      derivationPath: '0',
      ethNode: '',
    });
    expect(client.createMnemonic).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('imports the supplied phrase without generating another one and preserves connection input on failed setup', async () => {
    const pending = deferred<void>();
    const completed = vi
      .fn()
      .mockImplementationOnce(() => pending.promise)
      .mockResolvedValue(undefined);
    const { client, onCompleted } = renderOnboarding({
      onCompleted: completed,
    });
    selectWallet('import');
    acceptTerms();
    enterPassword();

    expect(
      screen.getByRole('heading', { name: 'Import your wallet' }),
    ).toBeVisible();
    expect(client.createMnemonic).not.toHaveBeenCalled();
    fireEvent.change(screen.getByTestId('mnemonic-field'), {
      target: { value: TEST_PHRASE },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Select address' }));
    expect(
      await screen.findByRole('combobox', { name: 'Wallet address' }),
    ).toHaveValue('0');
    expect(client.suggestAddresses).toHaveBeenCalledWith(TEST_PHRASE);
    fireEvent.click(screen.getByRole('button', { name: 'Continue' }));

    expect(
      screen.getByRole('heading', { name: 'Choose your connection' }),
    ).toBeVisible();
    const nodeUrl = 'https://rpc.example.invalid';
    fireEvent.change(screen.getByTestId('ethNode-field'), {
      target: { value: nodeUrl },
    });
    fireEvent.click(
      screen.getByRole('button', { name: 'Use custom connection' }),
    );
    fireEvent.click(
      screen.getByRole('button', { name: 'Use custom connection' }),
    );
    expect(onCompleted).toHaveBeenCalledTimes(1);
    expect(screen.getByTestId('ethNode-field')).toBeDisabled();
    await act(async () =>
      pending.reject(new Error('The local node could not save the wallet.')),
    );

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'The local node could not save the wallet.',
    );
    expect(
      screen.getByRole('heading', { name: 'Choose your connection' }),
    ).toBeVisible();
    expect(screen.getByTestId('ethNode-field')).toHaveValue(nodeUrl);
    fireEvent.click(
      screen.getByRole('button', { name: 'Use custom connection' }),
    );
    await waitFor(() => expect(onCompleted).toHaveBeenCalledTimes(2));
    expect(onCompleted).toHaveBeenLastCalledWith({
      password: TEST_PASSWORD,
      mnemonic: TEST_PHRASE,
      privateKey: '',
      derivationPath: '0',
      ethNode: nodeUrl,
    });
    expect(client.createMnemonic).not.toHaveBeenCalled();
  });

  it('preserves the new password if phrase generation fails and can retry without re-entry', async () => {
    const generated = vi
      .fn()
      .mockRejectedValueOnce(new Error('Random generation failed'))
      .mockResolvedValue(TEST_PHRASE);
    const { client, onCompleted } = renderOnboarding({
      createMnemonic: generated,
    });
    selectWallet('create');
    acceptTerms();
    enterPassword();

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Could not generate a recovery phrase',
    );
    expect(
      screen.getByRole('heading', { name: 'Create your app password' }),
    ).toBeVisible();
    expect(screen.getByTestId('pass-field')).toHaveValue(TEST_PASSWORD);
    expect(screen.getByTestId('pass-again-field')).toHaveValue(TEST_PASSWORD);
    expect(screen.getByTestId('pass-field')).toBeEnabled();
    expect(onCompleted).not.toHaveBeenCalled();
    fireEvent.submit(screen.getByTestId('pass-form'));

    expect(await screen.findByTestId('mnemonic-label')).toHaveTextContent(
      TEST_PHRASE,
    );
    expect(client.createMnemonic).toHaveBeenCalledTimes(2);
    expect(onCompleted).not.toHaveBeenCalled();
  });

  it('retains an imported phrase when address lookup fails and offers a working retry', async () => {
    const pendingAddresses = deferred<string[]>();
    const { client, onCompleted } = renderOnboarding();
    client.suggestAddresses.mockImplementationOnce(
      () => pendingAddresses.promise,
    );
    selectWallet('import');
    acceptTerms();
    enterPassword();
    fireEvent.change(screen.getByTestId('mnemonic-field'), {
      target: { value: TEST_PHRASE },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Select address' }));

    // The displayed accounts must correspond to the phrase that was queried.
    expect(screen.getByTestId('mnemonic-field')).toBeDisabled();
    expect(
      screen.getByRole('combobox', { name: 'Import method' }),
    ).toBeDisabled();
    expect(
      screen.getByRole('button', { name: 'Loading addresses…' }),
    ).toBeDisabled();
    await act(async () =>
      pendingAddresses.reject(new Error('Address lookup unavailable')),
    );

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Address lookup unavailable',
    );
    expect(screen.getByTestId('mnemonic-field')).toHaveValue(TEST_PHRASE);
    expect(screen.getByTestId('mnemonic-field')).toBeEnabled();
    expect(
      screen.queryByRole('combobox', { name: 'Wallet address' }),
    ).not.toBeInTheDocument();
    expect(onCompleted).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: 'Select address' }));

    expect(
      await screen.findByRole('combobox', { name: 'Wallet address' }),
    ).toHaveValue('0');
    expect(client.suggestAddresses).toHaveBeenCalledTimes(2);
    expect(client.createMnemonic).not.toHaveBeenCalled();
  });

  it('uses the default connection even if a custom URL was typed previously', async () => {
    const { onCompleted } = renderOnboarding();
    selectWallet('import');
    acceptTerms();
    enterPassword();
    fireEvent.change(screen.getByTestId('mnemonic-field'), {
      target: { value: TEST_PHRASE },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Select address' }));
    await screen.findByRole('combobox', { name: 'Wallet address' });
    fireEvent.click(screen.getByRole('button', { name: 'Continue' }));
    fireEvent.change(screen.getByTestId('ethNode-field'), {
      target: { value: 'https://custom.example.invalid' },
    });
    fireEvent.click(
      screen.getByRole('button', { name: 'Use default connection' }),
    );

    await waitFor(() => expect(onCompleted).toHaveBeenCalledTimes(1));
    expect(onCompleted).toHaveBeenCalledWith({
      password: TEST_PASSWORD,
      mnemonic: TEST_PHRASE,
      privateKey: '',
      derivationPath: '0',
      ethNode: '',
    });
  });
});
