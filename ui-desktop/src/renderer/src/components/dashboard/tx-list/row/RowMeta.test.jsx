import { describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ApproveMeta, defaultMeta } from './RowMeta';

const WALLET = '0x1111111111111111111111111111111111111111';
const OTHER = '0x2222222222222222222222222222222222222222';

const transferTx = (tokenSymbol, value = '1500000', tokenDecimals = 6) => ({
  transfers: [{ from: OTHER, to: WALLET, value, tokenSymbol, tokenDecimals }],
});

describe('RowMeta amount formatting', () => {
  it('renders a transfer whose token symbol is not a 3-letter ISO code', () => {
    // Intl.NumberFormat in currency style throws "Invalid currency code"
    // for anything but 3 letters, which took the whole wallet tab down
    // whenever a USDC transfer appeared in the history.
    render(defaultMeta({ tx: transferTx('USDC'), walletAddress: WALLET }));

    expect(screen.getByText('+USDC 1.5000')).toBeInTheDocument();
  });

  it('keeps the symbol next to the amount for 3-letter tokens', () => {
    render(
      defaultMeta({
        tx: transferTx('MOR', '2500000000000000000', 18),
        walletAddress: WALLET,
      }),
    );

    expect(screen.getByText('+MOR 2.5000')).toBeInTheDocument();
  });

  it('formats approvals for tokens with long symbols', () => {
    const contractAddress = '0x3333333333333333333333333333333333333333';
    const tx = {
      contract: {
        contractAddress,
        decodedInput: [
          { key: 'spender', value: OTHER },
          { key: 'amount', value: '42000000' },
        ],
      },
    };
    const tokens = { [contractAddress]: { decimals: 6, symbol: 'USDC' } };

    render(<ApproveMeta tx={tx} tokens={tokens} />);

    expect(screen.getByText('USDC 42.000')).toBeInTheDocument();
  });
});
