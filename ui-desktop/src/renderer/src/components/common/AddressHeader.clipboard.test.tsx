import React from 'react';
import '@testing-library/jest-dom/vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ThemeProvider as StyledThemeProvider } from 'styled-components';
import { describe, expect, it, vi } from 'vitest';
import { AddressHeader } from './AddressHeader';
import { ReceiveForm } from '../dashboard/tx-modal/ReceiveForm';
import { SendForm } from '../dashboard/tx-modal/SendForm';
import { ToastsContext } from '../toasts';
import theme from '../../ui/theme';

vi.mock('qrcode.react', () => ({
  default: () => <div data-testid="address-qr" />,
}));

const address = '0x51d01234567890abcdef1234567890abcdefc25a1';
const ThemeProvider = StyledThemeProvider as unknown as React.FC<
  React.PropsWithChildren<{ theme: typeof theme }>
>;

function renderWithToasts(element: React.ReactElement, toast = vi.fn()) {
  return {
    toast,
    ...render(
      <ThemeProvider theme={theme}>
        <ToastsContext.Provider value={{ toast }}>
          {element}
        </ToastsContext.Provider>
      </ThemeProvider>,
    ),
  };
}

describe('wallet address clipboard controls', () => {
  it('copies the full address by keyboard and confirms only after the write completes', async () => {
    let complete!: () => void;
    const copy = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          complete = resolve;
        }),
    );
    const user = userEvent.setup();
    const { toast } = renderWithToasts(
      <AddressHeader address={address} copyToClipboard={copy} />,
    );

    await user.tab();
    expect(
      screen.getByRole('button', { name: 'Copy wallet address' }),
    ).toHaveFocus();
    await user.keyboard('{Enter}');
    expect(copy).toHaveBeenCalledWith(address);
    expect(toast).not.toHaveBeenCalled();

    complete();
    await waitFor(() =>
      expect(toast).toHaveBeenCalledWith(
        'success',
        'Address copied to clipboard',
        { autoClose: 1500 },
      ),
    );
  });

  it('offers a retry message when copying fails and never reports success', async () => {
    const copy = vi.fn().mockRejectedValue(new Error('Clipboard unavailable'));
    const user = userEvent.setup();
    const { toast } = renderWithToasts(
      <AddressHeader address={address} copyToClipboard={copy} />,
    );

    await user.click(
      screen.getByRole('button', { name: 'Copy wallet address' }),
    );
    expect(toast).toHaveBeenCalledTimes(1);
    expect(toast).toHaveBeenCalledWith(
      'error',
      'Could not copy your address. Please try again.',
    );
  });

  it('does not replace the clipboard while the wallet address is unavailable', () => {
    const copy = vi.fn();
    renderWithToasts(
      <AddressHeader address={undefined} copyToClipboard={copy} />,
    );
    expect(
      screen.getByRole('button', { name: 'Copy wallet address' }),
    ).toBeDisabled();
    expect(screen.getByText('Wallet not connected')).toBeVisible();
    expect(copy).not.toHaveBeenCalled();
  });

  it('shows the complete receive address and copies that exact value', async () => {
    const copy = vi.fn().mockResolvedValue(undefined);
    const user = userEvent.setup();
    renderWithToasts(
      <ReceiveForm
        activeTab
        address={address}
        onRequestClose={vi.fn()}
        copyToClipboard={copy}
        explorerUrl="https://basescan.org"
        eth={{ symbol: 'ETH', value: 0 }}
        mor={{ symbol: 'MOR', value: 0 }}
      />,
    );

    expect(screen.getByText(address)).toBeVisible();
    await user.click(
      screen.getByRole('button', { name: 'Copy wallet address' }),
    );
    expect(copy).toHaveBeenCalledWith(address);
  });

  it('allows normal paste into the recipient field without sending a transaction', async () => {
    const onInputChange = vi.fn();
    const onSubmit = vi.fn();
    const user = userEvent.setup();
    renderWithToasts(
      <SendForm
        activeTab
        selectedCurrency={{ label: 'MOR', value: 'mor' }}
        currencyOptions={[]}
        toAddress=""
        coinAmount=""
        symbolEth="ETH"
        onInputChange={onInputChange}
        onSubmit={onSubmit}
      />,
    );

    await user.click(
      screen.getByRole('textbox', { name: 'Recipient wallet address' }),
    );
    await user.paste(address);
    expect(onInputChange).toHaveBeenCalledWith({
      id: 'toAddress',
      value: address,
    });
    expect(onSubmit).not.toHaveBeenCalled();
  });
});
