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
    expect(screen.queryByText('Unavailable')).toBeNull();

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
    expect(screen.getByText('Unavailable')).toBeTruthy();
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
