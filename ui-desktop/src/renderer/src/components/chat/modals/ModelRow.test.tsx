import { render, screen } from '@testing-library/react';
import type { FC, PropsWithChildren } from 'react';
import { ThemeProvider as StyledThemeProvider } from 'styled-components';
import { describe, expect, it, vi } from 'vitest';
import theme from '../../../ui/theme';
import ModelRow from './ModelRow';

const ThemeProvider = StyledThemeProvider as unknown as FC<
  PropsWithChildren<{ theme: typeof theme }>
>;

describe('ModelRow provider discovery state', () => {
  it('keeps an unchecked marketplace model selectable without calling it unavailable', () => {
    const { rerender } = render(
      <ThemeProvider theme={theme}>
        <ModelRow
          model={{ Id: 'model-1', Name: 'Model one', Tags: ['llm'] }}
          onChangeModel={vi.fn()}
          symbol="MOR"
        />
      </ThemeProvider>,
    );

    const row = screen.getByRole('button', { name: /Model one/ });
    expect((row as HTMLButtonElement).disabled).toBe(false);
    expect(screen.getByText('Check price')).toBeTruthy();
    expect(screen.queryByText('No providers')).toBeNull();

    rerender(
      <ThemeProvider theme={theme}>
        <ModelRow
          model={{
            Id: 'model-1',
            Name: 'Model one',
            Tags: ['llm'],
            bids: [],
          }}
          onChangeModel={vi.fn()}
          symbol="MOR"
        />
      </ThemeProvider>,
    );

    const unavailableRow = screen.getByRole('button', { name: /Model one/ });
    expect((unavailableRow as HTMLButtonElement).disabled).toBe(true);
    expect(screen.getByText('No providers')).toBeTruthy();
  });

  it.each([
    ['null', null],
    ['an object', { unexpected: 'shape' }],
    ['mixed invalid entries', ['llm', null, { nested: true }, ['tee']]],
  ])('renders safely when Tags is %s', (_description, Tags) => {
    render(
      <ThemeProvider theme={theme}>
        <ModelRow
          model={{
            Id: 'malformed-tags-model',
            Name: 'Malformed tags model',
            Tags,
            bids: undefined,
          }}
          onChangeModel={vi.fn()}
          symbol="MOR"
        />
      </ThemeProvider>,
    );

    expect(
      screen.getByRole('button', { name: /Malformed tags model/ }),
    ).toBeTruthy();
    expect(screen.getByText('Check price')).toBeTruthy();
  });

  it('prints the swept price when the model carries no bids of its own', () => {
    const { rerender } = render(
      <ThemeProvider theme={theme}>
        <ModelRow
          model={{ Id: 'swept', Name: 'Swept model', Tags: ['llm'] }}
          priceEntry={{
            model_id: 'swept',
            min_price_per_second_wei: '1000000000000000',
            max_price_per_second_wei: '1000000000000000',
            bid_count: 1,
          }}
          onChangeModel={vi.fn()}
          symbol="MOR"
        />
      </ThemeProvider>,
    );

    // A single offer is quoted as one figure, with no provider count, because
    // "1 providers" qualifies nothing.
    expect(screen.getByText('MOR/s')).toBeTruthy();
    expect(screen.queryByText('Check price')).toBeNull();

    rerender(
      <ThemeProvider theme={theme}>
        <ModelRow
          model={{ Id: 'swept', Name: 'Swept model', Tags: ['llm'] }}
          priceEntry={{
            model_id: 'swept',
            min_price_per_second_wei: '1000000000000000',
            max_price_per_second_wei: '3000000000000000',
            bid_count: 3,
          }}
          onChangeModel={vi.fn()}
          symbol="MOR"
        />
      </ThemeProvider>,
    );

    // Several offers become a range, and the count says what the range is one
    // of rather than leaving the low end looking like the price.
    expect(screen.getByText(/MOR\/s · 3 providers/)).toBeTruthy();
  });

  it('separates a model nobody serves from one the sweep has not reached', () => {
    const { rerender } = render(
      <ThemeProvider theme={theme}>
        <ModelRow
          model={{ Id: 'dead', Name: 'Dead model', Tags: ['llm'] }}
          priceEntry={{
            model_id: 'dead',
            min_price_per_second_wei: '',
            max_price_per_second_wei: '',
            bid_count: 0,
          }}
          onChangeModel={vi.fn()}
          symbol="MOR"
        />
      </ThemeProvider>,
    );

    // The sweep answered, and the answer was that there is nobody to pay.
    expect(screen.getByText('No providers')).toBeTruthy();

    // The sweep has a cached, wallet-filtered view of the market. Letting it
    // disable the row would grey out a model the user can actually open.
    expect(
      (screen.getByRole('button', { name: /Dead model/ }) as HTMLButtonElement)
        .disabled,
    ).toBe(false);

    rerender(
      <ThemeProvider theme={theme}>
        <ModelRow
          model={{ Id: 'dead', Name: 'Dead model', Tags: ['llm'] }}
          onChangeModel={vi.fn()}
          symbol="MOR"
        />
      </ThemeProvider>,
    );

    expect(screen.getByText('Check price')).toBeTruthy();
    expect(screen.queryByText('No providers')).toBeNull();
  });

  it('normalizes comma-separated tag strings before rendering badges', () => {
    render(
      <ThemeProvider theme={theme}>
        <ModelRow
          model={{
            Id: 'string-tags-model',
            Name: 'String tags model',
            Tags: 'llm, tee',
          }}
          onChangeModel={vi.fn()}
          symbol="MOR"
        />
      </ThemeProvider>,
    );

    expect(screen.getByText('LLM')).toBeTruthy();
    expect(screen.getByText('Secure')).toBeTruthy();
  });
});
